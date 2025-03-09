package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/i-tozer/grpc-solana-trump/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// Jupiter API endpoints
	jupiterAPIBaseURL = "https://api.jup.ag/price/v2"
	
	// Birdeye API endpoints
	birdeyeAPIBaseURL = "https://public-api.birdeye.so"
	birdeyePriceEndpoint = "/public/price"
	birdeyeHistoryEndpoint = "/public/history_price"
	
	// USDC token mint address on Solana
	usdcTokenMint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC mint address
)

// TokenPriceService implements the gRPC token price service
type TokenPriceService struct {
	proto.UnimplementedTokenPriceServiceServer
	solanaClient *rpc.Client
	rpcEndpoint  string
	httpClient   *http.Client
	birdeyeAPIKey string
	trumpTokenMint string
}

// NewTokenPriceService creates a new token price service
func NewTokenPriceService(rpcEndpoint string, birdeyeAPIKey string) *TokenPriceService {
	client := rpc.New(rpcEndpoint)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Get TRUMP token mint from command line flag
	trumpTokenMint := "3vHSsSV1iRoLbspdNtRcHN3nwmYqGJ5hRQQFSvfUHYJz" // Default TRUMP token mint
	
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	
	return &TokenPriceService{
		solanaClient: client,
		rpcEndpoint:  rpcEndpoint,
		httpClient:   httpClient,
		birdeyeAPIKey: birdeyeAPIKey,
		trumpTokenMint: trumpTokenMint,
	}
}

// JupiterPriceResponse represents the response from Jupiter API v2
type JupiterPriceResponse struct {
	Data map[string]struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Price string `json:"price"`
	} `json:"data"`
	TimeTaken float64 `json:"timeTaken"`
}

// JupiterPriceHistoryResponse represents the response from Jupiter API for price history
type JupiterPriceHistoryResponse struct {
	Data struct {
		ID     string `json:"id"`
		MintSymbol string `json:"mintSymbol"`
		VsToken string `json:"vsToken"`
		VsTokenSymbol string `json:"vsTokenSymbol"`
		PriceHistory []struct {
			Price     float64 `json:"price"`
			UnixTime  int64   `json:"unixTime"`
		} `json:"priceHistory"`
	} `json:"data"`
	TimeTaken float64 `json:"timeTaken"`
}

// BirdeyePriceResponse represents the response from Birdeye API
type BirdeyePriceResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Value            float64 `json:"value"`
		UpdatedAt        int64   `json:"updateAt"`
		PriceChange      struct {
			Volume24h      float64 `json:"volume24h"`
			PriceChange24h float64 `json:"priceChange24h"`
		} `json:"priceChange"`
	} `json:"data"`
}

// BirdeyeHistoryResponse represents the response from Birdeye API for price history
type BirdeyeHistoryResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
		Volume    float64 `json:"volume,omitempty"`
		Open      float64 `json:"open,omitempty"`
		High      float64 `json:"high,omitempty"`
		Low       float64 `json:"low,omitempty"`
		Close     float64 `json:"close,omitempty"`
	} `json:"data"`
}

