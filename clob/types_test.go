package clob

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestOrderTypeConstants(t *testing.T) {
	if OrderTypeGTC != "GTC" {
		t.Errorf("OrderTypeGTC = %s, expected GTC", OrderTypeGTC)
	}
	if OrderTypeGTD != "GTD" {
		t.Errorf("OrderTypeGTD = %s, expected GTD", OrderTypeGTD)
	}
	if OrderTypeFOK != "FOK" {
		t.Errorf("OrderTypeFOK = %s, expected FOK", OrderTypeFOK)
	}
	if OrderTypeFAK != "FAK" {
		t.Errorf("OrderTypeFAK = %s, expected FAK", OrderTypeFAK)
	}
}

func TestOrderSideConstants(t *testing.T) {
	if OrderSideBuy != "BUY" {
		t.Errorf("OrderSideBuy = %s, expected BUY", OrderSideBuy)
	}
	if OrderSideSell != "SELL" {
		t.Errorf("OrderSideSell = %s, expected SELL", OrderSideSell)
	}
}

func TestOrderSideToInt(t *testing.T) {
	if OrderSideBuy.ToInt() != 0 {
		t.Errorf("OrderSideBuy.ToInt() = %d, expected 0", OrderSideBuy.ToInt())
	}
	if OrderSideSell.ToInt() != 1 {
		t.Errorf("OrderSideSell.ToInt() = %d, expected 1", OrderSideSell.ToInt())
	}
}

func TestOrderStatusConstants(t *testing.T) {
	if OrderStatusLive != "LIVE" {
		t.Errorf("OrderStatusLive = %s, expected LIVE", OrderStatusLive)
	}
	if OrderStatusMatched != "MATCHED" {
		t.Errorf("OrderStatusMatched = %s, expected MATCHED", OrderStatusMatched)
	}
	if OrderStatusCanceled != "CANCELED" {
		t.Errorf("OrderStatusCanceled = %s, expected CANCELED", OrderStatusCanceled)
	}
	if OrderStatusDelayed != "DELAYED" {
		t.Errorf("OrderStatusDelayed = %s, expected DELAYED", OrderStatusDelayed)
	}
}

func TestOrderGetRemainingSize(t *testing.T) {
	tests := []struct {
		name         string
		originalSize decimal.Decimal
		sizeMatched  decimal.Decimal
		expected     decimal.Decimal
	}{
		{
			name:         "no matches",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(0),
			expected:     decimal.NewFromInt(100),
		},
		{
			name:         "partial fill",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(30),
			expected:     decimal.NewFromInt(70),
		},
		{
			name:         "full fill",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(100),
			expected:     decimal.NewFromInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{
				OriginalSize: tt.originalSize,
				SizeMatched:  tt.sizeMatched,
			}
			if !order.GetRemainingSize().Equal(tt.expected) {
				t.Errorf("GetRemainingSize() = %s, expected %s", order.GetRemainingSize(), tt.expected)
			}
		})
	}
}

func TestOrderIsFilled(t *testing.T) {
	tests := []struct {
		name         string
		originalSize decimal.Decimal
		sizeMatched  decimal.Decimal
		expected     bool
	}{
		{
			name:         "not filled",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(50),
			expected:     false,
		},
		{
			name:         "exactly filled",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(100),
			expected:     true,
		},
		{
			name:         "over filled",
			originalSize: decimal.NewFromInt(100),
			sizeMatched:  decimal.NewFromInt(110),
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{
				OriginalSize: tt.originalSize,
				SizeMatched:  tt.sizeMatched,
			}
			if order.IsFilled() != tt.expected {
				t.Errorf("IsFilled() = %v, expected %v", order.IsFilled(), tt.expected)
			}
		})
	}
}

func TestOrderIsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{
			name:     "live order",
			status:   OrderStatusLive,
			expected: true,
		},
		{
			name:     "delayed order",
			status:   OrderStatusDelayed,
			expected: true,
		},
		{
			name:     "matched order",
			status:   OrderStatusMatched,
			expected: false,
		},
		{
			name:     "canceled order",
			status:   OrderStatusCanceled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Status: tt.status}
			if order.IsActive() != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", order.IsActive(), tt.expected)
			}
		})
	}
}

func TestAssetTypeConstants(t *testing.T) {
	if AssetTypeCollateral != "COLLATERAL" {
		t.Errorf("AssetTypeCollateral = %s, expected COLLATERAL", AssetTypeCollateral)
	}
	if AssetTypeConditional != "CONDITIONAL" {
		t.Errorf("AssetTypeConditional = %s, expected CONDITIONAL", AssetTypeConditional)
	}
}

func TestDecimal6Constant(t *testing.T) {
	if Decimal6 != 1000000 {
		t.Errorf("Decimal6 = %d, expected 1000000", Decimal6)
	}
}

