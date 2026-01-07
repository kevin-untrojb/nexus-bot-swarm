package nexus

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestClient_Connect_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	chainID := client.ChainID()
	if chainID == nil {
		t.Fatal("chain ID is nil after connect")
	}
	if chainID.Int64() != 3945 {
		t.Errorf("expected chain ID 3945, got %d", chainID.Int64())
	}
}

func TestClient_Connect_WrongChainID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// expect chain 1, but testnet is 3945
	client := NewClient("https://testnet.rpc.nexus.xyz", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		client.Close()
		t.Fatal("expected error for chain ID mismatch")
	}
}

func TestClient_Connect_InvalidURL(t *testing.T) {
	client := NewClient("http://invalid-url-that-does-not-exist.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		client.Close()
		t.Fatal("expected error for invalid URL")
	}
}

func TestClient_BlockNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	blockNum, err := client.BlockNumber(ctx)
	if err != nil {
		t.Fatalf("failed to get block number: %v", err)
	}

	if blockNum == 0 {
		t.Error("block number should be greater than 0")
	}

	t.Logf("current block number: %d", blockNum)
}

func TestClient_Balance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	address := os.Getenv("WALLET_ADDRESS")
	if address == "" {
		t.Skip("WALLET_ADDRESS not set, skipping balance test")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	balance, err := client.Balance(ctx, address)
	if err != nil {
		t.Fatalf("failed to get balance: %v", err)
	}

	t.Logf("balance of %s: %s wei", address, balance.String())
}

func TestClient_Balance_InvalidAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	_, err := client.Balance(ctx, "not-a-valid-address")
	if err == nil {
		t.Fatal("expected error for invalid address")
	}
}

func TestClient_SendETH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	privateKey := os.Getenv("NEXUS_PRIVATE_KEY")
	if privateKey == "" {
		t.Skip("NEXUS_PRIVATE_KEY not set, skipping send test")
	}

	address := os.Getenv("WALLET_ADDRESS")
	if address == "" {
		t.Skip("WALLET_ADDRESS not set, skipping send test")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	// send 1 wei to self (minimal amount)
	amount := big.NewInt(1)

	txHash, err := client.SendETH(ctx, privateKey, address, amount)
	if err != nil {
		t.Fatalf("failed to send ETH: %v", err)
	}

	t.Logf("tx sent: %s", txHash)

	if txHash == "" {
		t.Error("tx hash should not be empty")
	}
}

func TestClient_SendETH_InvalidPrivateKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	address := os.Getenv("WALLET_ADDRESS")
	if address == "" {
		t.Skip("WALLET_ADDRESS not set, skipping test")
	}

	client := NewClient("https://testnet.rpc.nexus.xyz", 3945)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	amount := big.NewInt(1)

	_, err := client.SendETH(ctx, "invalid-private-key", address, amount)
	if err == nil {
		t.Fatal("expected error for invalid private key")
	}
}
