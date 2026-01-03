package polymarket

import (
	"encoding/base64"
	"testing"

	"github.com/binary-jerry/polymarket-sdk/auth"
)

// 测试用私钥（请勿在生产环境使用）
const sdkTestPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewSDK(t *testing.T) {
	// Test with nil config
	sdk, err := NewSDK(nil, sdkTestPrivateKey)
	if err != nil {
		t.Fatalf("NewSDK() error: %v", err)
	}
	if sdk == nil {
		t.Fatal("NewSDK() returned nil")
	}
	defer sdk.Close()

	// Verify all components are initialized
	if sdk.OrderBook == nil {
		t.Error("OrderBook should not be nil")
	}
	if sdk.Markets == nil {
		t.Error("Markets should not be nil")
	}
	if sdk.Trading == nil {
		t.Error("Trading should not be nil")
	}
}

func TestNewSDKWithConfig(t *testing.T) {
	config := &Config{
		GammaEndpoint: "https://custom-gamma.example.com",
		CLOBEndpoint:  "https://custom-clob.example.com",
	}

	sdk, err := NewSDK(config, sdkTestPrivateKey)
	if err != nil {
		t.Fatalf("NewSDK() error: %v", err)
	}
	defer sdk.Close()

	if sdk.GetConfig().GammaEndpoint != "https://custom-gamma.example.com" {
		t.Error("Config should use custom GammaEndpoint")
	}
}

func TestNewSDKInvalidPrivateKey(t *testing.T) {
	_, err := NewSDK(nil, "invalid-key")
	if err == nil {
		t.Error("NewSDK() should fail with invalid private key")
	}
}

func TestNewPublicSDK(t *testing.T) {
	sdk := NewPublicSDK(nil)
	if sdk == nil {
		t.Fatal("NewPublicSDK() returned nil")
	}
	defer sdk.Close()

	// Verify public components are initialized
	if sdk.OrderBook == nil {
		t.Error("OrderBook should not be nil")
	}
	if sdk.Markets == nil {
		t.Error("Markets should not be nil")
	}

	// Trading should be nil for public SDK
	if sdk.Trading != nil {
		t.Error("Trading should be nil for public SDK")
	}
}

func TestNewPublicSDKWithConfig(t *testing.T) {
	config := &Config{
		GammaEndpoint: "https://custom-gamma.example.com",
	}

	sdk := NewPublicSDK(config)
	if sdk == nil {
		t.Fatal("NewPublicSDK() returned nil")
	}
	defer sdk.Close()

	if sdk.GetConfig().GammaEndpoint != "https://custom-gamma.example.com" {
		t.Error("Config should use custom GammaEndpoint")
	}
}

func TestNewTradingSDK(t *testing.T) {
	sdk, err := NewTradingSDK(nil, sdkTestPrivateKey, nil)
	if err != nil {
		t.Fatalf("NewTradingSDK() error: %v", err)
	}
	defer sdk.Close()

	if sdk.Trading == nil {
		t.Error("Trading should not be nil")
	}
}

func TestNewTradingSDKWithCredentials(t *testing.T) {
	creds := &auth.Credentials{
		APIKey:     "test-api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-passphrase",
	}

	sdk, err := NewTradingSDK(nil, sdkTestPrivateKey, creds)
	if err != nil {
		t.Fatalf("NewTradingSDK() error: %v", err)
	}
	defer sdk.Close()

	if sdk.GetCredentials() != creds {
		t.Error("GetCredentials() should return the provided credentials")
	}
}

func TestNewTradingSDKInvalidPrivateKey(t *testing.T) {
	_, err := NewTradingSDK(nil, "invalid-key", nil)
	if err == nil {
		t.Error("NewTradingSDK() should fail with invalid private key")
	}
}

func TestSDKClose(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	// Close should not panic
	sdk.Close()
}

func TestPublicSDKClose(t *testing.T) {
	sdk := NewPublicSDK(nil)
	// Close should not panic
	sdk.Close()
}

func TestSDKGetAddress(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	defer sdk.Close()

	addr := sdk.GetAddress()
	if addr == "" {
		t.Error("GetAddress() should not return empty string")
	}
	if len(addr) != 42 {
		t.Errorf("Address should be 42 chars, got %d", len(addr))
	}
}

