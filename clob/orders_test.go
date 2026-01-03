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

const ordersTestPrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
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

func TestGetOrder(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/order-123" {
			t.Errorf("Expected path /order/order-123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		order := Order{
			ID:           "order-123",
			Market:       "market-456",
			Side:         OrderSideBuy,
			Price:        decimal.NewFromFloat(0.55),
			OriginalSize: decimal.NewFromInt(100),
			SizeMatched:  decimal.NewFromInt(0),
			Status:       OrderStatusLive,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})
	defer server.Close()

	order, err := client.GetOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("GetOrder() error: %v", err)
	}
	if order.ID != "order-123" {
		t.Errorf("Order ID = %s, expected order-123", order.ID)
	}
}

func TestGetOrderEmptyID(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty ID")
	})
	defer server.Close()

	_, err := client.GetOrder(context.Background(), "")
	if err == nil {
		t.Error("GetOrder() should fail with empty order ID")
	}
}

func TestGetOrders(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Errorf("Expected path /orders, got %s", r.URL.Path)
		}

		orders := []*Order{
			{ID: "order-1", Status: OrderStatusLive},
			{ID: "order-2", Status: OrderStatusLive},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	})
	defer server.Close()

	orders, err := client.GetOrders(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}
}

func TestGetOrdersWithParams(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("market") != "market-123" {
			t.Errorf("Expected market=market-123, got %s", r.URL.Query().Get("market"))
		}
		if r.URL.Query().Get("status") != "LIVE" {
			t.Errorf("Expected status=LIVE, got %s", r.URL.Query().Get("status"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*Order{})
	})
	defer server.Close()

	params := &OrdersQueryParams{
		Market: "market-123",
		Status: "LIVE",
	}
	_, err := client.GetOrders(context.Background(), params)
	if err != nil {
		t.Fatalf("GetOrders() error: %v", err)
	}
}

func TestGetOpenOrders(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Errorf("Expected path /orders, got %s", r.URL.Path)
		}

		orders := []*Order{
			{ID: "order-1", Status: OrderStatusLive},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	})
	defer server.Close()

	orders, err := client.GetOpenOrders(context.Background())
	if err != nil {
		t.Fatalf("GetOpenOrders() error: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders))
	}
}

func TestCancelOrder(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/order-123" {
			t.Errorf("Expected path /order/order-123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.CancelOrder(context.Background(), "order-123")
	if err != nil {
		t.Fatalf("CancelOrder() error: %v", err)
	}
}

func TestCancelOrderEmptyID(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty ID")
	})
	defer server.Close()

	err := client.CancelOrder(context.Background(), "")
	if err == nil {
		t.Error("CancelOrder() should fail with empty order ID")
	}
}

func TestCancelOrders(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Errorf("Expected path /orders, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		var body BatchCancelRequest
		json.NewDecoder(r.Body).Decode(&body)
		if len(body.OrderIDs) != 2 {
			t.Errorf("Expected 2 order IDs, got %d", len(body.OrderIDs))
		}

		resp := CancelResponse{
			Canceled:    []string{"order-1", "order-2"},
			NotCanceled: []string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CancelOrders(context.Background(), []string{"order-1", "order-2"})
	if err != nil {
		t.Fatalf("CancelOrders() error: %v", err)
	}
	if len(resp.Canceled) != 2 {
		t.Errorf("Expected 2 canceled orders, got %d", len(resp.Canceled))
	}
}

func TestCancelOrdersEmpty(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty order IDs")
	})
	defer server.Close()

	resp, err := client.CancelOrders(context.Background(), []string{})
	if err != nil {
		t.Fatalf("CancelOrders() error: %v", err)
	}
	if resp != nil {
		t.Error("Response should be nil for empty order IDs")
	}
}

func TestCancelOrdersByMarket(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		var body BatchCancelRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Market != "market-123" {
			t.Errorf("Expected market=market-123, got %s", body.Market)
		}

		resp := CancelResponse{Canceled: []string{"order-1"}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CancelOrdersByMarket(context.Background(), "market-123")
	if err != nil {
		t.Fatalf("CancelOrdersByMarket() error: %v", err)
	}
	if len(resp.Canceled) != 1 {
		t.Errorf("Expected 1 canceled order, got %d", len(resp.Canceled))
	}
}

func TestCancelOrdersByMarketEmpty(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty market ID")
	})
	defer server.Close()

	_, err := client.CancelOrdersByMarket(context.Background(), "")
	if err == nil {
		t.Error("CancelOrdersByMarket() should fail with empty market ID")
	}
}

func TestCancelOrdersByAsset(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body BatchCancelRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.AssetID != "asset-123" {
			t.Errorf("Expected asset_id=asset-123, got %s", body.AssetID)
		}

		resp := CancelResponse{Canceled: []string{"order-1"}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	resp, err := client.CancelOrdersByAsset(context.Background(), "asset-123")
	if err != nil {
		t.Fatalf("CancelOrdersByAsset() error: %v", err)
	}
	if len(resp.Canceled) != 1 {
		t.Errorf("Expected 1 canceled order, got %d", len(resp.Canceled))
	}
}

func TestCancelOrdersByAssetEmpty(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty asset ID")
	})
	defer server.Close()

	_, err := client.CancelOrdersByAsset(context.Background(), "")
	if err == nil {
		t.Error("CancelOrdersByAsset() should fail with empty asset ID")
	}
}

func TestCancelAllOrders(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cancel-all" {
			t.Errorf("Expected path /cancel-all, got %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	err := client.CancelAllOrders(context.Background())
	if err != nil {
		t.Fatalf("CancelAllOrders() error: %v", err)
	}
}

func TestCreateOrdersEmpty(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with empty orders")
	})
	defer server.Close()

	resp, err := client.CreateOrders(context.Background(), []*CreateOrderRequest{})
	if err != nil {
		t.Fatalf("CreateOrders() error: %v", err)
	}
	if resp != nil {
		t.Error("Response should be nil for empty orders")
	}
}

func TestCreateOrdersTooMany(t *testing.T) {
	client, server := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("Request should not be made with too many orders")
	})
	defer server.Close()

	reqs := make([]*CreateOrderRequest, 20)
	for i := range reqs {
		reqs[i] = &CreateOrderRequest{
			TokenID: "12345",
			Side:    OrderSideBuy,
			Price:   decimal.NewFromFloat(0.5),
			Size:    decimal.NewFromInt(10),
		}
	}

	_, err := client.CreateOrders(context.Background(), reqs)
	if err == nil {
		t.Error("CreateOrders() should fail with more than 15 orders")
	}
}
