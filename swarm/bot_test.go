package swarm

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/nexus-bot-swarm/domain"
)

func TestNewBot(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(1000), big.NewInt(2000))
	bot := NewBot(1, pool)

	if bot.ID != 1 {
		t.Errorf("expected ID=1, got %d", bot.ID)
	}
	if bot.pool != pool {
		t.Error("pool not set correctly")
	}
}

func TestBot_Run_Shutdown(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(1000000), big.NewInt(2000000))
	bot := NewBot(1, pool)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go bot.Run(ctx, errCh)

	// let it run briefly
	time.Sleep(100 * time.Millisecond)

	// shutdown
	cancel()

	// wait for clean exit
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("bot did not shutdown in time")
	}
}

func TestBot_Run_PerformsSwaps(t *testing.T) {
	pool := domain.NewPool("ETH", "USDC", big.NewInt(1000000), big.NewInt(2000000))
	bot := NewBot(1, pool)

	initialReserveA := new(big.Int).Set(pool.ReserveA)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go bot.Run(ctx, errCh)

	// let it do some swaps
	time.Sleep(250 * time.Millisecond)

	cancel()
	<-errCh

	// reserves should have changed
	if pool.ReserveA.Cmp(initialReserveA) == 0 {
		t.Error("expected reserves to change after swaps")
	}
}

