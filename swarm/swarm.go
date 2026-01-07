package swarm

import (
	"context"
	"log"
	"sync"

	"github.com/nexus-bot-swarm/domain"
	"github.com/nexus-bot-swarm/internal/nonce"
	"github.com/nexus-bot-swarm/ports"
)

// Swarm coordinates multiple bots operating on a shared pool
type Swarm struct {
	bots         []*Bot
	pool         *domain.Pool
	nonceManager *nonce.Manager
}

// NewSwarm creates a swarm with the specified number of bots (simulation only)
func NewSwarm(botCount int, pool *domain.Pool) *Swarm {
	bots := make([]*Bot, botCount)
	for i := 0; i < botCount; i++ {
		bots[i] = NewBot(i+1, pool)
	}
	return &Swarm{
		bots: bots,
		pool: pool,
	}
}

// NewSwarmWithClient creates a swarm that can send real transactions
// startNonce should be fetched from the RPC before calling this
// tokenAddress is optional - if provided, bots will transfer ERC20 tokens instead of NEX
func NewSwarmWithClient(botCount int, pool *domain.Pool, client ports.BlockchainClient, privateKey, walletAddress, tokenAddress string, startNonce uint64) *Swarm {
	// create shared nonce manager
	nm := nonce.NewManager(startNonce)

	bots := make([]*Bot, botCount)
	for i := 0; i < botCount; i++ {
		// all bots share the same nonce manager
		bots[i] = NewBotWithClient(i+1, pool, client, privateKey, walletAddress, tokenAddress, nm)
	}
	return &Swarm{
		bots:         bots,
		pool:         pool,
		nonceManager: nm,
	}
}

// Start launches all bots and returns a channel for errors
// The channel is closed when all bots have stopped
func (s *Swarm) Start(ctx context.Context) <-chan error {
	// buffered channel to collect errors from all bots
	errCh := make(chan error, len(s.bots))

	var wg sync.WaitGroup
	wg.Add(len(s.bots))

	log.Printf("Starting swarm with %d bots", len(s.bots))

	for _, bot := range s.bots {
		go func(b *Bot) {
			defer wg.Done()

			// each bot gets its own error channel
			botErrCh := make(chan error, 1)
			go b.Run(ctx, botErrCh)

			// wait for bot to finish and forward error
			for err := range botErrCh {
				if err != nil {
					errCh <- err
				}
			}
		}(bot)
	}

	// close errCh when all bots are done
	go func() {
		wg.Wait()
		close(errCh)
		log.Println("Swarm stopped")
	}()

	return errCh
}

// Pool returns the shared pool for inspection
func (s *Swarm) Pool() *domain.Pool {
	return s.pool
}

// BotCount returns the number of bots in the swarm
func (s *Swarm) BotCount() int {
	return len(s.bots)
}
