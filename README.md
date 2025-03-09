# Solana gRPC Exploration

A project exploring the implementation of gRPC with Solana blockchain.

## Overview

This project explores how to build a gRPC service layer on top of Solana blockchain, enabling efficient, type-safe, and bidirectional streaming communication.

### What is gRPC?

gRPC is an open-source RPC framework that uses HTTP/2 for transport and Protocol Buffers as the interface description language.

### What is Solana?

Solana is a blockchain platform designed for decentralized applications with high transaction throughput.

## Features

- Bidirectional streaming for real-time blockchain data
- Type-safe interfaces using Protocol Buffers
- Direct interaction with Solana blockchain
- Performance benchmarking comparing gRPC vs JSON-RPC
- Token price data retrieval and visualization
- Real-time token price streaming

## gRPC vs JSON-RPC Comparison

This project explores the differences between gRPC and JSON-RPC for Solana blockchain interactions. Here's how these technologies compare:

### Protocol & Transport

| Feature | gRPC | JSON-RPC |
|---------|------|----------|
| **Serialization** | Protocol Buffers (binary) | JSON (text) |
| **Transport** | HTTP/2 | Typically HTTP/1.1 |
| **Connection** | Persistent, multiplexed | Usually new connection per request |

### Data Format

| Feature | gRPC | JSON-RPC |
|---------|------|----------|
| **Format** | Binary (smaller payloads) | Text (human-readable) |
| **Schema** | Formal (.proto files) | Often informal or separate (OpenAPI) |
| **Type Safety** | Strong, compile-time checking | Loose, runtime checking |
| **Code Generation** | Built-in | Optional, not built into protocol |

### Features

| Feature | gRPC | JSON-RPC |
|---------|------|----------|
| **Streaming** | Comprehensive (unary, server, client, bidirectional) | Limited or none (depends on transport) |
| **Timeouts** | Built-in deadline propagation | Not standardized |
| **Middleware** | Interceptors for auth, logging, etc. | Varies by implementation |
| **Error Handling** | Structured with status codes | Simple error objects |

### Performance

| Aspect | gRPC | JSON-RPC |
|--------|------|----------|
| **Speed** | Faster (binary serialization, HTTP/2) | Slower (text parsing overhead) |
| **Latency** | Lower (header compression, multiplexing) | Higher (especially with many small requests) |
| **Resource Usage** | More efficient CPU and bandwidth | Higher bandwidth consumption |

### Solana Context

For Solana blockchain operations:

- **JSON-RPC**: Currently the standard API for most Solana applications
- **gRPC**: Potential benefits include:
  - Faster transaction submission and account data retrieval
  - Real-time streaming of blockchain updates
  - Type safety for complex blockchain data structures
  - Lower bandwidth usage for high-frequency operations

## Project Structure

```
solana-grpc-exploration/
├── proto/                  # Protocol Buffer definitions
├── server/                 # gRPC server implementation
│   ├── solana/             # Solana blockchain integration
│   └── services/           # gRPC service implementations
├── client/                 # Sample client implementations
│   ├── go/                 # Go client example
│   ├── js/                 # JavaScript client example (coming soon)
│   └── python/             # Python client example (coming soon)
├── tests/                  # Integration and unit tests (coming soon)
└── docs/                   # Documentation (coming soon)
```

## Getting Started

### Prerequisites

