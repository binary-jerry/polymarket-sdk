package gamma

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Endpoint != "https://gamma-api.polymarket.com" {
		t.Errorf("Endpoint = %s, expected https://gamma-api.polymarket.com", config.Endpoint)
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, expected 30s", config.Timeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, expected 3", config.MaxRetries)
	}
	if config.RetryDelayMs != 1000 {
		t.Errorf("RetryDelayMs = %d, expected 1000", config.RetryDelayMs)
	}
}

func TestNewClient(t *testing.T) {
	// Test with nil config
	client := NewClient(nil)
	if client == nil {
		t.Error("NewClient(nil) should not return nil")
	}
	if client.config == nil {
		t.Error("Client config should not be nil")
	}
	if client.httpClient == nil {
		t.Error("Client httpClient should not be nil")
	}

	// Test with custom config
	config := &Config{
		Endpoint:     "https://custom.example.com",
		Timeout:      10 * time.Second,
		MaxRetries:   5,
		RetryDelayMs: 500,
	}
	client = NewClient(config)
	if client.config.Endpoint != config.Endpoint {
		t.Errorf("Config endpoint = %s, expected %s", client.config.Endpoint, config.Endpoint)
	}
}

func TestClientClose(t *testing.T) {
	client := NewClient(nil)
	// Close should not panic
	client.Close()
}

func TestClientGetConfig(t *testing.T) {
	config := &Config{
		Endpoint: "https://test.example.com",
	}
	client := NewClient(config)

	returnedConfig := client.GetConfig()
	if returnedConfig != config {
		t.Error("GetConfig should return the same config")
	}
}
