package domain

import (
	"math/big"
	"sync"
	"testing"
)

func TestNewPool(t *testing.T) {
	reserveA := big.NewInt(1000)
	reserveB := big.NewInt(2000)

	pool := NewPool("ETH", "USDC", reserveA, reserveB)

	if pool.TokenA != "ETH" {
		t.Errorf("expected TokenA=ETH, got %s", pool.TokenA)
	}
	if pool.TokenB != "USDC" {
		t.Errorf("expected TokenB=USDC, got %s", pool.TokenB)
	}
	if pool.ReserveA.Cmp(reserveA) != 0 {
		t.Errorf("expected ReserveA=1000, got %s", pool.ReserveA.String())
	}
	if pool.ReserveB.Cmp(reserveB) != 0 {
		t.Errorf("expected ReserveB=2000, got %s", pool.ReserveB.String())
	}
}

func TestPool_SwapAForB(t *testing.T) {
	// pool with 1000 ETH and 2000 USDC
	reserveA := big.NewInt(1000)
	reserveB := big.NewInt(2000)
	pool := NewPool("ETH", "USDC", reserveA, reserveB)

	// swap 100 ETH for USDC
	amountIn := big.NewInt(100)
	amountOut, err := pool.SwapAForB(amountIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// formula: dy = (y * dx) / (x + dx)
	// dy = (2000 * 100) / (1000 + 100) = 200000 / 1100 = 181
	expected := big.NewInt(181)
	if amountOut.Cmp(expected) != 0 {
		t.Errorf("expected amountOut=%s, got %s", expected.String(), amountOut.String())
	}

	// check reserves updated
	// new reserveA = 1000 + 100 = 1100
	// new reserveB = 2000 - 181 = 1819
	expectedReserveA := big.NewInt(1100)
	expectedReserveB := big.NewInt(1819)

	if pool.ReserveA.Cmp(expectedReserveA) != 0 {
		t.Errorf("expected ReserveA=%s, got %s", expectedReserveA.String(), pool.ReserveA.String())
	}
	if pool.ReserveB.Cmp(expectedReserveB) != 0 {
		t.Errorf("expected ReserveB=%s, got %s", expectedReserveB.String(), pool.ReserveB.String())
	}
}

func TestPool_SwapBForA(t *testing.T) {
	// pool with 1000 ETH and 2000 USDC
	reserveA := big.NewInt(1000)
	reserveB := big.NewInt(2000)
	pool := NewPool("ETH", "USDC", reserveA, reserveB)

	// swap 200 USDC for ETH
	amountIn := big.NewInt(200)
	amountOut, err := pool.SwapBForA(amountIn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// formula: dx = (x * dy) / (y + dy)
	// dx = (1000 * 200) / (2000 + 200) = 200000 / 2200 = 90
	expected := big.NewInt(90)
	if amountOut.Cmp(expected) != 0 {
		t.Errorf("expected amountOut=%s, got %s", expected.String(), amountOut.String())
	}
}

func TestPool_SwapZeroAmount(t *testing.T) {
	pool := NewPool("ETH", "USDC", big.NewInt(1000), big.NewInt(2000))

	_, err := pool.SwapAForB(big.NewInt(0))
	if err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestPool_SwapNegativeAmount(t *testing.T) {
	pool := NewPool("ETH", "USDC", big.NewInt(1000), big.NewInt(2000))

	_, err := pool.SwapAForB(big.NewInt(-100))
	if err == nil {
		t.Error("expected error for negative amount")
	}
}

func TestPool_GetPrice(t *testing.T) {
	// pool with 1000 ETH and 2000 USDC
	// price of A in terms of B = reserveB / reserveA = 2000/1000 = 2
	pool := NewPool("ETH", "USDC", big.NewInt(1000), big.NewInt(2000))

	priceAInB := pool.PriceAInB()
	// 2000 / 1000 = 2
	if priceAInB != 2.0 {
		t.Errorf("expected price 2.0, got %f", priceAInB)
	}

	priceBInA := pool.PriceBInA()
	// 1000 / 2000 = 0.5
	if priceBInA != 0.5 {
		t.Errorf("expected price 0.5, got %f", priceBInA)
	}
}

func TestPool_ConcurrentSwaps(t *testing.T) {
	// large reserves to handle many swaps
	pool := NewPool("ETH", "USDC", big.NewInt(1000000), big.NewInt(2000000))

	numGoroutines := 100
	swapsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < swapsPerGoroutine; j++ {
				// alternate between swap directions
				if j%2 == 0 {
					pool.SwapAForB(big.NewInt(10))
				} else {
					pool.SwapBForA(big.NewInt(10))
				}
			}
		}()
	}

	wg.Wait()

	// if we got here without race detector panicking, we're good
	// just verify reserves are still positive
	if pool.ReserveA.Sign() <= 0 {
		t.Error("ReserveA should be positive")
	}
	if pool.ReserveB.Sign() <= 0 {
		t.Error("ReserveB should be positive")
	}

	t.Logf("after %d concurrent swaps: ReserveA=%s, ReserveB=%s",
		numGoroutines*swapsPerGoroutine,
		pool.ReserveA.String(),
		pool.ReserveB.String())
}
