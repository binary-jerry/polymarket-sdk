package clob

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
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
	funderAddress    string // 代理钱包地址（持有资金）
	signatureType    int    // 签名类型: 0=EOA, 1=POLY_PROXY, 2=GNOSIS_SAFE
}

// NewOrderSigner 创建订单签名器
func NewOrderSigner(signer *auth.L1Signer, chainID int, exchangeAddr, negRiskExchange, negRiskAdapter string) *OrderSigner {
	return &OrderSigner{
		signer:          signer,
		chainID:         chainID,
		exchangeAddr:    exchangeAddr,
		negRiskExchange: negRiskExchange,
		negRiskAdapter:  negRiskAdapter,
		signatureType:   int(auth.SignatureTypeEOA), // 默认 EOA 模式
	}
}

// SetFunderAddress 设置代理钱包地址（用于代理钱包模式）
func (s *OrderSigner) SetFunderAddress(addr string) {
	s.funderAddress = addr
}

// SetSignatureType 设置签名类型
func (s *OrderSigner) SetSignatureType(sigType int) {
	s.signatureType = sigType
}

// GetMakerAddress 获取 Maker 地址（如果设置了 funder 则返回 funder，否则返回签名者地址）
// 返回 checksum 格式的地址
func (s *OrderSigner) GetMakerAddress() string {
	var addr string
	if s.funderAddress != "" {
		addr = s.funderAddress
	} else {
		addr = s.signer.GetAddress()
	}
	// 转换为 checksum 格式
	return ethcommon.HexToAddress(addr).Hex()
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
	// 注意：NegRisk 市场也使用零地址作为 taker（2024年后 API 规则变更）
	takerAddr := common.ZeroAddress

	// 选择正确的交易合约
	exchangeAddr := s.exchangeAddr
	if req.IsNegRisk {
		exchangeAddr = s.negRiskExchange
	}

	// 确定 Maker 地址（代理钱包模式使用 funder，否则使用签名者地址）
	// 重要：所有地址必须使用 checksum 格式，以确保签名和提交时使用相同格式
	makerAddr := ethcommon.HexToAddress(s.GetMakerAddress()).Hex()
	signerAddr := ethcommon.HexToAddress(s.signer.GetAddress()).Hex()
	takerAddrChecksum := ethcommon.HexToAddress(takerAddr).Hex()

	// 构建订单载荷
	orderPayload := &auth.OrderPayload{
		Salt:          salt.String(),
		Maker:         makerAddr,              // 代理钱包模式: funder 地址; EOA 模式: 签名者地址
		Signer:        signerAddr,             // 始终是签名钱包地址
		Taker:         takerAddrChecksum,
		TokenID:       req.TokenID,
		MakerAmount:   makerAmount.String(),
		TakerAmount:   takerAmount.String(),
		Expiration:    fmt.Sprintf("%d", expiration),
		Nonce:         nonce.String(),
		FeeRateBps:    fmt.Sprintf("%d", req.FeeRateBps),
		Side:          req.Side.ToInt(),
		SignatureType: s.signatureType,        // 使用配置的签名类型
		IsNegRisk:     req.IsNegRisk,
	}

	// 签名
	signature, err := s.signer.SignOrder(orderPayload, exchangeAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	// 构建已签名订单
	// 将 salt 字符串转为 int64
	saltInt := salt.Int64()
	signedOrder := &SignedOrder{
		Salt:          saltInt,
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
//
// Polymarket 精度限制:
// - makerAmount (USDC 金额 = price * size): 最多 2 位小数
// - takerAmount (shares 数量 = size): 最多 4 位小数
func (s *OrderSigner) calculateAmounts(side OrderSide, price, size decimal.Decimal) (*big.Int, *big.Int) {
	// USDC 有 6 位小数
	usdcDecimals := decimal.NewFromInt(Decimal6)

	// 精度截断 (Truncate 向下截断，避免超出可用余额)
	// price: 最多 4 位小数
	// size: 最多 2 位小数 (保证 price * size 最多 6 位小数，且符合 maker amount 2位精度限制)
	truncatedPrice := price.Truncate(4)
	truncatedSize := size.Truncate(2)

	// 计算 USDC 数量 = price * size * 10^6
	// 截断到 2 位小数后再乘以 10^6，确保是整数
	usdcRaw := truncatedPrice.Mul(truncatedSize).Truncate(2)
	usdcAmount := usdcRaw.Mul(usdcDecimals)

	// 计算 shares 数量 = size * 10^6
	sharesAmount := truncatedSize.Mul(usdcDecimals)

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
