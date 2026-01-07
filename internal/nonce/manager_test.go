package nonce

import (
	"sync"
	"testing"
)

func TestManager_GetNonce_Sequential(t *testing.T) {
	// start from nonce 5
	manager := NewManager(5)

	// get 3 nonces sequentially
	n1 := manager.GetNonce()
	n2 := manager.GetNonce()
	n3 := manager.GetNonce()

	if n1 != 5 {
		t.Errorf("expected nonce 5, got %d", n1)
	}
	if n2 != 6 {
		t.Errorf("expected nonce 6, got %d", n2)
	}
	if n3 != 7 {
		t.Errorf("expected nonce 7, got %d", n3)
	}
}

func TestManager_GetNonce_Concurrent(t *testing.T) {
	manager := NewManager(0)

	numGoroutines := 100
	results := make(chan uint64, numGoroutines)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 100 goroutines getting nonces concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			nonce := manager.GetNonce()
			results <- nonce
		}()
	}

	wg.Wait()
	close(results)

	// collect all nonces
	seen := make(map[uint64]bool)
	for nonce := range results {
		if seen[nonce] {
			t.Errorf("duplicate nonce: %d", nonce)
		}
		seen[nonce] = true
	}

	// should have 100 unique nonces (0-99)
	if len(seen) != numGoroutines {
		t.Errorf("expected %d unique nonces, got %d", numGoroutines, len(seen))
	}
}

func TestManager_Reset(t *testing.T) {
	manager := NewManager(10)

	manager.GetNonce() // 10
	manager.GetNonce() // 11

	// reset to 5
	manager.Reset(5)

	n := manager.GetNonce()
	if n != 5 {
		t.Errorf("expected nonce 5 after reset, got %d", n)
	}
}

