package clob

import (
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"

	"github.com/binary-jerry/polymarket-sdk/auth"
	"github.com/binary-jerry/polymarket-sdk/common"
)

// OrderSigner 订单签名器
type OrderSigner struct {
	signer           *auth.L1Signer
	chainID          int
	exchangeAddr     string // 标准市场交易合约
	negRiskExchange  string // NegRisk 市场交易合约
	negRiskAdapter   string // NegRisk 适配器合约
}

// NewOrderSigner 创建订单签名器
func NewOrderSigner(signer *auth.L1Signer, chainID int, exchangeAddr, negRiskExchange, negRiskAdapter string) *OrderSigner {
	return &OrderSigner{
		signer:          signer,
		chainID:         chainID,
		exchangeAddr:    exchangeAddr,
		negRiskExchange: negRiskExchange,
		negRiskAdapter:  negRiskAdapter,
	}
}

// CreateSignedOrder 创建已签名订单
func (s *OrderSigner) CreateSignedOrder(req *CreateOrderRequest) (*SignedOrder, error) {
	// 生成盐值
	salt, err := common.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// 生成 nonce
	var nonce *big.Int
	if req.Nonce != "" {
		var ok bool
		nonce, ok = new(big.Int).SetString(req.Nonce, 10)
		if !ok {
			return nil, fmt.Errorf("invalid nonce: %s", req.Nonce)
		}
	} else {
		nonce = big.NewInt(0)
	}

	// 计算 makerAmount 和 takerAmount
	makerAmount, takerAmount := s.calculateAmounts(req.Side, req.Price, req.Size)

	// 确定过期时间
	expiration := int64(0)
	if req.ExpiresAt > 0 {
		expiration = req.ExpiresAt
	}

	// 确定 taker 地址
	takerAddr := common.ZeroAddress
	if req.IsNegRisk {
		takerAddr = s.negRiskAdapter
	}

	// 选择正确的交易合约
	exchangeAddr := s.exchangeAddr
	if req.IsNegRisk {
		exchangeAddr = s.negRiskExchange
	}

	// 构建订单载荷
	orderPayload := &auth.OrderPayload{
		Salt:          salt.String(),
		Maker:         s.signer.GetAddress(),
		Signer:        s.signer.GetAddress(),
		Taker:         takerAddr,
		TokenID:       req.TokenID,
		MakerAmount:   makerAmount.String(),
		TakerAmount:   takerAmount.String(),
		Expiration:    fmt.Sprintf("%d", expiration),
		Nonce:         nonce.String(),
		FeeRateBps:    fmt.Sprintf("%d", req.FeeRateBps),
		Side:          req.Side.ToInt(),
		SignatureType: int(auth.SignatureTypeEOA),
		IsNegRisk:     req.IsNegRisk,
	}

	// 签名
	signature, err := s.signer.SignOrder(orderPayload, exchangeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	// 构建已签名订单
	signedOrder := &SignedOrder{
		Salt:          orderPayload.Salt,
		Maker:         orderPayload.Maker,
		Signer:        orderPayload.Signer,
		Taker:         orderPayload.Taker,
		TokenId:       orderPayload.TokenID,
		MakerAmount:   orderPayload.MakerAmount,
		TakerAmount:   orderPayload.TakerAmount,
		Expiration:    orderPayload.Expiration,
		Nonce:         orderPayload.Nonce,
		FeeRateBps:    orderPayload.FeeRateBps,
		Side:          sideToString(req.Side),
		SignatureType: orderPayload.SignatureType,
		Signature:     signature,
	}

	return signedOrder, nil
}

// calculateAmounts 计算 makerAmount 和 takerAmount
// BUY: maker 给 USDC (makerAmount), taker 给 shares (takerAmount)
// SELL: maker 给 shares (makerAmount), taker 给 USDC (takerAmount)
func (s *OrderSigner) calculateAmounts(side OrderSide, price, size decimal.Decimal) (*big.Int, *big.Int) {
	// USDC 有 6 位小数
	usdcDecimals := decimal.NewFromInt(Decimal6)

	// 计算 USDC 数量 = price * size * 10^6
	usdcAmount := price.Mul(size).Mul(usdcDecimals)

	// 计算 shares 数量 = size * 10^6
	sharesAmount := size.Mul(usdcDecimals)

	usdcBigInt := usdcAmount.BigInt()
	sharesBigInt := sharesAmount.BigInt()

	if side == OrderSideBuy {
		// BUY: maker 给 USDC, taker 给 shares
		return usdcBigInt, sharesBigInt
	}

	// SELL: maker 给 shares, taker 给 USDC
	return sharesBigInt, usdcBigInt
}

// sideToString 将 OrderSide 转换为字符串
func sideToString(side OrderSide) string {
	if side == OrderSideBuy {
		return "BUY"
	}
	return "SELL"
}

// GetExchangeAddress 获取交易合约地址
func (s *OrderSigner) GetExchangeAddress(isNegRisk bool) string {
	if isNegRisk {
		return s.negRiskExchange
	}
	return s.exchangeAddr
}

// GetNegRiskAdapter 获取 NegRisk 适配器地址
func (s *OrderSigner) GetNegRiskAdapter() string {
	return s.negRiskAdapter
}

// GetSignerAddress 获取签名者地址
func (s *OrderSigner) GetSignerAddress() string {
	return s.signer.GetAddress()
}
