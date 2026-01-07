package swarm

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/nexus-bot-swarm/domain"
)

func TestNewSwarm(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(1000), big.NewInt(2000))
	swarm := NewSwarm(3, pool)

	if len(swarm.bots) != 3 {
		t.Errorf("expected 3 bots, got %d", len(swarm.bots))
	}
}

func TestSwarm_Start_Stop(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(1000000), big.NewInt(2000000))
	swarm := NewSwarm(3, pool)

	ctx, cancel := context.WithCancel(context.Background())

	// start swarm
	errCh := swarm.Start(ctx)

	// let it run
	time.Sleep(100 * time.Millisecond)

	// stop
	cancel()

	// collect errors (should be nil or context.Canceled)
	errorCount := 0
	timeout := time.After(2 * time.Second)

loop:
	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				break loop
			}
			if err != nil && err != context.Canceled {
				t.Errorf("unexpected error: %v", err)
				errorCount++
			}
		case <-timeout:
			t.Error("swarm did not shutdown in time")
			break loop
		}
	}

	if errorCount > 0 {
		t.Errorf("got %d errors", errorCount)
	}
}

func TestSwarm_ConcurrentSwaps(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(10000000), big.NewInt(20000000))
	swarm := NewSwarm(5, pool)

	initialReserveA := new(big.Int).Set(pool.ReserveA)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := swarm.Start(ctx)

	// let bots swap
	time.Sleep(300 * time.Millisecond)

	cancel()

	// drain error channel
	for range errCh {
	}

	// verify swaps happened
	if pool.ReserveA.Cmp(initialReserveA) == 0 {
		t.Error("expected reserves to change from concurrent swaps")
	}

	t.Logf("after concurrent swaps: ReserveA=%s, ReserveB=%s",
		pool.ReserveA.String(), pool.ReserveB.String())
}
