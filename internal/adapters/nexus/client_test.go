package nexus

import (
	"context"
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