// GetTokenPrice retrieves the current price of a token
func (s *TokenPriceService) GetTokenPrice(ctx context.Context, req *proto.TokenPriceRequest) (*proto.TokenPriceResponse, error) {
	// Use default TRUMP token if not specified
	tokenMint := req.TokenMint
	if tokenMint == "" {
		tokenMint = s.trumpTokenMint
	}
	
	// Use default USDC as quote token if not specified
	quoteMint := req.QuoteMint
	if quoteMint == "" {
		quoteMint = usdcTokenMint
	}
	
	// Default to Jupiter DEX if not specified
	dex := req.Dex
	if dex == "" {
		dex = "jupiter"
	}
	
	var response *proto.TokenPriceResponse
	var err error
	
	switch dex {
	case "jupiter":
		response, err = s.getSimulatedTokenPrice(ctx, tokenMint, quoteMint)
	case "birdeye":
		response, err = s.getSimulatedTokenPrice(ctx, tokenMint, quoteMint)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported DEX: %s", dex)
	}
	
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// getSimulatedTokenPrice returns simulated token price data
func (s *TokenPriceService) getSimulatedTokenPrice(ctx context.Context, tokenMint, quoteMint string) (*proto.TokenPriceResponse, error) {
	// Simulate a price for TRUMP token
	// In a real implementation, you would fetch this from a DEX or price oracle
	price := 0.15 + rand.Float64()*0.02 // Random price between 0.15 and 0.17 USDC
	
	// Construct response
	response := &proto.TokenPriceResponse{
		TokenMint:      tokenMint,
		TokenName:      "TRUMP",
		TokenSymbol:    "TRUMP",
		QuoteMint:      quoteMint,
		QuoteSymbol:    "USDC",
		Price:          price,
		Volume24h:      1000000 + rand.Float64()*500000, // Random volume
		PriceChange24h: (rand.Float64()*4 - 2), // Random price change between -2% and 2%
		Timestamp:      uint64(time.Now().Unix()),
		Dex:            "jupiter",
	}
	
	return response, nil
}

// getJupiterTokenPrice fetches token price from Jupiter API
func (s *TokenPriceService) getJupiterTokenPrice(ctx context.Context, tokenMint, quoteMint string) (*proto.TokenPriceResponse, error) {
	// Construct Jupiter API URL
	apiURL := fmt.Sprintf("%s?ids=%s&vsToken=%s", jupiterAPIBaseURL, tokenMint, quoteMint)
	
	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create request: %v", err)
	}
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch price data: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response: %v", err)
	}
	
	// Parse response
	var jupiterResp JupiterPriceResponse
	if err := json.Unmarshal(body, &jupiterResp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse response: %v", err)
	}
	
	// Extract price data for the token
	tokenData, ok := jupiterResp.Data[tokenMint]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "token price data not found")
	}
	
	// Parse price as float64
	price, err := parsePrice(tokenData.Price)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse price: %v", err)
	}
	
	// Construct response
	response := &proto.TokenPriceResponse{
		TokenMint:      tokenMint,
		TokenName:      "TRUMP", // Hardcoded for now
		TokenSymbol:    "TRUMP", // Hardcoded for now
		QuoteMint:      quoteMint,
		QuoteSymbol:    "USDC", // Assuming USDC as quote token
		Price:          price,
		Volume24h:      1000000, // Placeholder
		PriceChange24h: 0,       // Placeholder
		Timestamp:      uint64(time.Now().Unix()),
		Dex:            "jupiter",
	}
	
	return response, nil
}

// parsePrice parses a price string to float64
func parsePrice(priceStr string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(priceStr, "%f", &price)
	if err != nil {
		return 0, err
	}
	return price, nil
}

// getBirdeyeTokenPrice fetches token price from Birdeye API
func (s *TokenPriceService) getBirdeyeTokenPrice(ctx context.Context, tokenMint, quoteMint string) (*proto.TokenPriceResponse, error) {
	if s.birdeyeAPIKey == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "Birdeye API key not configured")
	}
	
	// Construct Birdeye API URL
	apiURL := fmt.Sprintf("%s%s?address=%s", birdeyeAPIBaseURL, birdeyePriceEndpoint, tokenMint)
	
	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create request: %v", err)
	}
	
	// Add API key header
	req.Header.Add("X-API-KEY", s.birdeyeAPIKey)
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch price data: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response: %v", err)
	}
	
	// Parse response
	var birdeyeResp BirdeyePriceResponse
	if err := json.Unmarshal(body, &birdeyeResp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse response: %v", err)
	}
	
	// Construct response
	response := &proto.TokenPriceResponse{
		TokenMint:      tokenMint,
		TokenName:      "", // Birdeye API doesn't provide token name in this endpoint
		TokenSymbol:    "", // Birdeye API doesn't provide token symbol in this endpoint
		QuoteMint:      quoteMint,
		QuoteSymbol:    "USDC", // Assuming USDC as quote token
		Price:          birdeyeResp.Data.Value,
		Volume24h:      birdeyeResp.Data.PriceChange.Volume24h,
		PriceChange24h: birdeyeResp.Data.PriceChange.PriceChange24h,
		Timestamp:      uint64(birdeyeResp.Data.UpdatedAt),
		Dex:            "birdeye",
	}
	
	return response, nil
}

// GetTokenPriceHistory retrieves historical price data for a token
func (s *TokenPriceService) GetTokenPriceHistory(ctx context.Context, req *proto.TokenPriceHistoryRequest) (*proto.TokenPriceHistoryResponse, error) {
	// Use default TRUMP token if not specified
	tokenMint := req.TokenMint
	if tokenMint == "" {
		tokenMint = s.trumpTokenMint
	}
	
	// Use default USDC as quote token if not specified
	quoteMint := req.QuoteMint
	if quoteMint == "" {
		quoteMint = usdcTokenMint
	}
	
	// Default to Jupiter DEX if not specified
	dex := req.Dex
	if dex == "" {
		dex = "jupiter"
	}
	
	// Default interval to 1h if not specified
	interval := req.Interval
	if interval == "" {
		interval = "1h"
	}
	
	// Default start time to 7 days ago if not specified
	startTime := req.StartTime
	if startTime == 0 {
		startTime = uint64(time.Now().AddDate(0, 0, -7).Unix())
	}
	
	// Default end time to now if not specified
	endTime := req.EndTime
	if endTime == 0 {
		endTime = uint64(time.Now().Unix())
	}
	
	// Default limit to 100 if not specified
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}
	
	// Generate simulated historical data
	response, err := s.getSimulatedTokenPriceHistory(ctx, tokenMint, quoteMint, interval, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	
	return response, nil
}

