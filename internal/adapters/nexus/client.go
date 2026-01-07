package nexus

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Client implements ports.BlockchainClient for Nexus testnet
type Client struct {
	rpcURL          string
	expectedChainID int64
	client          *ethclient.Client
	chainID         *big.Int
}

// NewClient creates a new Nexus client (does not connect yet)
func NewClient(rpcURL string, expectedChainID int64) *Client {
	return &Client{
		rpcURL:          rpcURL,
		expectedChainID: expectedChainID,
	}
}

// Connect establishes connection to Nexus RPC and validates chain ID
func (c *Client) Connect(ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, c.rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC %s: %w", c.rpcURL, err)
	}

	// Validate chain ID as requested - fail fast if wrong network
	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != c.expectedChainID {
		client.Close()
		return fmt.Errorf("chain ID mismatch: expected %d, got %d", c.expectedChainID, chainID.Int64())
	}

	c.client = client
	c.chainID = chainID
	return nil
}

// ChainID returns the connected chain's ID
func (c *Client) ChainID() *big.Int {
	return c.chainID
}

// Close gracefully closes the RPC connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
