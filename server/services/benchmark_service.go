package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/i-tozer/grpc-solana-trump/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BenchmarkService implements the gRPC benchmark service
type BenchmarkService struct {
	proto.UnimplementedBenchmarkServiceServer
	solanaClient *rpc.Client
	rpcEndpoint  string
}

// NewBenchmarkService creates a new benchmark service
func NewBenchmarkService(rpcEndpoint string) *BenchmarkService {
	client := rpc.New(rpcEndpoint)
	return &BenchmarkService{
		solanaClient: client,
		rpcEndpoint:  rpcEndpoint,
	}
}

// GetAccountInfo retrieves account information and measures performance
func (s *BenchmarkService) GetAccountInfo(ctx context.Context, req *proto.AccountInfoRequest) (*proto.AccountInfoResponse, error) {
	pubkey, err := solana.PublicKeyFromBase58(req.Pubkey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid pubkey: %v", err)
	}

	startTime := time.Now()

	// Get account info
	accountInfo, err := s.solanaClient.GetAccountInfo(ctx, pubkey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get account info: %v", err)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// Convert account info to response
	response := &proto.AccountInfoResponse{
		Pubkey:         req.Pubkey,
		Data:           accountInfo.Value.Data.GetBinary(),
		Owner:          accountInfo.Value.Owner.String(),
		Lamports:       accountInfo.Value.Lamports,
		Executable:     accountInfo.Value.Executable,
		RentEpoch:      accountInfo.Value.RentEpoch,
		ResponseTimeMs: uint64(responseTime),
	}

	return response, nil
}

// GetTransaction retrieves transaction information and measures performance
func (s *BenchmarkService) GetTransaction(ctx context.Context, req *proto.TransactionRequest) (*proto.TransactionResponse, error) {
	signature, err := solana.SignatureFromBase58(req.Signature)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid signature: %v", err)
	}

	startTime := time.Now()

	// Get transaction
	tx, err := s.solanaClient.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transaction: %v", err)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// Convert transaction to response
	response := &proto.TransactionResponse{
		Signature:      req.Signature,
		Slot:           tx.Slot,
		Transaction:    []byte(fmt.Sprintf("%v", tx.Transaction)), // Simplified for demo
		Success:        tx.Meta.Err == nil,
		ResponseTimeMs: uint64(responseTime),
	}

	return response, nil
}

// GetBlock retrieves block information and measures performance
func (s *BenchmarkService) GetBlock(ctx context.Context, req *proto.BlockRequest) (*proto.BlockResponse, error) {
	startTime := time.Now()

	// Get block
	block, err := s.solanaClient.GetBlock(ctx, uint64(req.Slot))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get block: %v", err)
	}

	responseTime := time.Since(startTime).Milliseconds()

	// Extract transaction signatures
	txSigs := make([]string, 0, len(block.Transactions))
	for _, tx := range block.Transactions {
		txWithMeta, err := tx.GetTransaction()
		if err != nil {
			continue
		}
		if len(txWithMeta.Signatures) > 0 {
			txSigs = append(txSigs, txWithMeta.Signatures[0].String())
		}
	}

	// Convert block to response
	response := &proto.BlockResponse{
		Slot:              req.Slot,
		Blockhash:         block.Blockhash.String(),
		PreviousBlockhash: block.PreviousBlockhash.String(),
		ParentSlot:        block.ParentSlot,
		Transactions:      txSigs,
		ResponseTimeMs:    uint64(responseTime),
	}

	return response, nil
}

