package nexus

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
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
// Gets nonce automatically from the RPC (not safe for concurrent use with same wallet)
func (c *Client) SendETH(ctx context.Context, privateKeyHex string, to string, amount *big.Int) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	// parse private key to get from address
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to get public key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// get nonce from RPC
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	return c.SendETHWithNonce(ctx, privateKeyHex, to, amount, nonce)
}

// SendETHWithNonce sends native currency with a specific nonce
// Use this for concurrent bots with a shared nonce manager
func (c *Client) SendETHWithNonce(ctx context.Context, privateKeyHex string, to string, amount *big.Int, nonce uint64) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	// parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// validate to address
	if !common.IsHexAddress(to) {
		return "", fmt.Errorf("invalid to address: %s", to)
	}
	toAddress := common.HexToAddress(to)

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

// GetNonce returns the current pending nonce for an address
func (c *Client) GetNonce(ctx context.Context, address string) (uint64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("client not connected")
	}
	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("invalid address: %s", address)
	}
	return c.client.PendingNonceAt(ctx, common.HexToAddress(address))
}

// Close gracefully closes the RPC connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// =============================================================================
// ERC20 Token Methods
// =============================================================================

// TokenBalance returns the token balance of an address
func (c *Client) TokenBalance(ctx context.Context, tokenAddress string, walletAddress string) (*big.Int, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	if !common.IsHexAddress(tokenAddress) || !common.IsHexAddress(walletAddress) {
		return nil, fmt.Errorf("invalid address")
	}

	token := common.HexToAddress(tokenAddress)
	wallet := common.HexToAddress(walletAddress)

	// balanceOf(address) selector = keccak256("balanceOf(address)")[:4] = 0x70a08231
	selector := []byte{0x70, 0xa0, 0x82, 0x31}

	// ABI encode the address (32 bytes, left-padded)
	paddedAddress := common.LeftPadBytes(wallet.Bytes(), 32)

	// data = selector + paddedAddress
	data := append(selector, paddedAddress...)

	// Call the contract
	result, err := c.client.CallContract(ctx, ethereum.CallMsg{
		To:   &token,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	// Parse result (uint256)
	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// TransferToken sends ERC20 tokens to an address
func (c *Client) TransferToken(ctx context.Context, tokenAddress string, privateKeyHex string, to string, amount *big.Int, nonce uint64) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	// parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// validate addresses
	if !common.IsHexAddress(tokenAddress) || !common.IsHexAddress(to) {
		return "", fmt.Errorf("invalid address")
	}

	token := common.HexToAddress(tokenAddress)
	toAddress := common.HexToAddress(to)

	// transfer(address,uint256) selector = keccak256("transfer(address,uint256)")[:4] = 0xa9059cbb
	selector := []byte{0xa9, 0x05, 0x9c, 0xbb}

	// ABI encode: address (32 bytes) + uint256 (32 bytes)
	paddedTo := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	// data = selector + paddedTo + paddedAmount
	data := append(selector, paddedTo...)
	data = append(data, paddedAmount...)

	// estimate gas price
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// gas limit for token transfer (higher than simple ETH transfer)
	gasLimit := uint64(100000)

	// create transaction (value = 0, we're calling a contract)
	tx := types.NewTransaction(nonce, token, big.NewInt(0), gasLimit, gasPrice, data)

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
