package swarm

import (
	"context"
	"log"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/nexus-bot-swarm/domain"
	"github.com/nexus-bot-swarm/internal/nonce"
	"github.com/nexus-bot-swarm/ports"
)

// Bot represents an individual trading bot in the swarm
type Bot struct {
	ID            int
	pool          *domain.Pool
	client        ports.BlockchainClient
	privateKey    string
	walletAddress string
	nonceManager  *nonce.Manager
	tokenAddress  string // ERC20 token contract address
}

// NewBot creates a new bot with the given ID and pool reference
func NewBot(id int, pool *domain.Pool) *Bot {
	return &Bot{
		ID:   id,
		pool: pool,
	}
}

// NewBotWithClient creates a bot that can send real transactions
func NewBotWithClient(id int, pool *domain.Pool, client ports.BlockchainClient, privateKey, walletAddress, tokenAddress string, nonceManager *nonce.Manager) *Bot {
	return &Bot{
		ID:            id,
		pool:          pool,
		client:        client,
		privateKey:    privateKey,
		walletAddress: walletAddress,
		nonceManager:  nonceManager,
		tokenAddress:  tokenAddress,
	}
}

// CanSendRealTX returns true if bot is configured for real transactions
func (b *Bot) CanSendRealTX() bool {
	return b.client != nil && b.privateKey != "" && b.walletAddress != "" && b.nonceManager != nil
}

// CanTransferTokens returns true if bot is configured for ERC20 transfers
func (b *Bot) CanTransferTokens() bool {
	return b.CanSendRealTX() && b.tokenAddress != ""
}

// Run starts the bot's main loop
// It performs swaps until context is cancelled
// Sends any errors to errCh and closes it when done
func (b *Bot) Run(ctx context.Context, errCh chan<- error) {
	defer close(errCh)

	// simulated swap ticker (fast)
	swapTicker := time.NewTicker(500 * time.Millisecond)
	defer swapTicker.Stop()

	// real TX ticker (slow, every 10 seconds to avoid rate limiting)
	var realTxTicker *time.Ticker
	if b.CanSendRealTX() {
		realTxTicker = time.NewTicker(10 * time.Second)
		defer realTxTicker.Stop()
		log.Printf("[Bot %d] Started (real TX enabled with nonce manager)", b.ID)
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

	// get nonce from manager (atomic, no collisions)
	txNonce := b.nonceManager.GetNonce()

	var txHash string
	var err error

	if b.CanTransferTokens() {
		// Transfer ERC20 tokens (1 token = 1e18 wei of token)
		amount := big.NewInt(1000000000000000000) // 1 KEVZ token

		log.Printf("[Bot %d] 🪙 Transferring 1 KEVZ (nonce %d)...", b.ID, txNonce)

		txHash, err = b.client.TransferToken(ctx, b.tokenAddress, b.privateKey, b.walletAddress, amount, txNonce)
	} else {
		// Fallback: send 1 wei NEX to self
		amount := big.NewInt(1)

		log.Printf("[Bot %d] 📤 Sending 1 wei NEX (nonce %d)...", b.ID, txNonce)

		txHash, err = b.client.SendETHWithNonce(ctx, b.privateKey, b.walletAddress, amount, txNonce)
	}

	if err != nil {
		log.Printf("[Bot %d] ❌ TX failed (nonce %d): %v", b.ID, txNonce, err)

		// if nonce too low, sync with RPC
		if isNonceTooLowError(err) {
			b.syncNonce(ctx)
		}
		return
	}

	log.Printf("[Bot %d] ✅ TX sent (nonce %d): %s", b.ID, txNonce, txHash)
}

// isNonceTooLowError checks if the error is a nonce too low error
func isNonceTooLowError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "nonce too low") || strings.Contains(errStr, "already known")
}

// syncNonce fetches the current nonce from RPC and updates the manager
func (b *Bot) syncNonce(ctx context.Context) {
	newNonce, err := b.client.GetNonce(ctx, b.walletAddress)
	if err != nil {
		log.Printf("[Bot %d] ⚠️ Failed to sync nonce: %v", b.ID, err)
		return
	}

	currentNonce := b.nonceManager.Current()
	if newNonce > currentNonce {
		b.nonceManager.Reset(newNonce)
		log.Printf("[Bot %d] 🔄 Nonce synced: %d → %d", b.ID, currentNonce, newNonce)
	}
}
