package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus-bot-swarm/internal/adapters/nexus"
	"github.com/nexus-bot-swarm/internal/config"
)

func main() {
	log.Println("🚀 Starting Nexus Bot Swarm...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Printf("📋 Config loaded: RPC=%s, ChainID=%d, Bots=%d", cfg.RPCURL, cfg.ExpectedChainID, cfg.BotCount)

	// Create Nexus client
	client := nexus.NewClient(cfg.RPCURL, cfg.ExpectedChainID)

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("❌ Failed to connect to Nexus: %v", err)
	}
	defer client.Close()

	log.Printf("✅ Connected to Nexus Testnet (Chain ID: %d)", client.ChainID().Int64())

	// Show current block to prove connection works
	blockNum, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get block number: %v", err)
	}
	log.Printf("📦 Current block: %d", blockNum)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Println("⏳ Bot swarm running. Press Ctrl+C to stop...")
	<-sigCh

	log.Println("👋 Shutting down...")
}
