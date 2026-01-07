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

	// BlockNumber returns the current block number
	BlockNumber(ctx context.Context) (uint64, error)

	// Balance returns the balance of an address in wei
	Balance(ctx context.Context, address string) (*big.Int, error)

	// Close gracefully closes the connection
	Close()
}
