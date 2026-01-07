package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nexus-bot-swarm/domain"
	"github.com/nexus-bot-swarm/internal/adapters/nexus"
	"github.com/nexus-bot-swarm/internal/config"
	"github.com/nexus-bot-swarm/swarm"
)

func main() {
	log.Println("🚀 Starting Nexus Bot Swarm...")

	// Load .env file (ignore error if not exists)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Printf("📋 Config loaded: RPC=%s, ChainID=%d, Bots=%d", cfg.RPCURL, cfg.ExpectedChainID, cfg.BotCount)

	// Create Nexus client
	client := nexus.NewClient(cfg.RPCURL, cfg.ExpectedChainID)

	// Connect with timeout
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()

	if err := client.Connect(connectCtx); err != nil {
		log.Fatalf("❌ Failed to connect to Nexus: %v", err)
	}
	defer client.Close()

	log.Printf("✅ Connected to Nexus Testnet (Chain ID: %d)", client.ChainID().Int64())

	// Show current block to prove connection works
	blockNum, err := client.BlockNumber(connectCtx)
	if err != nil {
		log.Fatalf("❌ Failed to get block number: %v", err)
	}
	log.Printf("📦 Current block: %d", blockNum)

	// Create simulated AMM pool
	// Initial reserves: 1000 ETH, 2000 USDC (in wei-like units)
	initialReserveA := big.NewInt(1000000000) // 1 billion units
	initialReserveB := big.NewInt(2000000000) // 2 billion units
	pool := domain.NewPool("ETH", "USDC", initialReserveA, initialReserveB)
	log.Printf("💱 AMM Pool created: %s/%s (ReserveA=%s, ReserveB=%s)",
		pool.TokenA, pool.TokenB, pool.ReserveA.String(), pool.ReserveB.String())
	log.Printf("💰 Initial price: 1 %s = %.4f %s", pool.TokenA, pool.PriceAInB(), pool.TokenB)

	// Create and start swarm
	ctx, cancel := context.WithCancel(context.Background())

	var botSwarm *swarm.Swarm
	if cfg.PrivateKey != "" && cfg.WalletAddress != "" {
		// get current nonce from RPC
		startNonce, err := client.GetNonce(connectCtx, cfg.WalletAddress)
		if err != nil {
			log.Fatalf("❌ Failed to get nonce: %v", err)
		}
		log.Printf("🔢 Starting nonce: %d", startNonce)

		// real TX mode with nonce manager
		botSwarm = swarm.NewSwarmWithClient(cfg.BotCount, pool, client, cfg.PrivateKey, cfg.WalletAddress, startNonce)
		log.Printf("🤖 Swarm started with %d bots (REAL TX MODE). Press Ctrl+C to stop...", cfg.BotCount)
		log.Printf("💸 Bots will send real transactions every 5 seconds (with nonce manager)")
	} else {
		// simulation only
		botSwarm = swarm.NewSwarm(cfg.BotCount, pool)
		log.Printf("🤖 Swarm started with %d bots (SIMULATION MODE). Press Ctrl+C to stop...", cfg.BotCount)
		log.Println("ℹ️  Set NEXUS_PRIVATE_KEY and WALLET_ADDRESS in .env for real TX")
	}

	errCh := botSwarm.Start(ctx)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Main loop: wait for signal or errors
	go func() {
		for err := range errCh {
			if err != nil && err != context.Canceled {
				log.Printf("⚠️ Bot error: %v", err)
			}
		}
	}()

	<-sigCh
	log.Println("\n🛑 Shutdown signal received...")

	// Cancel context to stop all bots
	cancel()

	// Wait briefly for bots to finish
	time.Sleep(100 * time.Millisecond)

	// Show final state
	log.Printf("📊 Final pool state: ReserveA=%s, ReserveB=%s",
		pool.ReserveA.String(), pool.ReserveB.String())
	log.Printf("💰 Final price: 1 %s = %.4f %s", pool.TokenA, pool.PriceAInB(), pool.TokenB)

	log.Println("👋 Goodbye!")
}
