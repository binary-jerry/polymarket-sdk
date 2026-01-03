package clob

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/polymarket-sdk/auth"
)

// 测试用私钥
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewOrderSigner(t *testing.T) {
	signer, err := auth.NewL1Signer(testPrivateKey, 137)
	if err != nil {
		t.Fatalf("Failed to create L1Signer: %v", err)
	}

	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	if orderSigner == nil {
		t.Fatal("NewOrderSigner() returned nil")
	}
}

func TestOrderSignerCreateSignedOrder(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	req := &CreateOrderRequest{
		TokenID:    "12345",
		Side:       OrderSideBuy,
		Price:      decimal.NewFromFloat(0.55),
		Size:       decimal.NewFromInt(100),
		Type:       OrderTypeGTC,
		FeeRateBps: 0,
		IsNegRisk:  false,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder == nil {
		t.Fatal("CreateSignedOrder() returned nil")
	}

	// Verify fields
	if signedOrder.Salt == "" {
		t.Error("Salt should not be empty")
	}
	if signedOrder.Maker == "" {
		t.Error("Maker should not be empty")
	}
	if signedOrder.Signer == "" {
		t.Error("Signer should not be empty")
	}
	if signedOrder.TokenId != "12345" {
		t.Errorf("TokenId = %s, expected 12345", signedOrder.TokenId)
	}
	if signedOrder.Signature == "" {
		t.Error("Signature should not be empty")
	}
	if !strings.HasPrefix(signedOrder.Signature, "0x") {
		t.Error("Signature should start with 0x")
	}
}

func TestOrderSignerCreateSignedOrderSellSide(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	req := &CreateOrderRequest{
		TokenID:    "12345",
		Side:       OrderSideSell,
		Price:      decimal.NewFromFloat(0.45),
		Size:       decimal.NewFromInt(50),
		Type:       OrderTypeGTC,
		FeeRateBps: 0,
		IsNegRisk:  false,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder.Side != "SELL" {
		t.Errorf("Side = %s, expected SELL", signedOrder.Side)
	}
}

func TestOrderSignerCreateSignedOrderNegRisk(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	negRiskAdapter := "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		negRiskAdapter,
	)

	req := &CreateOrderRequest{
		TokenID:   "12345",
		Side:      OrderSideBuy,
		Price:     decimal.NewFromFloat(0.55),
		Size:      decimal.NewFromInt(100),
		IsNegRisk: true,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	// For NegRisk, taker should be the adapter address (case-insensitive comparison)
	if !strings.EqualFold(signedOrder.Taker, negRiskAdapter) {
		t.Errorf("Taker = %s, expected %s for NegRisk order", signedOrder.Taker, negRiskAdapter)
	}
}

func TestOrderSignerCreateSignedOrderWithExpiration(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	expiration := int64(1735689600) // 2025-01-01
	req := &CreateOrderRequest{
		TokenID:   "12345",
		Side:      OrderSideBuy,
		Price:     decimal.NewFromFloat(0.55),
		Size:      decimal.NewFromInt(100),
		Type:      OrderTypeGTD,
		ExpiresAt: expiration,
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder.Expiration != "1735689600" {
		t.Errorf("Expiration = %s, expected 1735689600", signedOrder.Expiration)
	}
}

func TestOrderSignerCreateSignedOrderWithNonce(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	req := &CreateOrderRequest{
		TokenID: "12345",
		Side:    OrderSideBuy,
		Price:   decimal.NewFromFloat(0.55),
		Size:    decimal.NewFromInt(100),
		Nonce:   "42",
	}

	signedOrder, err := orderSigner.CreateSignedOrder(req)
	if err != nil {
		t.Fatalf("CreateSignedOrder() error: %v", err)
	}

	if signedOrder.Nonce != "42" {
		t.Errorf("Nonce = %s, expected 42", signedOrder.Nonce)
	}
}

func TestOrderSignerCreateSignedOrderInvalidNonce(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	req := &CreateOrderRequest{
		TokenID: "12345",
		Side:    OrderSideBuy,
		Price:   decimal.NewFromFloat(0.55),
		Size:    decimal.NewFromInt(100),
		Nonce:   "invalid",
	}

	_, err := orderSigner.CreateSignedOrder(req)
	if err == nil {
		t.Error("CreateSignedOrder() should fail with invalid nonce")
	}
}

func TestOrderSignerGetExchangeAddress(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	exchangeAddr := "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e"
	negRiskExchange := "0xC5d563A36AE78145C45a50134d48A1215220f80a"
	orderSigner := NewOrderSigner(
		signer,
		137,
		exchangeAddr,
		negRiskExchange,
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	if orderSigner.GetExchangeAddress(false) != exchangeAddr {
		t.Errorf("GetExchangeAddress(false) = %s, expected %s", orderSigner.GetExchangeAddress(false), exchangeAddr)
	}
	if orderSigner.GetExchangeAddress(true) != negRiskExchange {
		t.Errorf("GetExchangeAddress(true) = %s, expected %s", orderSigner.GetExchangeAddress(true), negRiskExchange)
	}
}

func TestOrderSignerGetNegRiskAdapter(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	negRiskAdapter := "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		negRiskAdapter,
	)

	if orderSigner.GetNegRiskAdapter() != negRiskAdapter {
		t.Errorf("GetNegRiskAdapter() = %s, expected %s", orderSigner.GetNegRiskAdapter(), negRiskAdapter)
	}
}

func TestOrderSignerGetSignerAddress(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	addr := orderSigner.GetSignerAddress()
	if addr == "" {
		t.Error("GetSignerAddress() should not return empty string")
	}
	if !strings.HasPrefix(addr, "0x") {
		t.Error("GetSignerAddress() should return address starting with 0x")
	}
}

func TestSideToString(t *testing.T) {
	if sideToString(OrderSideBuy) != "BUY" {
		t.Errorf("sideToString(OrderSideBuy) = %s, expected BUY", sideToString(OrderSideBuy))
	}
	if sideToString(OrderSideSell) != "SELL" {
		t.Errorf("sideToString(OrderSideSell) = %s, expected SELL", sideToString(OrderSideSell))
	}
}

func TestCalculateAmountsBuy(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	// BUY: price = 0.5, size = 100
	// makerAmount = price * size * 10^6 = 0.5 * 100 * 1000000 = 50000000
	// takerAmount = size * 10^6 = 100 * 1000000 = 100000000
	makerAmount, takerAmount := orderSigner.calculateAmounts(
		OrderSideBuy,
		decimal.NewFromFloat(0.5),
		decimal.NewFromInt(100),
	)

	expectedMaker := int64(50000000)
	expectedTaker := int64(100000000)

	if makerAmount.Int64() != expectedMaker {
		t.Errorf("makerAmount = %d, expected %d", makerAmount.Int64(), expectedMaker)
	}
	if takerAmount.Int64() != expectedTaker {
		t.Errorf("takerAmount = %d, expected %d", takerAmount.Int64(), expectedTaker)
	}
}

func TestCalculateAmountsSell(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	// SELL: price = 0.5, size = 100
	// makerAmount = size * 10^6 = 100 * 1000000 = 100000000 (shares)
	// takerAmount = price * size * 10^6 = 0.5 * 100 * 1000000 = 50000000 (USDC)
	makerAmount, takerAmount := orderSigner.calculateAmounts(
		OrderSideSell,
		decimal.NewFromFloat(0.5),
		decimal.NewFromInt(100),
	)

	expectedMaker := int64(100000000) // shares
	expectedTaker := int64(50000000)  // USDC

	if makerAmount.Int64() != expectedMaker {
		t.Errorf("makerAmount = %d, expected %d", makerAmount.Int64(), expectedMaker)
	}
	if takerAmount.Int64() != expectedTaker {
		t.Errorf("takerAmount = %d, expected %d", takerAmount.Int64(), expectedTaker)
	}
}

func TestOrderSignerDeterministicWithSameNonceAndSalt(t *testing.T) {
	signer, _ := auth.NewL1Signer(testPrivateKey, 137)
	orderSigner := NewOrderSigner(
		signer,
		137,
		"0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		"0xC5d563A36AE78145C45a50134d48A1215220f80a",
		"0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
	)

	req := &CreateOrderRequest{
		TokenID: "12345",
		Side:    OrderSideBuy,
		Price:   decimal.NewFromFloat(0.55),
		Size:    decimal.NewFromInt(100),
	}

	// Create two orders (salt will be different due to random generation)
	order1, _ := orderSigner.CreateSignedOrder(req)
	order2, _ := orderSigner.CreateSignedOrder(req)

	// Salt should be different (random)
	if order1.Salt == order2.Salt {
		t.Error("Each order should have a unique salt")
	}
}
