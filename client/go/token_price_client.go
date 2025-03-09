package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/i-tozer/grpc-solana-trump/proto"
	"github.com/wcharczuk/go-chart/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tokenPriceCmd = flag.NewFlagSet("token-price", flag.ExitOnError)
	tokenPriceHistoryCmd = flag.NewFlagSet("token-price-history", flag.ExitOnError)
	tokenPriceStreamCmd = flag.NewFlagSet("token-price-stream", flag.ExitOnError)
	
	// Common flags
	tokenMint = tokenPriceCmd.String("token-mint", "", "Token mint address (default: TRUMP token)")
	quoteMint = tokenPriceCmd.String("quote-mint", "", "Quote token mint address (default: USDC)")
	dex = tokenPriceCmd.String("dex", "jupiter", "DEX name (jupiter or birdeye)")
	
	// History flags
	interval = tokenPriceHistoryCmd.String("interval", "1h", "Time interval (1m, 5m, 15m, 1h, 4h, 1d)")
	startTime = tokenPriceHistoryCmd.Int64("start-time", 0, "Start timestamp (Unix timestamp in seconds, default: 7 days ago)")
	endTime = tokenPriceHistoryCmd.Int64("end-time", 0, "End timestamp (Unix timestamp in seconds, default: now)")
	limit = tokenPriceHistoryCmd.Int("limit", 100, "Maximum number of data points to return")
	outputFile = tokenPriceHistoryCmd.String("output", "price_chart.png", "Output chart file path")
	
	// Stream flags
	updateInterval = tokenPriceStreamCmd.Int("update-interval", 5000, "Update interval in milliseconds")
	duration = tokenPriceStreamCmd.Int("duration", 60, "Stream duration in seconds")
)

func init() {
	// Add common flags to other command sets
	tokenPriceHistoryCmd.String("token-mint", "", "Token mint address (default: TRUMP token)")
	tokenPriceHistoryCmd.String("quote-mint", "", "Quote token mint address (default: USDC)")
	tokenPriceHistoryCmd.String("dex", "jupiter", "DEX name (jupiter or birdeye)")
	
	tokenPriceStreamCmd.String("token-mint", "", "Token mint address (default: TRUMP token)")
	tokenPriceStreamCmd.String("quote-mint", "", "Quote token mint address (default: USDC)")
	tokenPriceStreamCmd.String("dex", "jupiter", "DEX name (jupiter or birdeye)")
}

func getTokenPrice(ctx context.Context, client proto.TokenPriceServiceClient) {
	// Get token price
	fmt.Printf("Getting token price...\n")
	resp, err := client.GetTokenPrice(ctx, &proto.TokenPriceRequest{
		TokenMint: *tokenMint,
		QuoteMint: *quoteMint,
		Dex:       *dex,
	})
	if err != nil {
		log.Fatalf("Error getting token price: %v", err)
	}

	// Print results
	fmt.Printf("\nToken Price Info:\n")
	fmt.Printf("Token: %s (%s)\n", resp.TokenSymbol, resp.TokenMint)
	fmt.Printf("Price: %.6f %s\n", resp.Price, resp.QuoteSymbol)
	fmt.Printf("24h Volume: %.2f %s\n", resp.Volume24h, resp.QuoteSymbol)
	fmt.Printf("24h Price Change: %.2f%%\n", resp.PriceChange24h)
	fmt.Printf("Timestamp: %s\n", time.Unix(int64(resp.Timestamp), 0).Format(time.RFC3339))
	fmt.Printf("DEX: %s\n", resp.Dex)
}

