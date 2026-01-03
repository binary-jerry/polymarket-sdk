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

func setupAccountTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
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

func TestGetBalanceAllowance(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/balance-allowance" {
			t.Errorf("Expected path /balance-allowance, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		ba := BalanceAllowance{
			Balance:   decimal.NewFromInt(1000000000), // 1000 USDC
			Allowance: decimal.NewFromInt(500000000),  // 500 USDC
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ba)
	})
	defer server.Close()

	params := &BalanceAllowanceParams{
		AssetType: AssetTypeCollateral,
	}
	ba, err := client.GetBalanceAllowance(context.Background(), params)
	if err != nil {
		t.Fatalf("GetBalanceAllowance() error: %v", err)
	}
	if !ba.Balance.Equal(decimal.NewFromInt(1000000000)) {
		t.Errorf("Balance = %s, expected 1000000000", ba.Balance)
	}
}

func TestGetCollateralBalance(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("asset_type") != "COLLATERAL" {
			t.Errorf("Expected asset_type=COLLATERAL, got %s", r.URL.Query().Get("asset_type"))
		}

		ba := BalanceAllowance{
			Balance:   decimal.NewFromInt(500000000),
			Allowance: decimal.NewFromInt(500000000),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ba)
	})
	defer server.Close()

	ba, err := client.GetCollateralBalance(context.Background())
	if err != nil {
		t.Fatalf("GetCollateralBalance() error: %v", err)
	}
	if ba == nil {
		t.Fatal("GetCollateralBalance() returned nil")
	}
}

func TestGetConditionalBalance(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("asset_type") != "CONDITIONAL" {
			t.Errorf("Expected asset_type=CONDITIONAL, got %s", r.URL.Query().Get("asset_type"))
		}
		if r.URL.Query().Get("token_id") != "token-123" {
			t.Errorf("Expected token_id=token-123, got %s", r.URL.Query().Get("token_id"))
		}

		ba := BalanceAllowance{
			Balance:   decimal.NewFromInt(100000000),
			Allowance: decimal.NewFromInt(100000000),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ba)
	})
	defer server.Close()

	ba, err := client.GetConditionalBalance(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("GetConditionalBalance() error: %v", err)
	}
	if ba == nil {
		t.Fatal("GetConditionalBalance() returned nil")
	}
}

func TestGetConditionalBalanceEmptyTokenID(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty token ID")
	})
	defer server.Close()

	_, err := client.GetConditionalBalance(context.Background(), "")
	if err == nil {
		t.Error("GetConditionalBalance() should fail with empty token ID")
	}
}

func TestGetTickSize(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tick-size" {
			t.Errorf("Expected path /tick-size, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "token-123" {
			t.Errorf("Expected token_id=token-123, got %s", r.URL.Query().Get("token_id"))
		}

		ts := TickSize{
			TickSize: decimal.NewFromFloat(0.01),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ts)
	})
	defer server.Close()

	ts, err := client.GetTickSize(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("GetTickSize() error: %v", err)
	}
	if !ts.TickSize.Equal(decimal.NewFromFloat(0.01)) {
		t.Errorf("TickSize = %s, expected 0.01", ts.TickSize)
	}
}

func TestGetTickSizeEmptyTokenID(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty token ID")
	})
	defer server.Close()

	_, err := client.GetTickSize(context.Background(), "")
	if err == nil {
		t.Error("GetTickSize() should fail with empty token ID")
	}
}

func TestGetPrice(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/price" {
			t.Errorf("Expected path /price, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("token_id") != "token-123" {
			t.Errorf("Expected token_id=token-123, got %s", r.URL.Query().Get("token_id"))
		}

		price := PriceInfo{
			TokenID: "token-123",
			Price:   decimal.NewFromFloat(0.55),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(price)
	})
	defer server.Close()

	price, err := client.GetPrice(context.Background(), "token-123")
	if err != nil {
		t.Fatalf("GetPrice() error: %v", err)
	}
	if price.TokenID != "token-123" {
		t.Errorf("TokenID = %s, expected token-123", price.TokenID)
	}
	if !price.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Errorf("Price = %s, expected 0.55", price.Price)
	}
}

func TestGetPriceEmptyTokenID(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty token ID")
	})
	defer server.Close()

	_, err := client.GetPrice(context.Background(), "")
	if err == nil {
		t.Error("GetPrice() should fail with empty token ID")
	}
}

func TestGetPrices(t *testing.T) {
	callCount := 0
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		tokenID := r.URL.Query().Get("token_id")

		price := PriceInfo{
			TokenID: tokenID,
			Price:   decimal.NewFromFloat(0.5),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(price)
	})
	defer server.Close()

	tokenIDs := []string{"token-1", "token-2", "token-3"}
	prices, err := client.GetPrices(context.Background(), tokenIDs)
	if err != nil {
		t.Fatalf("GetPrices() error: %v", err)
	}
	if len(prices) != 3 {
		t.Errorf("Expected 3 prices, got %d", len(prices))
	}
	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestGetPricesEmpty(t *testing.T) {
	client, server := setupAccountTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty token IDs")
	})
	defer server.Close()

	prices, err := client.GetPrices(context.Background(), []string{})
	if err != nil {
		t.Fatalf("GetPrices() error: %v", err)
	}
	if prices != nil {
		t.Error("Prices should be nil for empty token IDs")
	}
}
