package clob

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/binary-jerry/polymarket-sdk/auth"
)

const testPrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Endpoint != "https://clob.polymarket.com" {
		t.Errorf("Endpoint = %s, expected https://clob.polymarket.com", config.Endpoint)
	}
	if config.ChainID != 137 {
		t.Errorf("ChainID = %d, expected 137", config.ChainID)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, expected 30s", config.Timeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, expected 3", config.MaxRetries)
	}
	if config.ExchangeAddress != "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e" {
		t.Errorf("ExchangeAddress = %s, expected 0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e", config.ExchangeAddress)
	}
	if config.NegRiskExchangeAddress != "0xC5d563A36AE78145C45a50134d48A1215220f80a" {
		t.Errorf("NegRiskExchangeAddress = %s, expected 0xC5d563A36AE78145C45a50134d48A1215220f80a", config.NegRiskExchangeAddress)
	}
	if config.NegRiskAdapterAddress != "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296" {
		t.Errorf("NegRiskAdapterAddress = %s, expected 0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296", config.NegRiskAdapterAddress)
	}
}

func TestNewClient(t *testing.T) {
	// Test with nil config
	client, err := NewClient(nil, testPrivKey)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	// Test with custom config
	config := &Config{
		Endpoint:               "https://custom.example.com",
		ChainID:                1,
		Timeout:                10 * time.Second,
		ExchangeAddress:        "0x1234567890123456789012345678901234567890",
		NegRiskExchangeAddress: "0x1234567890123456789012345678901234567891",
		NegRiskAdapterAddress:  "0x1234567890123456789012345678901234567892",
	}
	client, err = NewClient(config, testPrivKey)
	if err != nil {
		t.Fatalf("NewClient() with config error: %v", err)
	}
	if client.GetConfig().Endpoint != "https://custom.example.com" {
		t.Errorf("Config endpoint = %s, expected https://custom.example.com", client.GetConfig().Endpoint)
	}
}

func TestNewClientInvalidPrivateKey(t *testing.T) {
	_, err := NewClient(nil, "invalid-key")
	if err == nil {
		t.Error("NewClient() should fail with invalid private key")
	}
}

func TestNewClientWithCredentials(t *testing.T) {
	creds := &auth.Credentials{
		APIKey:     "test-api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-passphrase",
	}

	client, err := NewClientWithCredentials(nil, testPrivKey, creds)
	if err != nil {
		t.Fatalf("NewClientWithCredentials() error: %v", err)
	}

	if client.GetCredentials() != creds {
		t.Error("GetCredentials() should return the provided credentials")
	}
}

func TestClientClose(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)
	// Close should not panic
	client.Close()
}

func TestClientGetAddress(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)
	addr := client.GetAddress()

	if addr == "" {
		t.Error("GetAddress() should not return empty string")
	}
	if len(addr) != 42 {
		t.Errorf("Address should be 42 chars, got %d", len(addr))
	}
}

func TestClientGetCredentials(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	// Initially nil
	if client.GetCredentials() != nil {
		t.Error("GetCredentials() should return nil initially")
	}
}

func TestClientSetCredentials(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	creds := &auth.Credentials{
		APIKey:     "api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("secret")),
		Passphrase: "passphrase",
	}

	client.SetCredentials(creds)

	if client.GetCredentials() != creds {
		t.Error("GetCredentials() should return the set credentials")
	}
}

func TestClientGetL1Signer(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	signer := client.GetL1Signer()
	if signer == nil {
		t.Error("GetL1Signer() should not return nil")
	}
}

func TestClientGetOrderSigner(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	orderSigner := client.GetOrderSigner()
	if orderSigner == nil {
		t.Error("GetOrderSigner() should not return nil")
	}
}

func TestClientGetConfig(t *testing.T) {
	config := &Config{
		Endpoint: "https://test.example.com",
		ChainID:  137,
	}
	client, _ := NewClient(config, testPrivKey)

	returnedConfig := client.GetConfig()
	if returnedConfig.Endpoint != "https://test.example.com" {
		t.Error("GetConfig() should return the correct endpoint")
	}
}

func TestClientGetL2AuthHeadersWithoutCredentials(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	_, err := client.getL2AuthHeaders("GET", "/orders", "")
	if err == nil {
		t.Error("getL2AuthHeaders() should fail without credentials")
	}
}

func TestClientGetL2AuthHeadersWithCredentials(t *testing.T) {
	creds := &auth.Credentials{
		APIKey:     "test-api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-passphrase",
	}

	client, _ := NewClientWithCredentials(nil, testPrivKey, creds)

	headers, err := client.getL2AuthHeaders("GET", "/orders", "")
	if err != nil {
		t.Fatalf("getL2AuthHeaders() error: %v", err)
	}

	if headers["POLY_ADDRESS"] == "" {
		t.Error("POLY_ADDRESS should not be empty")
	}
	if headers["POLY_API_KEY"] != creds.APIKey {
		t.Error("POLY_API_KEY should match credentials")
	}
	if headers["POLY_PASSPHRASE"] != creds.Passphrase {
		t.Error("POLY_PASSPHRASE should match credentials")
	}
	if headers["POLY_TIMESTAMP"] == "" {
		t.Error("POLY_TIMESTAMP should not be empty")
	}
	if headers["POLY_SIGNATURE"] == "" {
		t.Error("POLY_SIGNATURE should not be empty")
	}
}

func TestClientConcurrentAccess(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = client.GetAddress()
			_ = client.GetCredentials()
			_ = client.GetConfig()
			_ = client.GetL1Signer()
			_ = client.GetOrderSigner()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClientSetCredentialsConcurrent(t *testing.T) {
	client, _ := NewClient(nil, testPrivKey)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			creds := &auth.Credentials{
				APIKey:     "api-key-" + string(rune('0'+idx)),
				Secret:     base64.StdEncoding.EncodeToString([]byte("secret")),
				Passphrase: "passphrase",
			}
			client.SetCredentials(creds)
			_ = client.GetCredentials()
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConfigContractAddresses(t *testing.T) {
	config := DefaultConfig()

	// Verify contract addresses are valid Ethereum addresses
	addresses := []string{
		config.ExchangeAddress,
		config.NegRiskExchangeAddress,
		config.NegRiskAdapterAddress,
		config.CollateralAddress,
	}

	for _, addr := range addresses {
		if len(addr) != 42 {
			t.Errorf("Address %s should be 42 chars", addr)
		}
		if addr[:2] != "0x" {
			t.Errorf("Address %s should start with 0x", addr)
		}
	}
}
