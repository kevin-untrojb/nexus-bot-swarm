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

	// SendETH sends native currency to an address, returns tx hash
	SendETH(ctx context.Context, privateKey string, to string, amount *big.Int) (string, error)

	// SendETHWithNonce sends native currency with a specific nonce (for concurrent use)
	SendETHWithNonce(ctx context.Context, privateKey string, to string, amount *big.Int, nonce uint64) (string, error)

	// GetNonce returns the current pending nonce for an address
	GetNonce(ctx context.Context, address string) (uint64, error)

	// TokenBalance returns the ERC20 token balance of an address
	TokenBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error)

	// TransferToken sends ERC20 tokens to an address
	TransferToken(ctx context.Context, tokenAddress string, privateKey string, to string, amount *big.Int, nonce uint64) (string, error)

	// Close gracefully closes the connection
	Close()
}