- Go 1.16+
- Protocol Buffers compiler
- Solana CLI tools (optional)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/i-tozer/grpc-solana-trump.git
   cd grpc-solana-trump
   ```

2. Install Protocol Buffers compiler:
   ```bash
   # macOS
   brew install protobuf

   # Ubuntu
   sudo apt-get install protobuf-compiler

   # Windows (using Chocolatey)
   choco install protoc
   ```

3. Install Go dependencies:
   ```bash
   go mod tidy
   ```

4. Generate Go code from Protocol Buffers:
   ```bash
   make proto
   ```

5. Build the server and client:
   ```bash
   make all
   ```

### Running the Server

Start the gRPC server:

```bash
make run-server
```

You can specify a different Solana RPC endpoint and the TRUMP token mint address:

```bash
./bin/server --port=50051 --rpc-endpoint=https://api.mainnet-beta.solana.com --trump-token-mint=3vHSsSV1iRoLbspdNtRcHN3nwmYqGJ5hRQQFSvfUHYJz
```

### Running the Client

The client provides several commands to interact with the gRPC server:

#### Benchmark

Run a performance benchmark comparing gRPC vs JSON-RPC:

```bash
make run-benchmark
```

Or with custom parameters:

```bash
./bin/client --command=benchmark --pubkey=SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt4 --iterations=10
```

#### Get Account Info

Retrieve information about a Solana account:

```bash
make run-account
```

Or with custom parameters:

```bash
./bin/client --command=account --pubkey=SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt4
```

#### Get Transaction Info

Retrieve information about a Solana transaction:

```bash
./bin/client --command=transaction --signature=YOUR_TRANSACTION_SIGNATURE
```

#### Get Block Info

Retrieve information about a Solana block:

```bash
make run-block
```

Or with custom parameters:

```bash
./bin/client --command=block --slot=150000000
```

#### Stream Account Updates

Stream real-time updates for a Solana account:

```bash
make run-stream-accounts
```

Or with custom parameters:

```bash
./bin/client --command=stream-accounts --pubkey=SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt4
```

#### Stream Transaction Updates

Stream real-time transaction updates:

```bash
make run-stream-transactions
```

#### Stream Block Updates

Stream real-time block updates:

```bash
make run-stream-blocks
```

#### Get Token Price

Retrieve the current price of a token (defaults to TRUMP token):

```bash
make run-token-price
```

Or with custom parameters:

```bash
./bin/client --command=token-price get --token-mint=3vHSsSV1iRoLbspdNtRcHN3nwmYqGJ5hRQQFSvfUHYJz --quote-mint=EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v
```

#### Get Token Price History

Retrieve historical price data for a token and generate a price chart:

```bash
make run-token-price-history
```

Or with custom parameters:

```bash
./bin/client --command=token-price history --token-mint=3vHSsSV1iRoLbspdNtRcHN3nwmYqGJ5hRQQFSvfUHYJz --interval=1h --limit=100 --output=trump_price_chart.png
```

#### Stream Token Price Updates

Stream real-time token price updates and generate a live price chart:

```bash
make run-token-price-stream
```

Or with custom parameters:

```bash
./bin/client --command=token-price stream --token-mint=3vHSsSV1iRoLbspdNtRcHN3nwmYqGJ5hRQQFSvfUHYJz --update-interval=1000 --duration=60
```

## Performance Benchmarking

This project includes a benchmarking tool to compare the performance of gRPC vs JSON-RPC for Solana operations. The benchmark measures:

- Account information retrieval
- Transaction information retrieval
- Block information retrieval

The benchmark results include:
- Average response time
- Minimum response time
- Maximum response time
- Success/failure counts
- Overall speedup factor

## Token Price Service

The project includes a token price service that provides real-time and historical price data for tokens on Solana, with a focus on the TRUMP token. The service offers:

### Features

- Current token price retrieval
- Historical price data with customizable time intervals
- Real-time price streaming with configurable update intervals
- Price visualization with charts
- Support for different DEXes (currently simulated data, can be extended to use Jupiter, Birdeye, or other price sources)

### Data Visualization

The token price service includes chart generation capabilities:

- Historical price charts with OHLC (Open, High, Low, Close) data
- Volume indicators
- Real-time price charts with live updates
- Customizable chart output formats

### Implementation Details

The token price service is implemented using:

- Protocol Buffers for type-safe data definitions
- gRPC for efficient client-server communication
- Streaming RPCs for real-time updates
- Go-chart library for chart generation

## Technical Details

### Protocol Buffers

The project defines several Protocol Buffer message types for Solana entities:

#### Benchmark Service

- `AccountInfoRequest/Response`: For retrieving account information
- `TransactionRequest/Response`: For retrieving transaction information
- `BlockRequest/Response`: For retrieving block information
- `AccountStreamRequest/AccountUpdate`: For streaming account updates
- `TransactionStreamRequest/TransactionUpdate`: For streaming transaction updates
- `BlockStreamRequest/BlockUpdate`: For streaming block updates
- `BenchmarkRequest/Results`: For running and reporting benchmark results

#### Token Price Service

- `TokenPriceRequest/Response`: For retrieving current token price information
- `TokenPriceHistoryRequest/Response`: For retrieving historical token price data
- `PriceDataPoint`: Represents a single price data point with OHLC and volume data
- `TokenPriceStreamRequest/TokenPriceUpdate`: For streaming real-time token price updates

### gRPC Services

The project implements the following gRPC services:

#### BenchmarkService

Provides methods for benchmarking Solana RPC vs gRPC performance:

- `GetAccountInfo`: Retrieves account information
- `GetTransaction`: Retrieves transaction information
- `GetBlock`: Retrieves block information
- `StreamAccountUpdates`: Streams account updates
- `StreamTransactions`: Streams transaction updates
- `StreamBlocks`: Streams block updates
- `RunBenchmark`: Runs a comprehensive benchmark suite

#### TokenPriceService

Provides methods for fetching token price data from DEXes:

- `GetTokenPrice`: Gets the current price of a token
- `GetTokenPriceHistory`: Gets historical price data for a token
- `StreamTokenPrice`: Streams real-time price updates for a token

### Chart Generation

The project uses the [go-chart](https://github.com/wcharczuk/go-chart) library for generating price charts:

- Historical price charts with OHLC data
- Volume indicators
- Real-time price charts with live updates
- Customizable chart styling and output formats

## Contributing

Contributions are welcome. Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.