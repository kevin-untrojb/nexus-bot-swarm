package swarm

import (
	"context"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/nexus-bot-swarm/domain"
	"github.com/nexus-bot-swarm/ports"
)

// Bot represents an individual trading bot in the swarm
type Bot struct {
	ID            int
	pool          *domain.Pool
	client        ports.BlockchainClient
	privateKey    string
	walletAddress string
}

// NewBot creates a new bot with the given ID and pool reference
func NewBot(id int, pool *domain.Pool) *Bot {
	return &Bot{
		ID:   id,
		pool: pool,
	}
}

// NewBotWithClient creates a bot that can send real transactions
func NewBotWithClient(id int, pool *domain.Pool, client ports.BlockchainClient, privateKey, walletAddress string) *Bot {
	return &Bot{
		ID:            id,
		pool:          pool,
		client:        client,
		privateKey:    privateKey,
		walletAddress: walletAddress,
	}
}

// CanSendRealTX returns true if bot is configured for real transactions
func (b *Bot) CanSendRealTX() bool {
	return b.client != nil && b.privateKey != "" && b.walletAddress != ""
}

// Run starts the bot's main loop
// It performs swaps until context is cancelled
// Sends any errors to errCh and closes it when done
func (b *Bot) Run(ctx context.Context, errCh chan<- error) {
	defer close(errCh)

	// simulated swap ticker (fast)
	swapTicker := time.NewTicker(500 * time.Millisecond)
	defer swapTicker.Stop()

	// real TX ticker (slow, every 5 seconds)
	var realTxTicker *time.Ticker
	if b.CanSendRealTX() {
		realTxTicker = time.NewTicker(5 * time.Second)
		defer realTxTicker.Stop()
		log.Printf("[Bot %d] Started (real TX enabled)", b.ID)
	} else {
		log.Printf("[Bot %d] Started (simulation only)", b.ID)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Bot %d] Shutting down", b.ID)
			errCh <- ctx.Err()
			return

		case <-swapTicker.C:
			b.performSwap()

		case <-func() <-chan time.Time {
			if realTxTicker != nil {
				return realTxTicker.C
			}
			return nil
		}():
			b.performRealTX(ctx)
		}
	}
}

// performSwap executes a random swap on the simulated pool
func (b *Bot) performSwap() {
	amount := big.NewInt(int64(rand.Intn(100) + 1))

	if rand.Intn(2) == 0 {
		out, err := b.pool.SwapAForB(amount)
		if err != nil {
			return
		}
		// less verbose logging
		if rand.Intn(10) == 0 {
			log.Printf("[Bot %d] Simulated: %s A -> %s B", b.ID, amount.String(), out.String())
		}
	} else {
		out, err := b.pool.SwapBForA(amount)
		if err != nil {
			return
		}
		if rand.Intn(10) == 0 {
			log.Printf("[Bot %d] Simulated: %s B -> %s A", b.ID, amount.String(), out.String())
		}
	}
}

// performRealTX sends a real transaction on the blockchain
func (b *Bot) performRealTX(ctx context.Context) {
	if !b.CanSendRealTX() {
		return
	}

	// send 1 wei to self
	amount := big.NewInt(1)

	log.Printf("[Bot %d] 📤 Sending real TX (1 wei to self)...", b.ID)

	txHash, err := b.client.SendETH(ctx, b.privateKey, b.walletAddress, amount)
	if err != nil {
		log.Printf("[Bot %d] ❌ Real TX failed: %v", b.ID, err)
		return
	}

	log.Printf("[Bot %d] ✅ Real TX sent: %s", b.ID, txHash)
}
