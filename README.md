# Polymer CLI

A command-line interface for interacting with the Polymer Prove API, built with Go.

## Overview

Polymer CLI allows you to interact with the Polymer Prove API, learn more about Polymer at https://polymerlabs.org.

## Features

- Request proofs for blockchain transactions using either:
  - Polymer Prove API required parameters (chain ID, block number, transaction index, log index)
  - Transaction hash with optional event signature or log index
- Check the status of proof generation jobs
- Wait for proofs to be generated with configurable polling parameters
- Pretty-print or raw JSON output options
- Configuration via file, environment variables, or command-line flags

## Installation

### Build from Source

```bash
git clone https://github.com/polymer/polymer-cli.git
cd polymer-cli
go build
```

The executable `polymer-cli` will be created in the current directory.

## Configuration

Polymer CLI can be configured in multiple ways (in order of precedence):

1. Command-line flags
2. Environment variables (prefixed with `POLYMER_`)
3. Configuration file

### Configuration File

By default, Polymer CLI looks for a configuration file at `$HOME/.polymer-cli.yaml`. You can specify a different file using the `--config` flag.

Example configuration file:

```yaml
api-key: "your-polymer-api-key"
api-url: "https://proof.testnet.polymer.zone"
debug: false
max-attempts: 20
interval: 3000
```

### Environment Variables

You can also use environment variables to configure Polymer CLI:

```bash
export POLYMER_API_KEY="your-polymer-api-key"
export POLYMER_API_URL="https://proof.testnet.polymer.zone"
export POLYMER_DEBUG=false
export POLYMER_MAX_ATTEMPTS=20
export POLYMER_INTERVAL=3000
```

## Usage

### Main Commands

- `request`: Request a new batch proof
  - Option 1: Without blockchain RPC:
    - `--chain-id`: Source chain ID
    - `--block-number`: Source block number
    - `--tx-index`: Transaction index in the block
    - `--log-index`: Log index in the transaction
  - Option 2: With blockchain RPC:
    - `--tx-hash`: Transaction hash to request proof for
    - `--rpc-url`: RPC URL for the blockchain (required when using --tx-hash)
    - `--event-signature`: Event signature to identify the log (e.g., 'Transfer(address,address,uint256)')
  - `--wait`: Wait for the proof to be generated
  - `--api-key`: Polymer API key
  - `--debug`: Enable debug logging
- `status <jobID>`: Check the status of a proof generation job
  - `--api-key`: Polymer API key
  - `--debug`: Enable debug logging
- `wait <jobID>`: Wait for a proof to be generated
  - `--api-key`: Polymer API key
  - `--debug`: Enable debug logging
  - `--max-attempts`: Maximum number of polling attempts
  - `--interval`: Polling interval in milliseconds
- `version`: Print the version number

### Request a Proof by Chain ID, Block Number, Transaction Index

```bash
polymer-cli request --chain-id=11155420 --block-number=24639225 --tx-index=4 --log-index=1 --api-key=your-polymer-api-key
```

### Request a Proof by Transaction Hash

You can also request proofs by specifying a transaction hash, with either a log index or event signature, which simplifies the process by automatically retrieving all required details:

```bash
# With log index
polymer-cli request --tx-hash=0x5138b0d6ffe7bfe8f1d7dca24d396dab804fa664930ef96bb9e6ebbc86426fbb --rpc-url=https://sepolia.optimism.io --log-index=1 --api-key=your-polymer-api-key

# With event signature (e.g., for a Transfer event)
polymer-cli request --tx-hash=0x5138b0d6ffe7bfe8f1d7dca24d396dab804fa664930ef96bb9e6ebbc86426fbb --rpc-url=https://sepolia.optimism.io --event-signature="ValueSet(address,string,bytes,uint256,bytes32,uint256)" --api-key=your-polymer-api-key
```

When using a transaction hash, you must specify an RPC URL of the target blockchain to fetch transaction details.

### Wait for Proof Generation

Add the `--wait` flag to any request command to automatically wait for the proof to be generated:

```bash
polymer-cli request --tx-hash=0x5138b0d6ffe7bfe8f1d7dca24d396dab804fa664930ef96bb9e6ebbc86426fbb --rpc-url=https://sepolia.optimism.io --event-signature="ValueSet(address,string,bytes,uint256,bytes32,uint256)" --api-key=your-polymer-api-key --wait
```

### Check Proof Status

```bash
polymer-cli status <job-id>
```

### Wait for Proof

```bash
polymer-cli wait <job-id> --max-attempts=30 --interval=5000
```

### Display Version

```bash
polymer-cli version
```

## Global Flags

- `--api-key string`: Polymer API key
- `--api-url string`: Polymer API URL (default "https://proof.testnet.polymer.zone")
- `--config string`: Config file (default is $HOME/.polymer-cli.yaml)
- `--debug`: Enable debug logging

## Request Command Flags

- `--chain-id string`: Source chain ID
- `--block-number string`: Source block number
- `--tx-index string`: Transaction index in the block
- `--log-index string`: Log index in the transaction
- `--tx-hash string`: Transaction hash to request proof for
- `--rpc-url string`: RPC URL for the blockchain (required when using --tx-hash)
- `--event-signature string`: Event signature to identify the log (e.g., 'Transfer(address,address,uint256)')
- `--raw`: Return raw JSON output
- `--wait`: Wait for the proof to be generated

## License

MIT
