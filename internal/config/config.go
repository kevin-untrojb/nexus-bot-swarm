package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the bot swarm
type Config struct {
	// Nexus RPC endpoint
	RPCURL string

	// Expected chain ID (Nexus Testnet III = 3945)
	ExpectedChainID int64

	// Number of bots in the swarm
	BotCount int

	// Wallet address for balance queries
	WalletAddress string

	// Private key for signing transactions
	PrivateKey string

	// ERC20 token contract address (KevzToken)
	TokenAddress string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	rpcURL := os.Getenv("NEXUS_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://testnet.rpc.nexus.xyz" // default but not hardcoded in logic
	}

	chainIDStr := os.Getenv("NEXUS_CHAIN_ID")
	if chainIDStr == "" {
		chainIDStr = "3945"
	}
	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid NEXUS_CHAIN_ID: %w", err)
	}

	botCountStr := os.Getenv("BOT_COUNT")
	if botCountStr == "" {
		botCountStr = "3"
	}
	botCount, err := strconv.Atoi(botCountStr)
	if err != nil {
		return nil, fmt.Errorf("invalid BOT_COUNT: %w", err)
	}

	walletAddress := os.Getenv("WALLET_ADDRESS")
	privateKey := os.Getenv("NEXUS_PRIVATE_KEY")
	tokenAddress := os.Getenv("TOKEN_ADDRESS")

	return &Config{
		RPCURL:          rpcURL,
		ExpectedChainID: chainID,
		BotCount:        botCount,
		WalletAddress:   walletAddress,
		PrivateKey:      privateKey,
		TokenAddress:    tokenAddress,
	}, nil
}
