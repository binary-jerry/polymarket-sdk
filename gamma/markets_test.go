package gamma

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(&Config{
		Endpoint:     server.URL,
		Timeout:      5 * time.Second,
		MaxRetries:   0,
		RetryDelayMs: 100,
	})
	return server, client
}

func TestGetMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets" {
			t.Errorf("Expected path /markets, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		markets := []Market{
			{ID: "1", Question: "Test Market 1"},
			{ID: "2", Question: "Test Market 2"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(markets)
	})
	defer server.Close()

	resp, err := client.GetMarkets(context.Background(), nil)
	if err != nil {
		t.Errorf("GetMarkets() error: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 markets, got %d", len(resp.Data))
	}
}

func TestGetMarketsWithParams(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", r.URL.Query().Get("active"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	params := &MarketListParams{
		Limit:  10,
		Active: BoolPtr(true),
	}
	_, err := client.GetMarkets(context.Background(), params)
	if err != nil {
		t.Errorf("GetMarkets() error: %v", err)
	}
}

func TestGetMarket(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/123" {
			t.Errorf("Expected path /markets/123, got %s", r.URL.Path)
		}

		market := Market{ID: "123", Question: "Test Market"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(market)
	})
	defer server.Close()

	market, err := client.GetMarket(context.Background(), "123")
	if err != nil {
		t.Errorf("GetMarket() error: %v", err)
	}
	if market.ID != "123" {
		t.Errorf("Market ID = %s, expected 123", market.ID)
	}
}

func TestGetMarketEmptyID(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMarket(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty market ID")
	}
}

func TestGetMarketBySlug(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/slug/test-market" {
			t.Errorf("Expected path /markets/slug/test-market, got %s", r.URL.Path)
		}

		market := Market{ID: "123", Slug: "test-market"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(market)
	})
	defer server.Close()

	market, err := client.GetMarketBySlug(context.Background(), "test-market")
	if err != nil {
		t.Errorf("GetMarketBySlug() error: %v", err)
	}
	if market.Slug != "test-market" {
		t.Errorf("Market Slug = %s, expected test-market", market.Slug)
	}
}

func TestGetMarketBySlugEmpty(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMarketBySlug(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty slug")
	}
}

func TestGetActiveMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("active") != "true" {
			t.Errorf("Expected active=true")
		}
		if r.URL.Query().Get("closed") != "false" {
			t.Errorf("Expected closed=false")
		}

		markets := []Market{
			{ID: "1", Active: true},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(markets)
	})
	defer server.Close()

	markets, err := client.GetActiveMarkets(context.Background(), 10)
	if err != nil {
		t.Errorf("GetActiveMarkets() error: %v", err)
	}
	if len(markets) != 1 {
		t.Errorf("Expected 1 market, got %d", len(markets))
	}
}

func TestGetFeaturedMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("featured") != "true" {
			t.Errorf("Expected featured=true")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetFeaturedMarkets(context.Background(), 5)
	if err != nil {
		t.Errorf("GetFeaturedMarkets() error: %v", err)
	}
}

func TestGetNegRiskMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("neg_risk") != "true" {
			t.Errorf("Expected neg_risk=true")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetNegRiskMarkets(context.Background(), 10)
	if err != nil {
		t.Errorf("GetNegRiskMarkets() error: %v", err)
	}
}

func TestSearchMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("text_query") != "bitcoin" {
			t.Errorf("Expected text_query=bitcoin, got %s", r.URL.Query().Get("text_query"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.SearchMarkets(context.Background(), "bitcoin", 10)
	if err != nil {
		t.Errorf("SearchMarkets() error: %v", err)
	}
}

func TestSearchMarketsEmptyQuery(t *testing.T) {
	client := NewClient(nil)
	_, err := client.SearchMarkets(context.Background(), "", 10)
	if err == nil {
		t.Error("Expected error for empty query")
	}
}

func TestGetMarketsByCategory(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("category") != "politics" {
			t.Errorf("Expected category=politics")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetMarketsByCategory(context.Background(), "politics", 10)
	if err != nil {
		t.Errorf("GetMarketsByCategory() error: %v", err)
	}
}

func TestGetMarketsByCategoryEmpty(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMarketsByCategory(context.Background(), "", 10)
	if err == nil {
		t.Error("Expected error for empty category")
	}
}

func TestGetMarketsByTag(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("tag_slug") != "crypto" {
			t.Errorf("Expected tag_slug=crypto")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetMarketsByTag(context.Background(), "crypto", 10)
	if err != nil {
		t.Errorf("GetMarketsByTag() error: %v", err)
	}
}

func TestGetMarketsByTagEmpty(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMarketsByTag(context.Background(), "", 10)
	if err == nil {
		t.Error("Expected error for empty tag slug")
	}
}

func TestGetTopVolumeMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("order") != "volume" {
			t.Errorf("Expected order=volume")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetTopVolumeMarkets(context.Background(), 10)
	if err != nil {
		t.Errorf("GetTopVolumeMarkets() error: %v", err)
	}
}

func TestGetEndingSoonMarkets(t *testing.T) {
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("order") != "end_date_min" {
			t.Errorf("Expected order=end_date_min")
		}
		if r.URL.Query().Get("ascending") != "true" {
			t.Errorf("Expected ascending=true")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Market{})
	})
	defer server.Close()

	_, err := client.GetEndingSoonMarkets(context.Background(), 10)
	if err != nil {
		t.Errorf("GetEndingSoonMarkets() error: %v", err)
	}
}

func TestGetAllMarketsLimit(t *testing.T) {
	callCount := 0
	server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return empty on second call to stop pagination
		var markets []Market
		if callCount == 1 {
			for i := 0; i < 100; i++ {
				markets = append(markets, Market{ID: string(rune(i))})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(markets)
	})
	defer server.Close()

	markets, err := client.GetAllMarkets(context.Background(), nil)
	if err != nil {
		t.Errorf("GetAllMarkets() error: %v", err)
	}
	if len(markets) != 100 {
		t.Errorf("Expected 100 markets, got %d", len(markets))
	}
}