// getSimulatedTokenPriceHistory generates simulated historical price data
func (s *TokenPriceService) getSimulatedTokenPriceHistory(ctx context.Context, tokenMint, quoteMint, interval string, startTime, endTime uint64, limit uint32) (*proto.TokenPriceHistoryResponse, error) {
	// Generate simulated historical data points
	dataPoints := make([]*proto.PriceDataPoint, 0, limit)
	basePrice := 0.15 // Base price for TRUMP token
	timeStep := (endTime - startTime) / uint64(limit)
	
	// Create a realistic price trend with some volatility
	for i := uint32(0); i < limit; i++ {
		timestamp := startTime + uint64(i)*timeStep
		
		// Create a trend with some random noise
		trend := 0.02 * math.Sin(float64(i)/10.0) // Sinusoidal trend
		noise := 0.005 * (rand.Float64()*2.0 - 1.0) // Random noise
		
		// Calculate price with trend and noise
		price := basePrice * (1.0 + trend + noise)
		
		// Add some volatility for open, high, low prices
		open := price * (1.0 + 0.01*(rand.Float64()*2.0-1.0))
		high := price * (1.0 + 0.02*rand.Float64())
		low := price * (1.0 - 0.02*rand.Float64())
		
		// Ensure high is the highest and low is the lowest
		if open > high {
			high = open
		}
		if open < low {
			low = open
		}
		
		// Create data point
		dataPoint := &proto.PriceDataPoint{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     price,
			Volume:    500000.0 + 500000.0*rand.Float64(), // Random volume between 500K and 1M
		}
		dataPoints = append(dataPoints, dataPoint)
	}
	
	// Construct response
	response := &proto.TokenPriceHistoryResponse{
		TokenMint:    tokenMint,
		TokenName:    "TRUMP",
		TokenSymbol:  "TRUMP",
		QuoteMint:    quoteMint,
		QuoteSymbol:  "USDC",
		Interval:     interval,
		DataPoints:   dataPoints,
		Dex:          "jupiter",
	}
	
	return response, nil
}

// getJupiterTokenPriceHistory fetches historical token price data from Jupiter API
func (s *TokenPriceService) getJupiterTokenPriceHistory(ctx context.Context, tokenMint, quoteMint, interval string, startTime, endTime uint64, limit uint32) (*proto.TokenPriceHistoryResponse, error) {
	// For now, we'll simulate historical data since Jupiter v2 API doesn't have a history endpoint
	// In a real implementation, you would use a proper history API or fetch from a database
	
	// Get current price
	currentPrice, err := s.getJupiterTokenPrice(ctx, tokenMint, quoteMint)
	if err != nil {
		return nil, err
	}
	
	// Generate simulated historical data points
	dataPoints := make([]*proto.PriceDataPoint, 0, limit)
	basePrice := currentPrice.Price
	timeStep := (endTime - startTime) / uint64(limit)
	
	for i := uint32(0); i < limit; i++ {
		timestamp := startTime + uint64(i)*timeStep
		
		// Add some random variation to the price
		variation := 0.05 * (float64(i%10) - 5.0) / 5.0
		price := basePrice * (1.0 + variation)
		
		dataPoint := &proto.PriceDataPoint{
			Timestamp: timestamp,
			Open:      price * 0.99,
			High:      price * 1.02,
			Low:       price * 0.98,
			Close:     price,
			Volume:    1000000.0 * (1.0 + variation),
		}
		dataPoints = append(dataPoints, dataPoint)
	}
	
	// Construct response
	response := &proto.TokenPriceHistoryResponse{
		TokenMint:    tokenMint,
		TokenName:    "TRUMP", // Hardcoded for now
		TokenSymbol:  "TRUMP", // Hardcoded for now
		QuoteMint:    quoteMint,
		QuoteSymbol:  "USDC", // Assuming USDC as quote token
		Interval:     interval,
		DataPoints:   dataPoints,
		Dex:          "jupiter",
	}
	
	return response, nil
}

