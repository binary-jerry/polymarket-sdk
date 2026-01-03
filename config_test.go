package polymarket

import (
	"testing"
	"time"
)

func TestConstants(t *testing.T) {
	// Test ChainID
	if ChainID != 137 {
		t.Errorf("ChainID = %d, expected 137", ChainID)
	}

	// Test API endpoints
	if GammaEndpoint != "https://gamma-api.polymarket.com" {
		t.Errorf("GammaEndpoint = %s, expected https://gamma-api.polymarket.com", GammaEndpoint)
	}
	if CLOBEndpoint != "https://clob.polymarket.com" {
		t.Errorf("CLOBEndpoint = %s, expected https://clob.polymarket.com", CLOBEndpoint)
	}
	if WSEndpoint != "wss://ws-subscriptions-clob.polymarket.com/ws/market" {
		t.Errorf("WSEndpoint = %s, expected wss://ws-subscriptions-clob.polymarket.com/ws/market", WSEndpoint)
	}

	// Test contract addresses
	if CTFExchangeAddress != "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e" {
		t.Errorf("CTFExchangeAddress mismatch")
	}
	if NegRiskCTFExchangeAddress != "0xC5d563A36AE78145C45a50134d48A1215220f80a" {
		t.Errorf("NegRiskCTFExchangeAddress mismatch")
	}
	if NegRiskAdapterAddress != "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296" {
		t.Errorf("NegRiskAdapterAddress mismatch")
	}
	if CollateralAddress != "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174" {
		t.Errorf("CollateralAddress mismatch")
	}
	if ConditionalTokensAddress != "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045" {
		t.Errorf("ConditionalTokensAddress mismatch")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test API endpoints
	if config.GammaEndpoint != GammaEndpoint {
		t.Errorf("GammaEndpoint = %s, expected %s", config.GammaEndpoint, GammaEndpoint)
	}
	if config.CLOBEndpoint != CLOBEndpoint {
		t.Errorf("CLOBEndpoint = %s, expected %s", config.CLOBEndpoint, CLOBEndpoint)
	}
	if config.WSEndpoint != WSEndpoint {
		t.Errorf("WSEndpoint = %s, expected %s", config.WSEndpoint, WSEndpoint)
	}

	// Test HTTP config
	if config.HTTPTimeout != 30*time.Second {
		t.Errorf("HTTPTimeout = %v, expected 30s", config.HTTPTimeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, expected 3", config.MaxRetries)
	}
	if config.RetryDelayMs != 1000 {
		t.Errorf("RetryDelayMs = %d, expected 1000", config.RetryDelayMs)
	}

	// Test WebSocket config
	if config.MaxTokensPerConn != 50 {
		t.Errorf("MaxTokensPerConn = %d, expected 50", config.MaxTokensPerConn)
	}
	if config.ReconnectMinInterval != 1000 {
		t.Errorf("ReconnectMinInterval = %d, expected 1000", config.ReconnectMinInterval)
	}
	if config.ReconnectMaxInterval != 30000 {
		t.Errorf("ReconnectMaxInterval = %d, expected 30000", config.ReconnectMaxInterval)
	}
	if config.ReconnectMaxAttempts != 0 {
		t.Errorf("ReconnectMaxAttempts = %d, expected 0", config.ReconnectMaxAttempts)
	}
	if config.PingInterval != 30 {
		t.Errorf("PingInterval = %d, expected 30", config.PingInterval)
	}
	if config.PongTimeout != 10 {
		t.Errorf("PongTimeout = %d, expected 10", config.PongTimeout)
	}
	if config.MessageBufferSize != 1000 {
		t.Errorf("MessageBufferSize = %d, expected 1000", config.MessageBufferSize)
	}
	if config.UpdateChannelSize != 1000 {
		t.Errorf("UpdateChannelSize = %d, expected 1000", config.UpdateChannelSize)
	}

	// Test contract addresses
	if config.CTFExchangeAddress != CTFExchangeAddress {
		t.Errorf("CTFExchangeAddress mismatch")
	}
	if config.NegRiskCTFExchangeAddress != NegRiskCTFExchangeAddress {
		t.Errorf("NegRiskCTFExchangeAddress mismatch")
	}
	if config.NegRiskAdapterAddress != NegRiskAdapterAddress {
		t.Errorf("NegRiskAdapterAddress mismatch")
	}
	if config.CollateralAddress != CollateralAddress {
		t.Errorf("CollateralAddress mismatch")
	}
}

func TestConfigValidate(t *testing.T) {
	// Test empty config gets filled with defaults
	config := &Config{}
	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	if config.GammaEndpoint != GammaEndpoint {
		t.Error("Validate should set default GammaEndpoint")
	}
	if config.CLOBEndpoint != CLOBEndpoint {
		t.Error("Validate should set default CLOBEndpoint")
	}
	if config.WSEndpoint != WSEndpoint {
		t.Error("Validate should set default WSEndpoint")
	}
	if config.HTTPTimeout != 30*time.Second {
		t.Error("Validate should set default HTTPTimeout")
	}
	if config.MaxRetries != 3 {
		t.Error("Validate should set default MaxRetries")
	}
	if config.RetryDelayMs != 1000 {
		t.Error("Validate should set default RetryDelayMs")
	}
	if config.MaxTokensPerConn != 50 {
		t.Error("Validate should set default MaxTokensPerConn")
	}
	if config.CTFExchangeAddress != CTFExchangeAddress {
		t.Error("Validate should set default CTFExchangeAddress")
	}
}

func TestConfigValidatePreservesCustomValues(t *testing.T) {
	config := &Config{
		GammaEndpoint: "https://custom-gamma.example.com",
		CLOBEndpoint:  "https://custom-clob.example.com",
		WSEndpoint:    "wss://custom-ws.example.com",
		HTTPTimeout:   10 * time.Second,
		MaxRetries:    5,
		RetryDelayMs:  500,
	}

	config.Validate()

	if config.GammaEndpoint != "https://custom-gamma.example.com" {
		t.Error("Validate should preserve custom GammaEndpoint")
	}
	if config.CLOBEndpoint != "https://custom-clob.example.com" {
		t.Error("Validate should preserve custom CLOBEndpoint")
	}
	if config.WSEndpoint != "wss://custom-ws.example.com" {
		t.Error("Validate should preserve custom WSEndpoint")
	}
	if config.HTTPTimeout != 10*time.Second {
		t.Error("Validate should preserve custom HTTPTimeout")
	}
	if config.MaxRetries != 5 {
		t.Error("Validate should preserve custom MaxRetries")
	}
}

func TestConfigClone(t *testing.T) {
	original := DefaultConfig()
	clone := original.Clone()

	if clone == original {
		t.Error("Clone should return a new instance")
	}

	// Modify clone
	clone.GammaEndpoint = "https://modified.example.com"
	clone.HTTPTimeout = 60 * time.Second

	// Original should be unchanged
	if original.GammaEndpoint == "https://modified.example.com" {
		t.Error("Modifying clone should not affect original")
	}
	if original.HTTPTimeout == 60*time.Second {
		t.Error("Modifying clone should not affect original")
	}
}

func TestConfigClonePreservesValues(t *testing.T) {
	original := &Config{
		GammaEndpoint:             "https://custom.example.com",
		CLOBEndpoint:              "https://custom-clob.example.com",
		HTTPTimeout:               45 * time.Second,
		MaxRetries:                7,
		MaxTokensPerConn:          100,
		CTFExchangeAddress:        "0x1234567890123456789012345678901234567890",
		NegRiskCTFExchangeAddress: "0x1234567890123456789012345678901234567891",
	}

	clone := original.Clone()

	if clone.GammaEndpoint != original.GammaEndpoint {
		t.Error("Clone should preserve GammaEndpoint")
	}
	if clone.CLOBEndpoint != original.CLOBEndpoint {
		t.Error("Clone should preserve CLOBEndpoint")
	}
	if clone.HTTPTimeout != original.HTTPTimeout {
		t.Error("Clone should preserve HTTPTimeout")
	}
	if clone.MaxRetries != original.MaxRetries {
		t.Error("Clone should preserve MaxRetries")
	}
	if clone.MaxTokensPerConn != original.MaxTokensPerConn {
		t.Error("Clone should preserve MaxTokensPerConn")
	}
	if clone.CTFExchangeAddress != original.CTFExchangeAddress {
		t.Error("Clone should preserve CTFExchangeAddress")
	}
}

func TestContractAddressFormats(t *testing.T) {
	addresses := []string{
		CTFExchangeAddress,
		NegRiskCTFExchangeAddress,
		NegRiskAdapterAddress,
		CollateralAddress,
		ConditionalTokensAddress,
	}

	for _, addr := range addresses {
		if len(addr) != 42 {
			t.Errorf("Address %s should be 42 chars, got %d", addr, len(addr))
		}
		if addr[:2] != "0x" {
			t.Errorf("Address %s should start with 0x", addr)
		}
	}
}
