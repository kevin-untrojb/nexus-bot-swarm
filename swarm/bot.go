package swarm

import (
	"context"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/nexus-bot-swarm/domain"
)

// Bot represents an individual trading bot in the swarm
type Bot struct {
	ID   int
	pool *domain.Pool
}

// NewBot creates a new bot with the given ID and pool reference
func NewBot(id int, pool *domain.Pool) *Bot {
	return &Bot{
		ID:   id,
		pool: pool,
	}
}

// Run starts the bot's main loop
// It performs swaps until context is cancelled
// Sends any errors to errCh and closes it when done
func (b *Bot) Run(ctx context.Context, errCh chan<- error) {
	defer close(errCh)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("[Bot %d] Started", b.ID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Bot %d] Shutting down", b.ID)
			errCh <- ctx.Err()
			return
		case <-ticker.C:
			b.performSwap()
		}
	}
}

// performSwap executes a random swap on the pool
func (b *Bot) performSwap() {
	// random amount between 1 and 100
	amount := big.NewInt(int64(rand.Intn(100) + 1))

	// randomly choose direction
	if rand.Intn(2) == 0 {
		out, err := b.pool.SwapAForB(amount)
		if err != nil {
			log.Printf("[Bot %d] SwapAForB error: %v", b.ID, err)
			return
		}
		log.Printf("[Bot %d] Swapped %s A -> %s B", b.ID, amount.String(), out.String())
	} else {
		out, err := b.pool.SwapBForA(amount)
		if err != nil {
			log.Printf("[Bot %d] SwapBForA error: %v", b.ID, err)
			return
		}
		log.Printf("[Bot %d] Swapped %s B -> %s A", b.ID, amount.String(), out.String())
	}
}