func getTokenPriceHistory(ctx context.Context, client proto.TokenPriceServiceClient) {
	// Parse flags
	tokenMint := tokenPriceHistoryCmd.Lookup("token-mint").Value.String()
	quoteMint := tokenPriceHistoryCmd.Lookup("quote-mint").Value.String()
	dex := tokenPriceHistoryCmd.Lookup("dex").Value.String()
	
	// Get token price history
	fmt.Printf("Getting token price history...\n")
	resp, err := client.GetTokenPriceHistory(ctx, &proto.TokenPriceHistoryRequest{
		TokenMint:  tokenMint,
		QuoteMint:  quoteMint,
		Dex:        dex,
		Interval:   *interval,
		StartTime:  uint64(*startTime),
		EndTime:    uint64(*endTime),
		Limit:      uint32(*limit),
	})
	if err != nil {
		log.Fatalf("Error getting token price history: %v", err)
	}

	// Print summary
	fmt.Printf("\nToken Price History:\n")
	fmt.Printf("Token: %s (%s)\n", resp.TokenSymbol, resp.TokenMint)
	fmt.Printf("Quote: %s\n", resp.QuoteSymbol)
	fmt.Printf("Interval: %s\n", resp.Interval)
	fmt.Printf("DEX: %s\n", resp.Dex)
	fmt.Printf("Data Points: %d\n", len(resp.DataPoints))
	
	if len(resp.DataPoints) == 0 {
		log.Fatalf("No price data points returned")
	}
	
	// Create chart
	fmt.Printf("Creating price chart...\n")
	createPriceChart(resp, *outputFile)
	fmt.Printf("Chart saved to %s\n", *outputFile)
	
	// Print data as JSON
	dataJSON, _ := json.MarshalIndent(resp.DataPoints, "", "  ")
	fmt.Printf("\nPrice Data (first 5 points):\n")
	if len(resp.DataPoints) > 5 {
		dataJSON, _ = json.MarshalIndent(resp.DataPoints[:5], "", "  ")
	}
	fmt.Println(string(dataJSON))
}

func streamTokenPrice(ctx context.Context, client proto.TokenPriceServiceClient) {
	// Parse flags
	tokenMint := tokenPriceStreamCmd.Lookup("token-mint").Value.String()
	quoteMint := tokenPriceStreamCmd.Lookup("quote-mint").Value.String()
	dex := tokenPriceStreamCmd.Lookup("dex").Value.String()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(*duration)*time.Second)
	defer cancel()
	
	// Stream token price updates
	fmt.Printf("Streaming token price updates for %d seconds...\n", *duration)
	stream, err := client.StreamTokenPrice(ctx, &proto.TokenPriceStreamRequest{
		TokenMint:        tokenMint,
		QuoteMint:        quoteMint,
		Dex:              dex,
		UpdateIntervalMs: uint32(*updateInterval),
	})
	if err != nil {
		log.Fatalf("Error streaming token price: %v", err)
	}
	
	// Collect price updates
	var prices []float64
	var timestamps []time.Time
	
	for {
		update, err := stream.Recv()
		if err != nil {
			break
		}
		
		// Print update
		timestamp := time.Unix(int64(update.Timestamp), 0)
		fmt.Printf("\nPrice Update at %s:\n", timestamp.Format(time.RFC3339))
		fmt.Printf("Token: %s\n", update.TokenSymbol)
		fmt.Printf("Price: %.6f %s\n", update.Price, update.QuoteSymbol)
		fmt.Printf("24h Volume: %.2f %s\n", update.Volume24h, update.QuoteSymbol)
		fmt.Printf("24h Price Change: %.2f%%\n", update.PriceChange24h)
		
		// Store data for chart
		prices = append(prices, update.Price)
		timestamps = append(timestamps, timestamp)
	}
	
	// Create live chart if we have enough data points
	if len(prices) > 1 {
		fmt.Printf("\nCreating live price chart...\n")
		createLiveChart(prices, timestamps, "live_price_chart.png")
		fmt.Printf("Live chart saved to live_price_chart.png\n")
	}
}

