package nexus

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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

// BlockNumber returns the current block number
func (c *Client) BlockNumber(ctx context.Context) (uint64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("client not connected")
	}
	return c.client.BlockNumber(ctx)
}

// Balance returns the balance of an address in wei
func (c *Client) Balance(ctx context.Context, address string) (*big.Int, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("invalid address: %s", address)
	}

	addr := common.HexToAddress(address)
	return c.client.BalanceAt(ctx, addr, nil)
}

// SendETH sends native currency to an address, returns tx hash
func (c *Client) SendETH(ctx context.Context, privateKeyHex string, to string, amount *big.Int) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	// parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// get public key and address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to get public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// validate to address
	if !common.IsHexAddress(to) {
		return "", fmt.Errorf("invalid to address: %s", to)
	}
	toAddress := common.HexToAddress(to)

	// get nonce (automatic)
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// estimate gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// gas limit for simple transfer
	gasLimit := uint64(21000)

	// create transaction
	tx := types.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)

	// sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	// send transaction
	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send tx: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// Close gracefully closes the RPC connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
