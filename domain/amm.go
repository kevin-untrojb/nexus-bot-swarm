package domain

import (
	"fmt"
	"math/big"
	"sync"
)

// Pool represents an AMM liquidity pool with two tokens
// Uses constant product formula: x * y = k
// Thread-safe for concurrent access
type Pool struct {
	mu       sync.Mutex
	TokenA   string
	TokenB   string
	ReserveA *big.Int
	ReserveB *big.Int
}

// NewPool creates a new liquidity pool
func NewPool(tokenA, tokenB string, reserveA, reserveB *big.Int) *Pool {
	return &Pool{
		TokenA:   tokenA,
		TokenB:   tokenB,
		ReserveA: new(big.Int).Set(reserveA),
		ReserveB: new(big.Int).Set(reserveB),
	}
}

// SwapAForB swaps amountIn of TokenA for TokenB
// Returns the amount of TokenB received
// Formula: dy = (y * dx) / (x + dx)
func (p *Pool) SwapAForB(amountIn *big.Int) (*big.Int, error) {
	if err := p.validateAmount(amountIn); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// dy = (y * dx) / (x + dx)
	numerator := new(big.Int).Mul(p.ReserveB, amountIn)
	denominator := new(big.Int).Add(p.ReserveA, amountIn)
	amountOut := new(big.Int).Div(numerator, denominator)

	// update reserves
	p.ReserveA.Add(p.ReserveA, amountIn)
	p.ReserveB.Sub(p.ReserveB, amountOut)

	return amountOut, nil
}

// SwapBForA swaps amountIn of TokenB for TokenA
// Returns the amount of TokenA received
// Formula: dx = (x * dy) / (y + dy)
func (p *Pool) SwapBForA(amountIn *big.Int) (*big.Int, error) {
	if err := p.validateAmount(amountIn); err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// dx = (x * dy) / (y + dy)
	numerator := new(big.Int).Mul(p.ReserveA, amountIn)
	denominator := new(big.Int).Add(p.ReserveB, amountIn)
	amountOut := new(big.Int).Div(numerator, denominator)

	// update reserves
	p.ReserveB.Add(p.ReserveB, amountIn)
	p.ReserveA.Sub(p.ReserveA, amountOut)

	return amountOut, nil
}

// PriceAInB returns the price of TokenA in terms of TokenB
func (p *Pool) PriceAInB() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	a := new(big.Float).SetInt(p.ReserveA)
	b := new(big.Float).SetInt(p.ReserveB)
	price := new(big.Float).Quo(b, a)
	result, _ := price.Float64()
	return result
}

// PriceBInA returns the price of TokenB in terms of TokenA
func (p *Pool) PriceBInA() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	a := new(big.Float).SetInt(p.ReserveA)
	b := new(big.Float).SetInt(p.ReserveB)
	price := new(big.Float).Quo(a, b)
	result, _ := price.Float64()
	return result
}

func (p *Pool) validateAmount(amount *big.Int) error {
	if amount.Sign() <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}
