package clob

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/polymarket-sdk/auth"
)

func setupTradesTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)

	config := &Config{
		Endpoint:               server.URL,
		ChainID:                137,
		Timeout:                5 * time.Second,
		MaxRetries:             0,
		ExchangeAddress:        "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		NegRiskExchangeAddress: "0xC5d563A36AE78145C45a50134d48A1215220f80a",
		NegRiskAdapterAddress:  "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	}

	creds := &auth.Credentials{
		APIKey:     "test-api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-passphrase",
	}

	client, err := NewClientWithCredentials(config, ordersTestPrivKey, creds)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client, server
}

func TestGetTrades(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trades" {
			t.Errorf("Expected path /trades, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		trades := []*Trade{
			{
				ID:        "trade-1",
				Market:    "market-123",
				AssetID:   "asset-456",
				Side:      OrderSideBuy,
				Price:     decimal.NewFromFloat(0.55),
				Size:      decimal.NewFromInt(100),
				Timestamp: time.Now(),
			},
			{
				ID:        "trade-2",
				Market:    "market-123",
				AssetID:   "asset-456",
				Side:      OrderSideSell,
				Price:     decimal.NewFromFloat(0.45),
				Size:      decimal.NewFromInt(50),
				Timestamp: time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	})
	defer server.Close()

	trades, err := client.GetTrades(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}
}

func TestGetTradesWithParams(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("market") != "market-123" {
			t.Errorf("Expected market=market-123, got %s", r.URL.Query().Get("market"))
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("Expected limit=50, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Trade{})
	})
	defer server.Close()

	params := &TradesQueryParams{
		Market: "market-123",
		Limit:  50,
	}
	_, err := client.GetTrades(context.Background(), params)
	if err != nil {
		t.Fatalf("GetTrades() error: %v", err)
	}
}

func TestGetTradesByMarket(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("market") != "market-123" {
			t.Errorf("Expected market=market-123, got %s", r.URL.Query().Get("market"))
		}

		trades := []*Trade{
			{ID: "trade-1", Market: "market-123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	})
	defer server.Close()

	trades, err := client.GetTradesByMarket(context.Background(), "market-123", 50)
	if err != nil {
		t.Fatalf("GetTradesByMarket() error: %v", err)
	}
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
}

func TestGetTradesByMarketEmptyID(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty market ID")
	})
	defer server.Close()

	_, err := client.GetTradesByMarket(context.Background(), "", 50)
	if err == nil {
		t.Error("GetTradesByMarket() should fail with empty market ID")
	}
}

func TestGetTradesByMarketDefaultLimit(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "100" {
			t.Errorf("Expected default limit=100, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Trade{})
	})
	defer server.Close()

	_, err := client.GetTradesByMarket(context.Background(), "market-123", 0)
	if err != nil {
		t.Fatalf("GetTradesByMarket() error: %v", err)
	}
}

func TestGetTradesByAsset(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("asset_id") != "asset-123" {
			t.Errorf("Expected asset_id=asset-123, got %s", r.URL.Query().Get("asset_id"))
		}

		trades := []*Trade{
			{ID: "trade-1", AssetID: "asset-123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	})
	defer server.Close()

	trades, err := client.GetTradesByAsset(context.Background(), "asset-123", 50)
	if err != nil {
		t.Fatalf("GetTradesByAsset() error: %v", err)
	}
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
}

func TestGetTradesByAssetEmptyID(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty asset ID")
	})
	defer server.Close()

	_, err := client.GetTradesByAsset(context.Background(), "", 50)
	if err == nil {
		t.Error("GetTradesByAsset() should fail with empty asset ID")
	}
}

func TestGetRecentTrades(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "25" {
			t.Errorf("Expected limit=25, got %s", r.URL.Query().Get("limit"))
		}

		trades := []*Trade{
			{ID: "trade-1"},
			{ID: "trade-2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	})
	defer server.Close()

	trades, err := client.GetRecentTrades(context.Background(), 25)
	if err != nil {
		t.Fatalf("GetRecentTrades() error: %v", err)
	}
	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}
}

func TestGetRecentTradesDefaultLimit(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("Expected default limit=50, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Trade{})
	})
	defer server.Close()

	_, err := client.GetRecentTrades(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetRecentTrades() error: %v", err)
	}
}

func TestGetTradesByTimeRange(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("after") != "2024-01-01" {
			t.Errorf("Expected after=2024-01-01, got %s", r.URL.Query().Get("after"))
		}
		if r.URL.Query().Get("before") != "2024-12-31" {
			t.Errorf("Expected before=2024-12-31, got %s", r.URL.Query().Get("before"))
		}

		trades := []*Trade{
			{ID: "trade-1"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	})
	defer server.Close()

	trades, err := client.GetTradesByTimeRange(context.Background(), "2024-01-01", "2024-12-31", 100)
	if err != nil {
		t.Fatalf("GetTradesByTimeRange() error: %v", err)
	}
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
}

func TestGetTradesByTimeRangeDefaultLimit(t *testing.T) {
	client, server := setupTradesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "100" {
			t.Errorf("Expected default limit=100, got %s", r.URL.Query().Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Trade{})
	})
	defer server.Close()

	_, err := client.GetTradesByTimeRange(context.Background(), "", "", 0)
	if err != nil {
		t.Fatalf("GetTradesByTimeRange() error: %v", err)
	}
}