func TestCreateOrderRequest(t *testing.T) {
	req := &CreateOrderRequest{
		TokenID:    "12345",
		Side:       OrderSideBuy,
		Price:      decimal.NewFromFloat(0.55),
		Size:       decimal.NewFromInt(100),
		Type:       OrderTypeGTC,
		ExpiresAt:  0,
		FeeRateBps: 0,
		Nonce:      "0",
		IsNegRisk:  false,
	}

	if req.TokenID != "12345" {
		t.Error("TokenID mismatch")
	}
	if req.Side != OrderSideBuy {
		t.Error("Side mismatch")
	}
	if !req.Price.Equal(decimal.NewFromFloat(0.55)) {
		t.Error("Price mismatch")
	}
}

func TestSignedOrder(t *testing.T) {
	order := &SignedOrder{
		Salt:          "12345",
		Maker:         "0x1234",
		Signer:        "0x1234",
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenId:       "100",
		MakerAmount:   "1000000",
		TakerAmount:   "500000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBps:    "0",
		Side:          "BUY",
		SignatureType: 0,
		Signature:     "0xabcdef",
	}

	if order.Salt != "12345" {
		t.Error("Salt mismatch")
	}
	if order.SignatureType != 0 {
		t.Error("SignatureType mismatch")
	}
}

func TestOrdersQueryParams(t *testing.T) {
	params := &OrdersQueryParams{
		Market:  "market-id",
		AssetID: "asset-id",
		Side:    "BUY",
		Status:  "LIVE",
		Limit:   10,
		Offset:  0,
	}

	if params.Market != "market-id" {
		t.Error("Market mismatch")
	}
	if params.Limit != 10 {
		t.Error("Limit mismatch")
	}
}

func TestTradesQueryParams(t *testing.T) {
	params := &TradesQueryParams{
		Market:  "market-id",
		AssetID: "asset-id",
		Maker:   "0x1234",
		Before:  "2024-12-01",
		After:   "2024-01-01",
		Limit:   50,
	}

	if params.Maker != "0x1234" {
		t.Error("Maker mismatch")
	}
	if params.Limit != 50 {
		t.Error("Limit mismatch")
	}
}

func TestBalanceAllowance(t *testing.T) {
	ba := &BalanceAllowance{
		Balance:   decimal.NewFromInt(1000),
		Allowance: decimal.NewFromInt(500),
	}

	if !ba.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Error("Balance mismatch")
	}
	if !ba.Allowance.Equal(decimal.NewFromInt(500)) {
		t.Error("Allowance mismatch")
	}
}

func TestTrade(t *testing.T) {
	now := time.Now()
	trade := &Trade{
		ID:        "trade-123",
		Market:    "market-456",
		AssetID:   "asset-789",
		Side:      OrderSideBuy,
		Price:     decimal.NewFromFloat(0.65),
		Size:      decimal.NewFromInt(50),
		Fee:       decimal.NewFromFloat(0.01),
		Timestamp: now,
		TradeType: "MAKER",
	}

	if trade.ID != "trade-123" {
		t.Error("ID mismatch")
	}
	if trade.Side != OrderSideBuy {
		t.Error("Side mismatch")
	}
	if trade.TradeType != "MAKER" {
		t.Error("TradeType mismatch")
	}
}

func TestPosition(t *testing.T) {
	pos := &Position{
		TokenID:  "token-123",
		MarketID: "market-456",
		Outcome:  "Yes",
		Size:     decimal.NewFromInt(100),
		AvgPrice: decimal.NewFromFloat(0.55),
		Value:    decimal.NewFromFloat(55.0),
	}

	if pos.TokenID != "token-123" {
		t.Error("TokenID mismatch")
	}
	if pos.Outcome != "Yes" {
		t.Error("Outcome mismatch")
	}
}

func TestCancelOrderRequest(t *testing.T) {
	req := &CancelOrderRequest{
		OrderID: "order-123",
	}

	if req.OrderID != "order-123" {
		t.Error("OrderID mismatch")
	}
}

func TestBatchCancelRequest(t *testing.T) {
	req := &BatchCancelRequest{
		OrderIDs: []string{"order-1", "order-2", "order-3"},
		Market:   "market-123",
		AssetID:  "asset-456",
	}

	if len(req.OrderIDs) != 3 {
		t.Error("OrderIDs length mismatch")
	}
	if req.Market != "market-123" {
		t.Error("Market mismatch")
	}
}

func TestCancelResponse(t *testing.T) {
	resp := &CancelResponse{
		Canceled:    []string{"order-1", "order-2"},
		NotCanceled: []string{"order-3"},
	}

	if len(resp.Canceled) != 2 {
		t.Error("Canceled length mismatch")
	}
	if len(resp.NotCanceled) != 1 {
		t.Error("NotCanceled length mismatch")
	}
}

func TestTickSize(t *testing.T) {
	ts := &TickSize{
		TickSize: decimal.NewFromFloat(0.01),
	}

	if !ts.TickSize.Equal(decimal.NewFromFloat(0.01)) {
		t.Error("TickSize mismatch")
	}
}

func TestPriceInfo(t *testing.T) {
	info := &PriceInfo{
		TokenID: "token-123",
		Price:   decimal.NewFromFloat(0.75),
	}

	if info.TokenID != "token-123" {
		t.Error("TokenID mismatch")
	}
	if !info.Price.Equal(decimal.NewFromFloat(0.75)) {
		t.Error("Price mismatch")
	}
}
