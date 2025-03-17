package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/stevenlei/polymer-cli/pkg/api"
	"github.com/stevenlei/polymer-cli/pkg/config"
	"github.com/stevenlei/polymer-cli/pkg/rpc"
)

var chainID string
var blockNumber string
var txIndex string
var logIndex string
var txHash string
var rpcURL string
var eventSignature string
var waitForProof bool
var returnRaw bool

// requestCmd represents the request command
var requestCmd = &cobra.Command{
	Use:   "request [flags]",
	Short: "Request a new batch proof",
	Long: `Request a new batch proof.

You can specify the transaction either by providing the chain ID, block number, transaction index, and log index,
or by providing a transaction hash (with an optional log index or event signature).

Examples using transaction parameters:
  polymer-cli request --chain-id=1 --block-number=17000000 --tx-index=5 --log-index=2

Examples using transaction hash:
  polymer-cli request --tx-hash=0x123... --log-index=1
  polymer-cli request --tx-hash=0x123... --event-signature="Transfer(address,address,uint256)"

Use --wait to wait for the proof to be generated.

The RPC URL is required when using --tx-hash, but not when providing direct transaction parameters.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Create API client
		client := api.NewClient(cfg.APIKey, cfg.APIURL, cfg.Debug)

		// Check if the user provided a transaction hash
		if txHash != "" {
			// Ensure RPC URL is provided
			if rpcURL == "" {
				return fmt.Errorf("RPC URL is required when using transaction hash")
			}

			return processTransactionByHash(client, txHash, rpcURL, cfg, waitForProof, returnRaw)
		}

		// Otherwise, proceed with chain ID, block number, etc.
		// Check if required flags are provided
		if chainID == "" || blockNumber == "" || txIndex == "" || logIndex == "" {
			return fmt.Errorf("chain-id, block-number, tx-index, and log-index are required")
		}

		// Parse chain ID
		chainIDUint, err := strconv.ParseUint(chainID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chain ID: %w", err)
		}

		// Parse block number
		blockNumberUint, err := strconv.ParseUint(blockNumber, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid block number: %w", err)
		}

		// Parse transaction index
		txIndexUint, err := strconv.ParseUint(txIndex, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid transaction index: %w", err)
		}

		// Parse log index
		logIndexUint, err := strconv.ParseUint(logIndex, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid log index: %w", err)
		}

		// Request proof
		fmt.Println("Requesting proof...")
		jobID, err := client.RequestProof(
			chainIDUint,
			blockNumberUint,
			uint(txIndexUint),
			uint(logIndexUint),
		)
		if err != nil {
			return fmt.Errorf("failed to request proof: %w", err)
		}

		if cfg.Debug {
			fmt.Println("Proof request submitted successfully")
			fmt.Printf("Job ID: %s\n", jobID)
		} else if !waitForProof {
			// Only print the job ID in non-debug mode if not waiting for proof
			fmt.Println(jobID)
		}

		if !waitForProof {
			return nil
		}

		return waitAndDisplayProof(client, jobID, cfg, returnRaw)
	},
}

// processTransactionByHash handles proof requests using a transaction hash
func processTransactionByHash(client *api.Client, txHash, rpcURL string, cfg config.Config, waitForProof, returnRaw bool) error {
	// Create RPC client
	if cfg.Debug {
		fmt.Printf("Connecting to RPC endpoint: %s\n", rpcURL)
	}
	rpcClient := rpc.NewRPCClient(rpcURL, cfg.Debug)

	// Fetch transaction details
	if cfg.Debug {
		fmt.Printf("Fetching transaction: %s\n", txHash)
	}
	tx, err := rpcClient.GetTransaction(txHash)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Fetch transaction receipt
	if cfg.Debug {
		fmt.Println("Fetching transaction receipt...")
	}
	receipt, err := rpcClient.GetTransactionReceipt(txHash)
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Extract block number
	blockNum, err := rpc.HexToUint64(receipt.BlockNumber)
	if err != nil {
		return fmt.Errorf("invalid block number in receipt: %w", err)
	}

	// Extract transaction index
	txIdx, err := rpc.HexToUint64(receipt.TransactionIndex)
	if err != nil {
		return fmt.Errorf("invalid transaction index in receipt: %w", err)
	}

	// Extract chain ID
	var chainIDUint uint64
	if tx.ChainID != "" {
		// Try to extract from transaction
		chainIDUint, err = rpc.HexToUint64(tx.ChainID)
		if err != nil {
			return fmt.Errorf("invalid chain ID in transaction: %w", err)
		}
	} else {
		// If not found in transaction, prompt user to provide it
		return fmt.Errorf("chain ID not found in transaction, please provide it with --chain-id flag")
	}

	// Determine which log to use
	logIdx := uint(0)
	logFound := false

	if len(receipt.Logs) == 0 {
		return fmt.Errorf("no logs found in transaction receipt")
	}

	// Case 1: User specified log index
	if logIndex != "" {
		logIdxParsed, err := strconv.ParseUint(logIndex, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid log index: %w", err)
		}

		if int(logIdxParsed) >= len(receipt.Logs) {
			return fmt.Errorf("log index %d is out of range, transaction has %d logs", logIdxParsed, len(receipt.Logs))
		}

		logIdx = uint(logIdxParsed)
		logFound = true
		if cfg.Debug {
			fmt.Printf("Using specified log index: %d\n", logIdx)
		}
	}

	// Case 2: User specified event signature
	if eventSignature != "" && !logFound {
		if cfg.Debug {
			fmt.Printf("Searching for log with event signature: %s\n", eventSignature)
		}

		// Normalize the event signature format
		normalizedSig := strings.TrimSpace(eventSignature)

		// Find matching log
		for i, log := range receipt.Logs {
			// Check if the topic matches the event signature
			if len(log.Topics) > 0 {
				// The first topic is the event signature hash
				// Compare to expected signature (for debugging purposes)
				if cfg.Debug {
					fmt.Printf("  Log %d Topic[0]: %s\n", i, log.Topics[0])
				}

				// For more accurate matching, we should use the Keccak256 hash of the event signature
				// But for now, we'll use a simpler approach that requires the API to add this feature
				eventHash, err := rpcClient.GetEventSignatureHash(normalizedSig)
				if err != nil {
					return fmt.Errorf("failed to get event signature hash: %w", err)
				}

				if strings.EqualFold(log.Topics[0], eventHash) {
					logIdx = uint(i)
					logFound = true
					if cfg.Debug {
						fmt.Printf("Found matching log at index %d\n", i)
					}
					break
				}
			}
		}

		if !logFound {
			return fmt.Errorf("no log found with event signature: %s", eventSignature)
		}
	}

	// Case 3: No log index or event signature provided, use the first log
	if !logFound {
		if cfg.Debug {
			fmt.Println("No log index or event signature provided, using first log")
		}
		logIdx = 0
	}

	// Display the transaction details
	if cfg.Debug {
		fmt.Printf("Transaction details:\n")
		fmt.Printf("  Chain ID: %d\n", chainIDUint)
		fmt.Printf("  Block Number: %d\n", blockNum)
		fmt.Printf("  Transaction Index: %d\n", txIdx)
		fmt.Printf("  Log Index: %d\n", logIdx)
	}

	// Request proof
	if cfg.Debug {
		fmt.Println("Requesting proof...")
	}
	jobID, err := client.RequestProof(
		chainIDUint,
		blockNum,
		uint(txIdx),
		logIdx,
	)
	if err != nil {
		return fmt.Errorf("failed to request proof: %w", err)
	}

	if cfg.Debug {
		fmt.Println("Proof request submitted successfully")
		fmt.Printf("Job ID: %s\n", jobID)
	} else if !waitForProof {
		// Only print the job ID in non-debug mode if not waiting for proof
		fmt.Println(jobID)
	}

	if !waitForProof {
		return nil
	}

	return waitAndDisplayProof(client, jobID, cfg, returnRaw)
}

// waitAndDisplayProof waits for a proof to be generated and displays it
func waitAndDisplayProof(client *api.Client, jobID string, cfg config.Config, returnRaw bool) error {
	// Wait for proof to be generated
	if cfg.Debug {
		fmt.Printf("Waiting for proof to be generated (max %d attempts, %dms interval)...\n",
			cfg.MaxAttempts, cfg.Interval)
	}

	proofStatus, err := client.WaitForProof(jobID, cfg.MaxAttempts, time.Duration(cfg.Interval)*time.Millisecond)
	if err != nil {
		return fmt.Errorf("failed while waiting for proof: %w", err)
	}

	if cfg.Debug {
		fmt.Println("Proof generated successfully!")
	}

	// Output proof
	if !cfg.Debug || returnRaw {
		// In non-debug mode, always use raw output
		// In debug mode, use raw output if returnRaw is true
		// Try to unmarshal if it's a JSON string
		var s string
		if err := json.Unmarshal(proofStatus.Proof, &s); err == nil {
			// It's a JSON string, so use the unquoted value
			fmt.Print(s)
		} else {
			// It's not a JSON string or there was an error
			rawStr := string(proofStatus.Proof)
			if len(rawStr) >= 2 && rawStr[0] == '"' && rawStr[len(rawStr)-1] == '"' {
				rawStr = rawStr[1 : len(rawStr)-1]
			}
			fmt.Print(rawStr)
		}
	} else {
		// Format as pretty JSON (only in debug mode and returnRaw is false)
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, proofStatus.Proof, "", "  "); err != nil {
			return fmt.Errorf("failed to format proof as JSON: %w", err)
		}
		fmt.Println(prettyJSON.String())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(requestCmd)

	// Flags for direct proof requests
	requestCmd.Flags().StringVar(&chainID, "chain-id", "", "Source chain ID")
	requestCmd.Flags().StringVar(&blockNumber, "block-number", "", "Source block number")
	requestCmd.Flags().StringVar(&txIndex, "tx-index", "", "Transaction index in the block")
	requestCmd.Flags().StringVar(&logIndex, "log-index", "", "Log index in the transaction")

	// Flags for transaction hash based requests
	requestCmd.Flags().StringVar(&txHash, "tx-hash", "", "Transaction hash to request proof for")
	requestCmd.Flags().StringVar(&rpcURL, "rpc-url", "", "RPC URL for the blockchain")
	requestCmd.Flags().StringVar(&eventSignature, "event-signature", "", "Event signature to identify the log (e.g., 'Transfer(address,address,uint256)')")

	// Optional flags
	requestCmd.Flags().BoolVar(&waitForProof, "wait", false, "Wait for the proof to be generated")
	requestCmd.Flags().BoolVar(&returnRaw, "raw", false, "Return raw JSON output")
}