func TestPublicSDKGetAddress(t *testing.T) {
	sdk := NewPublicSDK(nil)
	defer sdk.Close()

	addr := sdk.GetAddress()
	if addr != "" {
		t.Error("GetAddress() should return empty string for public SDK")
	}
}

func TestSDKGetConfig(t *testing.T) {
	config := DefaultConfig()
	sdk, _ := NewSDK(config, sdkTestPrivateKey)
	defer sdk.Close()

	if sdk.GetConfig() != config {
		t.Error("GetConfig() should return the same config")
	}
}

func TestSDKGetCredentials(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	defer sdk.Close()

	// Initially nil
	if sdk.GetCredentials() != nil {
		t.Error("GetCredentials() should return nil initially")
	}
}

func TestPublicSDKGetCredentials(t *testing.T) {
	sdk := NewPublicSDK(nil)
	defer sdk.Close()

	// Should return nil for public SDK
	if sdk.GetCredentials() != nil {
		t.Error("GetCredentials() should return nil for public SDK")
	}
}

func TestSDKSetCredentials(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	defer sdk.Close()

	creds := &auth.Credentials{
		APIKey:     "test-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-pass",
	}

	sdk.SetCredentials(creds)

	if sdk.GetCredentials() != creds {
		t.Error("GetCredentials() should return the set credentials")
	}
}

func TestPublicSDKSetCredentials(t *testing.T) {
	sdk := NewPublicSDK(nil)
	defer sdk.Close()

	creds := &auth.Credentials{
		APIKey:     "test-key",
		Secret:     "test-secret",
		Passphrase: "test-pass",
	}

	// Should not panic
	sdk.SetCredentials(creds)
}

func TestSDKIsTradingEnabled(t *testing.T) {
	// Full SDK should have trading enabled
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	defer sdk.Close()

	if !sdk.IsTradingEnabled() {
		t.Error("IsTradingEnabled() should return true for full SDK")
	}
}

func TestPublicSDKIsTradingEnabled(t *testing.T) {
	// Public SDK should not have trading enabled
	sdk := NewPublicSDK(nil)
	defer sdk.Close()

	if sdk.IsTradingEnabled() {
		t.Error("IsTradingEnabled() should return false for public SDK")
	}
}

func TestSDKCreateOrDeriveAPICredentialsWithoutTrading(t *testing.T) {
	sdk := NewPublicSDK(nil)
	defer sdk.Close()

	_, err := sdk.CreateOrDeriveAPICredentials(nil)
	if err == nil {
		t.Error("CreateOrDeriveAPICredentials() should fail for public SDK")
	}
}

func TestSDKComponentsIndependence(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	defer sdk.Close()

	// Each component should be accessible independently
	_ = sdk.OrderBook
	_ = sdk.Markets
	_ = sdk.Trading
}

func TestSDKConfigAppliedToComponents(t *testing.T) {
	config := &Config{
		GammaEndpoint: "https://custom-gamma.example.com",
		CLOBEndpoint:  "https://custom-clob.example.com",
	}

	sdk, _ := NewSDK(config, sdkTestPrivateKey)
	defer sdk.Close()

	// Verify config is propagated
	if sdk.Markets.GetConfig().Endpoint != "https://custom-gamma.example.com" {
		t.Error("Markets should use custom GammaEndpoint")
	}
	if sdk.Trading.GetConfig().Endpoint != "https://custom-clob.example.com" {
		t.Error("Trading should use custom CLOBEndpoint")
	}
}

func TestSDKMultipleClose(t *testing.T) {
	sdk, _ := NewSDK(nil, sdkTestPrivateKey)
	// Multiple closes should not panic
	sdk.Close()
	sdk.Close()
	sdk.Close()
}

func TestSDKNilComponents(t *testing.T) {
	sdk := &SDK{}

	// Operations on nil components should not panic
	sdk.Close()
	addr := sdk.GetAddress()
	if addr != "" {
		t.Error("GetAddress() should return empty for nil l1Signer")
	}
	creds := sdk.GetCredentials()
	if creds != nil {
		t.Error("GetCredentials() should return nil for nil Trading")
	}
	sdk.SetCredentials(&auth.Credentials{})
	if !sdk.IsTradingEnabled() == true {
		// This is expected to be false
	}
}
