package ports

import (
	"context"
	"math/big"
)

// BlockchainClient defines the interface for interacting with any EVM blockchain
// This is the PORT in hexagonal architecture - implementations are adapters
type BlockchainClient interface {
	// Connect establishes connection and validates chain ID
	Connect(ctx context.Context) error

	// ChainID returns the connected chain's ID
	ChainID() *big.Int

	// Close gracefully closes the connection
	Close()
}
