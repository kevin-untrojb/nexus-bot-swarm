package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// clear env vars to test defaults
	os.Unsetenv("NEXUS_RPC_URL")
	os.Unsetenv("NEXUS_CHAIN_ID")
	os.Unsetenv("BOT_COUNT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RPCURL != "https://testnet.rpc.nexus.xyz" {
		t.Errorf("expected default RPC URL, got %s", cfg.RPCURL)
	}
	if cfg.ExpectedChainID != 3945 {
		t.Errorf("expected chain ID 3945, got %d", cfg.ExpectedChainID)
	}
	if cfg.BotCount != 3 {
		t.Errorf("expected bot count 3, got %d", cfg.BotCount)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("NEXUS_RPC_URL", "http://localhost:8545")
	os.Setenv("NEXUS_CHAIN_ID", "1337")
	os.Setenv("BOT_COUNT", "5")
	defer func() {
		os.Unsetenv("NEXUS_RPC_URL")
		os.Unsetenv("NEXUS_CHAIN_ID")
		os.Unsetenv("BOT_COUNT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RPCURL != "http://localhost:8545" {
		t.Errorf("expected custom RPC URL, got %s", cfg.RPCURL)
	}
	if cfg.ExpectedChainID != 1337 {
		t.Errorf("expected chain ID 1337, got %d", cfg.ExpectedChainID)
	}
	if cfg.BotCount != 5 {
		t.Errorf("expected bot count 5, got %d", cfg.BotCount)
	}
}

func TestLoad_InvalidChainID(t *testing.T) {
	os.Setenv("NEXUS_CHAIN_ID", "not-a-number")
	defer os.Unsetenv("NEXUS_CHAIN_ID")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid chain ID")
	}
}

func TestLoad_InvalidBotCount(t *testing.T) {
	os.Setenv("BOT_COUNT", "abc")
	defer os.Unsetenv("BOT_COUNT")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid bot count")
	}
}