func createPriceChart(resp *proto.TokenPriceHistoryResponse, outputFile string) {
	// Extract data for chart
	var xValues []time.Time
	var openValues []float64
	var closeValues []float64
	var highValues []float64
	var lowValues []float64
	var volumeValues []float64
	
	for _, point := range resp.DataPoints {
		timestamp := time.Unix(int64(point.Timestamp), 0)
		xValues = append(xValues, timestamp)
		openValues = append(openValues, point.Open)
		closeValues = append(closeValues, point.Close)
		highValues = append(highValues, point.High)
		lowValues = append(lowValues, point.Low)
		volumeValues = append(volumeValues, point.Volume)
	}
	
	// Create price series
	mainSeries := chart.TimeSeries{
		Name:    fmt.Sprintf("%s/%s", resp.TokenSymbol, resp.QuoteSymbol),
		XValues: xValues,
		YValues: closeValues,
		Style: chart.Style{
			StrokeColor: chart.ColorBlue,
			StrokeWidth: 2,
		},
	}
	
	// Create volume series (if available)
	var volumeSeries chart.TimeSeries
	hasVolume := false
	for _, v := range volumeValues {
		if v > 0 {
			hasVolume = true
			break
		}
	}
	
	if hasVolume {
		volumeSeries = chart.TimeSeries{
			Name:    "Volume",
			XValues: xValues,
			YValues: volumeValues,
			Style: chart.Style{
				StrokeColor: chart.ColorGreen,
				StrokeWidth: 1,
				FillColor:   chart.ColorGreen.WithAlpha(64),
			},
		}
	}
	
	// Create chart
	graph := chart.Chart{
		Title: fmt.Sprintf("%s/%s Price Chart (%s)", resp.TokenSymbol, resp.QuoteSymbol, resp.Interval),
		TitleStyle: chart.Style{
			FontSize: 14,
		},
		XAxis: chart.XAxis{
			Name:           "Time",
			NameStyle:      chart.Style{FontSize: 12},
			Style:          chart.Style{FontSize: 10},
			ValueFormatter: chart.TimeValueFormatterWithFormat(time.RFC3339),
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
		},
		YAxis: chart.YAxis{
			Name:      fmt.Sprintf("Price (%s)", resp.QuoteSymbol),
			NameStyle: chart.Style{FontSize: 12},
			Style:     chart.Style{FontSize: 10},
			GridMajorStyle: chart.Style{
				StrokeColor: chart.ColorAlternateGray,
				StrokeWidth: 1.0,
			},
		},
		Series: []chart.Series{
			mainSeries,
		},
	}
	
	// Add volume as a secondary y-axis if available
	if hasVolume {
		graph.YAxisSecondary = chart.YAxis{
			Name:      "Volume",
			NameStyle: chart.Style{FontSize: 12},
			Style:     chart.Style{FontSize: 10},
		}
		graph.Series = append(graph.Series, volumeSeries)
	}
	
	// Save chart to file
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating chart file: %v", err)
	}
	defer f.Close()
	
	err = graph.Render(chart.PNG, f)
	if err != nil {
		log.Fatalf("Error rendering chart: %v", err)
	}
}

func createLiveChart(prices []float64, timestamps []time.Time, outputFile string) {
	// Create price series
	mainSeries := chart.TimeSeries{
		Name:    "Live Price",
		XValues: timestamps,
		YValues: prices,
		Style: chart.Style{
			StrokeColor: chart.ColorRed,
			StrokeWidth: 2,
		},
	}
	
	// Create chart
	graph := chart.Chart{
		Title: "Live Token Price Chart",
		TitleStyle: chart.Style{
			FontSize: 14,
		},
		XAxis: chart.XAxis{
			Name:           "Time",
			NameStyle:      chart.Style{FontSize: 12},
			Style:          chart.Style{FontSize: 10},
			ValueFormatter: chart.TimeValueFormatterWithFormat(time.RFC3339),
		},
		YAxis: chart.YAxis{
			Name:      "Price",
			NameStyle: chart.Style{FontSize: 12},
			Style:     chart.Style{FontSize: 10},
		},
		Series: []chart.Series{
			mainSeries,
		},
	}
	
	// Save chart to file
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Error creating chart file: %v", err)
	}
	defer f.Close()
	
	err = graph.Render(chart.PNG, f)
	if err != nil {
		log.Fatalf("Error rendering chart: %v", err)
	}
}

func runTokenPriceClient() {
	// Check if we have a command
	if len(os.Args) < 3 {
		fmt.Println("Usage: client token-price [command] [options]")
		fmt.Println("Commands: get, history, stream")
		os.Exit(1)
	}
	
	// Parse command
	command := os.Args[2]
	
	// Set up gRPC connection
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	
	// Create client
	client := proto.NewTokenPriceServiceClient(conn)
	ctx := context.Background()
	
	// Execute command
	switch command {
	case "get":
		tokenPriceCmd.Parse(os.Args[3:])
		getTokenPrice(ctx, client)
	case "history":
		tokenPriceHistoryCmd.Parse(os.Args[3:])
		getTokenPriceHistory(ctx, client)
	case "stream":
		tokenPriceStreamCmd.Parse(os.Args[3:])
		streamTokenPrice(ctx, client)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Commands: get, history, stream")
		os.Exit(1)
	}
}