// getBirdeyeTokenPriceHistory fetches historical token price data from Birdeye API
func (s *TokenPriceService) getBirdeyeTokenPriceHistory(ctx context.Context, tokenMint, quoteMint, interval string, startTime, endTime uint64, limit uint32) (*proto.TokenPriceHistoryResponse, error) {
	if s.birdeyeAPIKey == "" {
		return nil, status.Errorf(codes.FailedPrecondition, "Birdeye API key not configured")
	}
	
	// Map interval to Birdeye interval format
	birdeyeInterval := "1H" // Default to 1 hour
	switch interval {
	case "1m":
		birdeyeInterval = "1M"
	case "5m":
		birdeyeInterval = "5M"
	case "15m":
		birdeyeInterval = "15M"
	case "1h":
		birdeyeInterval = "1H"
	case "4h":
		birdeyeInterval = "4H"
	case "1d":
		birdeyeInterval = "1D"
	}
	
	// Construct Birdeye API URL
	apiURL := fmt.Sprintf("%s%s?address=%s&type=%s&time_from=%d&time_to=%d", 
		birdeyeAPIBaseURL, birdeyeHistoryEndpoint, tokenMint, birdeyeInterval, startTime, endTime)
	
	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create request: %v", err)
	}
	
	// Add API key header
	req.Header.Add("X-API-KEY", s.birdeyeAPIKey)
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch price history: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read response: %v", err)
	}
	
	// Parse response
	var birdeyeResp BirdeyeHistoryResponse
	if err := json.Unmarshal(body, &birdeyeResp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse response: %v", err)
	}
	
	// Convert price history to data points
	dataPoints := make([]*proto.PriceDataPoint, 0, len(birdeyeResp.Data))
	for _, point := range birdeyeResp.Data {
		dataPoint := &proto.PriceDataPoint{
			Timestamp: uint64(point.Timestamp),
			Open:      point.Open,
			High:      point.High,
			Low:       point.Low,
			Close:     point.Close,
			Volume:    point.Volume,
		}
		
		// If OHLC data is not available, use the value field
		if dataPoint.Open == 0 && dataPoint.High == 0 && dataPoint.Low == 0 && dataPoint.Close == 0 {
			dataPoint.Open = point.Value
			dataPoint.High = point.Value
			dataPoint.Low = point.Value
			dataPoint.Close = point.Value
		}
		
		dataPoints = append(dataPoints, dataPoint)
	}
	
	// Limit the number of data points if needed
	if uint32(len(dataPoints)) > limit {
		dataPoints = dataPoints[:limit]
	}
	
	// Construct response
	response := &proto.TokenPriceHistoryResponse{
		TokenMint:    tokenMint,
		TokenName:    "", // Birdeye API doesn't provide token name in this endpoint
		TokenSymbol:  "", // Birdeye API doesn't provide token symbol in this endpoint
		QuoteMint:    quoteMint,
		QuoteSymbol:  "USDC", // Assuming USDC as quote token
		Interval:     interval,
		DataPoints:   dataPoints,
		Dex:          "birdeye",
	}
	
	return response, nil
}

// StreamTokenPrice streams real-time price updates for a token
func (s *TokenPriceService) StreamTokenPrice(req *proto.TokenPriceStreamRequest, stream proto.TokenPriceService_StreamTokenPriceServer) error {
	// Use default TRUMP token if not specified
	tokenMint := req.TokenMint
	if tokenMint == "" {
		tokenMint = s.trumpTokenMint
	}
	
	// Use default USDC as quote token if not specified
	quoteMint := req.QuoteMint
	if quoteMint == "" {
		quoteMint = usdcTokenMint
	}
	
	// Default to Jupiter DEX if not specified
	dex := req.Dex
	if dex == "" {
		dex = "jupiter"
	}
	
	// Default update interval to 5 seconds if not specified
	updateInterval := req.UpdateIntervalMs
	if updateInterval == 0 {
		updateInterval = 5000 // 5 seconds
	}
	
	// Create ticker for periodic updates
	ticker := time.NewTicker(time.Duration(updateInterval) * time.Millisecond)
	defer ticker.Stop()
	
	// Context for cancellation
	ctx := stream.Context()
	
	// Stream price updates
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Fetch current price
			priceResp, err := s.getSimulatedTokenPrice(ctx, tokenMint, quoteMint)
			if err != nil {
				log.Printf("Error fetching price: %v", err)
				continue
			}
			
			// Send price update
			update := &proto.TokenPriceUpdate{
				TokenMint:      priceResp.TokenMint,
				TokenSymbol:    priceResp.TokenSymbol,
				QuoteSymbol:    priceResp.QuoteSymbol,
				Price:          priceResp.Price,
				Volume24h:      priceResp.Volume24h,
				PriceChange24h: priceResp.PriceChange24h,
				Timestamp:      priceResp.Timestamp,
				Dex:            priceResp.Dex,
			}
			
			if err := stream.Send(update); err != nil {
				return status.Errorf(codes.Internal, "failed to send price update: %v", err)
			}
		}
	}
}