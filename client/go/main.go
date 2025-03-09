package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/i-tozer/grpc-solana-trump/proto"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("server", "localhost:50051", "The server address in the format host:port")
	command    = flag.String("command", "benchmark", "Command to run: benchmark, account, transaction, block, stream-accounts, stream-transactions, stream-blocks, token-price")
	pubkey     = flag.String("pubkey", "", "Solana account public key")
	signature  = flag.String("signature", "", "Solana transaction signature")
	slot       = flag.Uint64("slot", 0, "Solana block slot")
	iterations = flag.Uint("iterations", 10, "Number of iterations for benchmark")
)

func main() {
	flag.Parse()

	// Check if we're running the token price client
	if *command == "token-price" {
		runTokenPriceClient()
		return
	}

	// Set up a connection to the server
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a client
	client := proto.NewBenchmarkServiceClient(conn)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute the requested command
	switch *command {
	case "benchmark":
		runBenchmark(ctx, client)
	case "account":
		getAccountInfo(ctx, client)
	case "transaction":
		getTransaction(ctx, client)
	case "block":
		getBlock(ctx, client)
	case "stream-accounts":
		streamAccounts(ctx, client)
	case "stream-transactions":
		streamTransactions(ctx, client)
	case "stream-blocks":
		streamBlocks(ctx, client)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		os.Exit(1)
	}
}

func runBenchmark(ctx context.Context, client proto.BenchmarkServiceClient) {
	if *pubkey == "" {
		log.Fatal("--pubkey is required for benchmark")
	}

	// Run benchmark
	fmt.Printf("Running benchmark with %d iterations...\n", *iterations)
	resp, err := client.RunBenchmark(ctx, &proto.BenchmarkRequest{
		Iterations:       uint32(*iterations),
		TestAccounts:     []string{*pubkey},
		RunGrpcTests:     true,
		RunJsonrpcTests:  true,
		SolanaRpcUrl:     "https://api.mainnet-beta.solana.com",
	})
	if err != nil {
		log.Fatalf("Error running benchmark: %v", err)
	}

	// Print results
	fmt.Printf("\nBenchmark Results:\n")
	fmt.Printf("Total Time: %d ms\n", resp.TotalTimeMs)
	fmt.Printf("Iterations: %d\n", resp.Iterations)
	fmt.Printf("Overall Speedup: %.2fx\n", resp.OverallSpeedup)

	// Print gRPC results
	fmt.Printf("\ngRPC Results:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Method", "Avg Time (ms)", "Min Time (ms)", "Max Time (ms)", "Success", "Failure"})
	for _, result := range resp.GrpcResults {
		table.Append([]string{
			result.Method,
			fmt.Sprintf("%d", result.AvgResponseTimeMs),
			fmt.Sprintf("%d", result.MinResponseTimeMs),
			fmt.Sprintf("%d", result.MaxResponseTimeMs),
			fmt.Sprintf("%d", result.SuccessCount),
			fmt.Sprintf("%d", result.FailureCount),
		})
	}
	table.Render()

	// Print JSON-RPC results
	fmt.Printf("\nJSON-RPC Results:\n")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Method", "Avg Time (ms)", "Min Time (ms)", "Max Time (ms)", "Success", "Failure"})
	for _, result := range resp.JsonrpcResults {
		table.Append([]string{
			result.Method,
			fmt.Sprintf("%d", result.AvgResponseTimeMs),
			fmt.Sprintf("%d", result.MinResponseTimeMs),
			fmt.Sprintf("%d", result.MaxResponseTimeMs),
			fmt.Sprintf("%d", result.SuccessCount),
			fmt.Sprintf("%d", result.FailureCount),
		})
	}
	table.Render()
}

func getAccountInfo(ctx context.Context, client proto.BenchmarkServiceClient) {
	if *pubkey == "" {
		log.Fatal("--pubkey is required")
	}

	// Get account info
	fmt.Printf("Getting account info for %s...\n", *pubkey)
	resp, err := client.GetAccountInfo(ctx, &proto.AccountInfoRequest{
		Pubkey:         *pubkey,
		Commitment:     "finalized",
		EncodingBinary: true,
	})
	if err != nil {
		log.Fatalf("Error getting account info: %v", err)
	}

	// Print results
	fmt.Printf("\nAccount Info:\n")
	fmt.Printf("Pubkey: %s\n", resp.Pubkey)
	fmt.Printf("Owner: %s\n", resp.Owner)
	fmt.Printf("Lamports: %d\n", resp.Lamports)
	fmt.Printf("Executable: %t\n", resp.Executable)
	fmt.Printf("Rent Epoch: %d\n", resp.RentEpoch)
	fmt.Printf("Data Length: %d bytes\n", len(resp.Data))
	fmt.Printf("Response Time: %d ms\n", resp.ResponseTimeMs)
}