// StreamAccountUpdates streams account updates in real-time
func (s *BenchmarkService) StreamAccountUpdates(req *proto.AccountStreamRequest, stream proto.BenchmarkService_StreamAccountUpdatesServer) error {
	// Convert pubkeys to solana.PublicKey
	pubkeys := make([]solana.PublicKey, 0, len(req.Pubkeys))
	for _, pubkeyStr := range req.Pubkeys {
		pubkey, err := solana.PublicKeyFromBase58(pubkeyStr)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid pubkey: %v", err)
		}
		pubkeys = append(pubkeys, pubkey)
	}

	// For demo purposes, we'll simulate account updates
	// In a real implementation, you would use WebSocket subscriptions
	for i := 0; i < 10; i++ {
		for _, pubkey := range pubkeys {
			// Get account info
			accountInfo, err := s.solanaClient.GetAccountInfo(context.Background(), pubkey)
			if err != nil {
				log.Printf("Error getting account info: %v", err)
				continue
			}

			// Send account update
			err = stream.Send(&proto.AccountUpdate{
				Pubkey:    pubkey.String(),
				Data:      accountInfo.Value.Data.GetBinary(),
				Owner:     accountInfo.Value.Owner.String(),
				Lamports:  accountInfo.Value.Lamports,
				Slot:      accountInfo.Context.Slot,
				Timestamp: uint64(time.Now().Unix()),
			})
			if err != nil {
				return status.Errorf(codes.Internal, "failed to send account update: %v", err)
			}
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// StreamTransactions streams transactions in real-time
func (s *BenchmarkService) StreamTransactions(req *proto.TransactionStreamRequest, stream proto.BenchmarkService_StreamTransactionsServer) error {
	// For demo purposes, we'll simulate transaction updates
	// In a real implementation, you would use WebSocket subscriptions
	for i := 0; i < 10; i++ {
		// Send transaction update
		err := stream.Send(&proto.TransactionUpdate{
			Signature:   "simulated_signature_" + fmt.Sprint(i),
			Slot:        uint64(100000 + i),
			Transaction: []byte("simulated_transaction_data"),
			Success:     true,
			Timestamp:   uint64(time.Now().Unix()),
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send transaction update: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// StreamBlocks streams blocks in real-time
func (s *BenchmarkService) StreamBlocks(req *proto.BlockStreamRequest, stream proto.BenchmarkService_StreamBlocksServer) error {
	// For demo purposes, we'll simulate block updates
	// In a real implementation, you would use WebSocket subscriptions
	for i := 0; i < 10; i++ {
		// Send block update
		err := stream.Send(&proto.BlockUpdate{
			Slot:              uint64(100000 + i),
			Blockhash:         fmt.Sprintf("simulated_blockhash_%d", i),
			PreviousBlockhash: fmt.Sprintf("simulated_previous_blockhash_%d", i-1),
			ParentSlot:        uint64(100000 + i - 1),
			Timestamp:         uint64(time.Now().Unix()),
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send block update: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	return nil
}

// RunBenchmark runs a comprehensive benchmark suite and returns results
func (s *BenchmarkService) RunBenchmark(ctx context.Context, req *proto.BenchmarkRequest) (*proto.BenchmarkResults, error) {
	// Validate request
	if req.Iterations == 0 {
		req.Iterations = 10 // Default to 10 iterations
	}

	// Prepare results
	results := &proto.BenchmarkResults{
		Iterations: req.Iterations,
	}

	startTime := time.Now()

	// Run gRPC benchmarks if requested
	if req.RunGrpcTests {
		grpcResults, err := s.runGrpcBenchmarks(ctx, req)
		if err != nil {
			return nil, err
		}
		results.GrpcResults = grpcResults
	}

	// Run JSON-RPC benchmarks if requested
	if req.RunJsonrpcTests {
		jsonrpcResults, err := s.runJsonrpcBenchmarks(ctx, req)
		if err != nil {
			return nil, err
		}
		results.JsonrpcResults = jsonrpcResults
	}

	// Calculate overall speedup if both tests were run
	if req.RunGrpcTests && req.RunJsonrpcTests && len(results.GrpcResults) > 0 && len(results.JsonrpcResults) > 0 {
		var grpcTotalTime, jsonrpcTotalTime uint64
		for _, result := range results.GrpcResults {
			grpcTotalTime += result.AvgResponseTimeMs
		}
		for _, result := range results.JsonrpcResults {
			jsonrpcTotalTime += result.AvgResponseTimeMs
		}

		if grpcTotalTime > 0 && jsonrpcTotalTime > 0 {
			results.OverallSpeedup = float64(jsonrpcTotalTime) / float64(grpcTotalTime)
		}
	}

	results.TotalTimeMs = uint64(time.Since(startTime).Milliseconds())

	return results, nil
}

// runGrpcBenchmarks runs gRPC benchmarks
func (s *BenchmarkService) runGrpcBenchmarks(ctx context.Context, req *proto.BenchmarkRequest) ([]*proto.BenchmarkResults_MethodResult, error) {
	var results []*proto.BenchmarkResults_MethodResult

	// Benchmark GetAccountInfo
	if len(req.TestAccounts) > 0 {
		result, err := s.benchmarkGetAccountInfo(ctx, req.TestAccounts, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	// Benchmark GetTransaction
	if len(req.TestSignatures) > 0 {
		result, err := s.benchmarkGetTransaction(ctx, req.TestSignatures, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	// Benchmark GetBlock
	if len(req.TestSlots) > 0 {
		result, err := s.benchmarkGetBlock(ctx, req.TestSlots, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// runJsonrpcBenchmarks runs JSON-RPC benchmarks
func (s *BenchmarkService) runJsonrpcBenchmarks(ctx context.Context, req *proto.BenchmarkRequest) ([]*proto.BenchmarkResults_MethodResult, error) {
	var results []*proto.BenchmarkResults_MethodResult

	// Create a new JSON-RPC client
	jsonrpcClient := rpc.New(req.SolanaRpcUrl)
	if jsonrpcClient == nil {
		jsonrpcClient = s.solanaClient // Fall back to the default client
	}

	// Benchmark GetAccountInfo
	if len(req.TestAccounts) > 0 {
		result, err := s.benchmarkJsonrpcGetAccountInfo(ctx, jsonrpcClient, req.TestAccounts, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	// Benchmark GetTransaction
	if len(req.TestSignatures) > 0 {
		result, err := s.benchmarkJsonrpcGetTransaction(ctx, jsonrpcClient, req.TestSignatures, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	// Benchmark GetBlock
	if len(req.TestSlots) > 0 {
		result, err := s.benchmarkJsonrpcGetBlock(ctx, jsonrpcClient, req.TestSlots, req.Iterations)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// benchmarkGetAccountInfo benchmarks the GetAccountInfo method
func (s *BenchmarkService) benchmarkGetAccountInfo(ctx context.Context, accounts []string, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "GetAccountInfo",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, account := range accounts {
			resp, err := s.GetAccountInfo(ctx, &proto.AccountInfoRequest{
				Pubkey: account,
			})
			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += resp.ResponseTimeMs
			if resp.ResponseTimeMs < minTime {
				minTime = resp.ResponseTimeMs
			}
			if resp.ResponseTimeMs > maxTime {
				maxTime = resp.ResponseTimeMs
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}

// benchmarkGetTransaction benchmarks the GetTransaction method
func (s *BenchmarkService) benchmarkGetTransaction(ctx context.Context, signatures []string, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "GetTransaction",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, signature := range signatures {
			resp, err := s.GetTransaction(ctx, &proto.TransactionRequest{
				Signature: signature,
			})
			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += resp.ResponseTimeMs
			if resp.ResponseTimeMs < minTime {
				minTime = resp.ResponseTimeMs
			}
			if resp.ResponseTimeMs > maxTime {
				maxTime = resp.ResponseTimeMs
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}

// benchmarkGetBlock benchmarks the GetBlock method
func (s *BenchmarkService) benchmarkGetBlock(ctx context.Context, slots []uint64, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "GetBlock",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, slot := range slots {
			resp, err := s.GetBlock(ctx, &proto.BlockRequest{
				Slot: slot,
			})
			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += resp.ResponseTimeMs
			if resp.ResponseTimeMs < minTime {
				minTime = resp.ResponseTimeMs
			}
			if resp.ResponseTimeMs > maxTime {
				maxTime = resp.ResponseTimeMs
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}

// benchmarkJsonrpcGetAccountInfo benchmarks the JSON-RPC GetAccountInfo method
func (s *BenchmarkService) benchmarkJsonrpcGetAccountInfo(ctx context.Context, client *rpc.Client, accounts []string, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "JSON-RPC GetAccountInfo",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, account := range accounts {
			pubkey, err := solana.PublicKeyFromBase58(account)
			if err != nil {
				failureCount++
				continue
			}

			startTime := time.Now()
			_, err = client.GetAccountInfo(ctx, pubkey)
			responseTime := uint64(time.Since(startTime).Milliseconds())

			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += responseTime
			if responseTime < minTime {
				minTime = responseTime
			}
			if responseTime > maxTime {
				maxTime = responseTime
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}

// benchmarkJsonrpcGetTransaction benchmarks the JSON-RPC GetTransaction method
func (s *BenchmarkService) benchmarkJsonrpcGetTransaction(ctx context.Context, client *rpc.Client, signatures []string, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "JSON-RPC GetTransaction",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, signatureStr := range signatures {
			signature, err := solana.SignatureFromBase58(signatureStr)
			if err != nil {
				failureCount++
				continue
			}

			startTime := time.Now()
			_, err = client.GetTransaction(ctx, signature, &rpc.GetTransactionOpts{})
			responseTime := uint64(time.Since(startTime).Milliseconds())

			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += responseTime
			if responseTime < minTime {
				minTime = responseTime
			}
			if responseTime > maxTime {
				maxTime = responseTime
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}

// benchmarkJsonrpcGetBlock benchmarks the JSON-RPC GetBlock method
func (s *BenchmarkService) benchmarkJsonrpcGetBlock(ctx context.Context, client *rpc.Client, slots []uint64, iterations uint32) (*proto.BenchmarkResults_MethodResult, error) {
	result := &proto.BenchmarkResults_MethodResult{
		Method: "JSON-RPC GetBlock",
	}

	var totalTime uint64
	var minTime uint64 = ^uint64(0) // Max uint64 value
	var maxTime uint64 = 0
	var successCount uint32 = 0
	var failureCount uint32 = 0

	for i := uint32(0); i < iterations; i++ {
		for _, slot := range slots {
			startTime := time.Now()
			_, err := client.GetBlock(ctx, slot)
			responseTime := uint64(time.Since(startTime).Milliseconds())

			if err != nil {
				failureCount++
				continue
			}

			successCount++
			totalTime += responseTime
			if responseTime < minTime {
				minTime = responseTime
			}
			if responseTime > maxTime {
				maxTime = responseTime
			}
		}
	}

	if successCount > 0 {
		result.AvgResponseTimeMs = totalTime / uint64(successCount)
		result.MinResponseTimeMs = minTime
		result.MaxResponseTimeMs = maxTime
	}
	result.SuccessCount = successCount
	result.FailureCount = failureCount

	return result, nil
}