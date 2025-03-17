package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stevenlei/polymer-cli/pkg/api"
	"github.com/stevenlei/polymer-cli/pkg/config"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [jobID]",
	Short: "Check the status of a proof generation job",
	Long: `Check the status of a proof generation job.

Provide the job ID that was returned when you requested a proof.

Example:
  polymer-cli status 12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get job ID from arguments
		jobID := args[0]

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

		// Get proof status
		if cfg.Debug {
			fmt.Printf("Checking status for job ID: %s...\n", jobID)
		}

		status, err := client.GetProofStatus(jobID)
		if err != nil {
			return fmt.Errorf("failed to get proof status: %w", err)
		}

		// In non-debug mode, just output the status
		if !cfg.Debug {
			fmt.Println(status.Status)

			// If the proof is ready, also print it
			if status.Status == "completed" && len(status.Proof) > 0 {
				// Always use raw output in non-debug mode
				// Try to unmarshal if it's a JSON string
				var s string
				if err := json.Unmarshal(status.Proof, &s); err == nil {
					// It's a JSON string, so use the unquoted value
					fmt.Print(s)
				} else {
					// It's not a JSON string or there was an error
					rawStr := string(status.Proof)
					if len(rawStr) >= 2 && rawStr[0] == '"' && rawStr[len(rawStr)-1] == '"' {
						rawStr = rawStr[1 : len(rawStr)-1]
					}
					fmt.Print(rawStr)
				}
			}

			return nil
		}

		// Print status (debug mode)
		fmt.Printf("Status: %s\n", status.Status)

		// If there's an error in the status response
		if status.Error != "" {
			fmt.Printf("Error: %s\n", status.Error)
		}

		// If the proof is ready, print it
		if status.Status == "completed" && len(status.Proof) > 0 {
			fmt.Println("Proof is ready!")

			if returnRaw {
				// Try to unmarshal if it's a JSON string
				var s string
				if err := json.Unmarshal(status.Proof, &s); err == nil {
					// It's a JSON string, so use the unquoted value
					fmt.Print(s)
				} else {
					// It's not a JSON string or there was an error
					rawStr := string(status.Proof)
					if len(rawStr) >= 2 && rawStr[0] == '"' && rawStr[len(rawStr)-1] == '"' {
						rawStr = rawStr[1 : len(rawStr)-1]
					}
					fmt.Print(rawStr)
				}
			} else {
				// Format as pretty JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, status.Proof, "", "  "); err != nil {
					return fmt.Errorf("failed to format proof as JSON: %w", err)
				}
				fmt.Println(prettyJSON.String())
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Optional flags
	statusCmd.Flags().BoolVar(&returnRaw, "raw", false, "Return raw JSON output")
}