func getTransaction(ctx context.Context, client proto.BenchmarkServiceClient) {
	if *signature == "" {
		log.Fatal("--signature is required")
	}

	// Get transaction
	fmt.Printf("Getting transaction %s...\n", *signature)
	resp, err := client.GetTransaction(ctx, &proto.TransactionRequest{
		Signature:  *signature,
		Commitment: "finalized",
	})
	if err != nil {
		log.Fatalf("Error getting transaction: %v", err)
	}

	// Print results
	fmt.Printf("\nTransaction Info:\n")
	fmt.Printf("Signature: %s\n", resp.Signature)
	fmt.Printf("Slot: %d\n", resp.Slot)
	fmt.Printf("Success: %t\n", resp.Success)
	fmt.Printf("Transaction Data Length: %d bytes\n", len(resp.Transaction))
	fmt.Printf("Response Time: %d ms\n", resp.ResponseTimeMs)
}

func getBlock(ctx context.Context, client proto.BenchmarkServiceClient) {
	if *slot == 0 {
		log.Fatal("--slot is required")
	}

	// Get block
	fmt.Printf("Getting block at slot %d...\n", *slot)
	resp, err := client.GetBlock(ctx, &proto.BlockRequest{
		Slot:       *slot,
		Commitment: "finalized",
	})
	if err != nil {
		log.Fatalf("Error getting block: %v", err)
	}

	// Print results
	fmt.Printf("\nBlock Info:\n")
	fmt.Printf("Slot: %d\n", resp.Slot)
	fmt.Printf("Blockhash: %s\n", resp.Blockhash)
	fmt.Printf("Previous Blockhash: %s\n", resp.PreviousBlockhash)
	fmt.Printf("Parent Slot: %d\n", resp.ParentSlot)
	fmt.Printf("Transactions: %d\n", len(resp.Transactions))
	fmt.Printf("Response Time: %d ms\n", resp.ResponseTimeMs)
}

func streamAccounts(ctx context.Context, client proto.BenchmarkServiceClient) {
	if *pubkey == "" {
		log.Fatal("--pubkey is required")
	}

	// Stream account updates
	fmt.Printf("Streaming account updates for %s...\n", *pubkey)
	stream, err := client.StreamAccountUpdates(ctx, &proto.AccountStreamRequest{
		Pubkeys:    []string{*pubkey},
		Commitment: "finalized",
	})
	if err != nil {
		log.Fatalf("Error streaming account updates: %v", err)
	}

	// Receive updates
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving account update: %v", err)
		}

		// Print update
		fmt.Printf("\nAccount Update at %s:\n", time.Unix(int64(update.Timestamp), 0).Format(time.RFC3339))
		fmt.Printf("Pubkey: %s\n", update.Pubkey)
		fmt.Printf("Owner: %s\n", update.Owner)
		fmt.Printf("Lamports: %d\n", update.Lamports)
		fmt.Printf("Slot: %d\n", update.Slot)
		fmt.Printf("Data Length: %d bytes\n", len(update.Data))
	}
}

func streamTransactions(ctx context.Context, client proto.BenchmarkServiceClient) {
	// Stream transaction updates
	fmt.Println("Streaming transaction updates...")
	stream, err := client.StreamTransactions(ctx, &proto.TransactionStreamRequest{
		Accounts:      []string{},
		IncludeFailed: false,
		Commitment:    "finalized",
	})
	if err != nil {
		log.Fatalf("Error streaming transaction updates: %v", err)
	}

	// Receive updates
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving transaction update: %v", err)
		}

		// Print update
		fmt.Printf("\nTransaction Update at %s:\n", time.Unix(int64(update.Timestamp), 0).Format(time.RFC3339))
		fmt.Printf("Signature: %s\n", update.Signature)
		fmt.Printf("Slot: %d\n", update.Slot)
		fmt.Printf("Success: %t\n", update.Success)
		fmt.Printf("Transaction Data Length: %d bytes\n", len(update.Transaction))
	}
}

func streamBlocks(ctx context.Context, client proto.BenchmarkServiceClient) {
	// Stream block updates
	fmt.Println("Streaming block updates...")
	stream, err := client.StreamBlocks(ctx, &proto.BlockStreamRequest{
		Commitment: "finalized",
	})
	if err != nil {
		log.Fatalf("Error streaming block updates: %v", err)
	}

	// Receive updates
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error receiving block update: %v", err)
		}

		// Print update
		fmt.Printf("\nBlock Update at %s:\n", time.Unix(int64(update.Timestamp), 0).Format(time.RFC3339))
		fmt.Printf("Slot: %d\n", update.Slot)
		fmt.Printf("Blockhash: %s\n", update.Blockhash)
		fmt.Printf("Previous Blockhash: %s\n", update.PreviousBlockhash)
		fmt.Printf("Parent Slot: %d\n", update.ParentSlot)
	}
}