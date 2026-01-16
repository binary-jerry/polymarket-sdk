package clob

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderType 订单类型
type OrderType string

const (
	// OrderTypeGTC Good Till Cancelled - 持续有效直到成交或取消
	OrderTypeGTC OrderType = "GTC"
	// OrderTypeGTD Good Till Date - 有效至指定时间
	OrderTypeGTD OrderType = "GTD"
	// OrderTypeFOK Fill Or Kill - 必须完全成交否则取消
	OrderTypeFOK OrderType = "FOK"
	// OrderTypeFAK Fill And Kill - 立即成交可成交部分，余额取消
	OrderTypeFAK OrderType = "FAK"
)

// OrderSide 订单方向
type OrderSide string

const (
	// OrderSideBuy 买入
	OrderSideBuy OrderSide = "BUY"
	// OrderSideSell 卖出
	OrderSideSell OrderSide = "SELL"
)

// ToInt 转换为整数
func (s OrderSide) ToInt() int {
	if s == OrderSideBuy {
		return 0
	}
	return 1
}

// OrderStatus 订单状态
type OrderStatus string

const (
	// OrderStatusLive 活跃
	OrderStatusLive OrderStatus = "LIVE"
	// OrderStatusMatched 已匹配
	OrderStatusMatched OrderStatus = "MATCHED"
	// OrderStatusCanceled 已取消
	OrderStatusCanceled OrderStatus = "CANCELED"
	// OrderStatusDelayed 延迟处理
	OrderStatusDelayed OrderStatus = "DELAYED"
)

// Order 订单
type Order struct {
	ID              string          `json:"id"`
	OrderID         string          `json:"order_id,omitempty"`
	Market          string          `json:"market"`
	AssetID         string          `json:"asset_id"`
	Side            OrderSide       `json:"side"`
	Type            OrderType       `json:"type,omitempty"`
	Price           decimal.Decimal `json:"price"`
	OriginalSize    decimal.Decimal `json:"original_size"`
	SizeMatched     decimal.Decimal `json:"size_matched"`
	Status          OrderStatus     `json:"status"`
	Owner           string          `json:"owner"`
	Outcome         string          `json:"outcome,omitempty"`

	// 合约相关
	MakerAddress    string          `json:"maker_address,omitempty"`
	Salt            string          `json:"salt,omitempty"`
	Signature       string          `json:"signature,omitempty"`

	// 时间
	CreatedAt       time.Time       `json:"created_at,omitempty"`
	ExpiresAt       *time.Time      `json:"expiration,omitempty"`

	// 费率
	FeeRateBps      int             `json:"fee_rate_bps,omitempty"`
}

// GetRemainingSize 获取剩余数量
func (o *Order) GetRemainingSize() decimal.Decimal {
	return o.OriginalSize.Sub(o.SizeMatched)
}

// IsFilled 是否已完全成交
func (o *Order) IsFilled() bool {
	return o.SizeMatched.GreaterThanOrEqual(o.OriginalSize)
}

// IsActive 是否活跃
func (o *Order) IsActive() bool {
	return o.Status == OrderStatusLive || o.Status == OrderStatusDelayed
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	TokenID       string          `json:"tokenID"`
	Side          OrderSide       `json:"side"`
	Price         decimal.Decimal `json:"price"`
	Size          decimal.Decimal `json:"size"`
	Type          OrderType       `json:"type,omitempty"`
	ExpiresAt     int64           `json:"expiration,omitempty"`  // GTD 订单的过期时间戳
	FeeRateBps    int             `json:"feeRateBps,omitempty"`
	Nonce         string          `json:"nonce,omitempty"`

	// NegRisk 标识（内部使用）
	IsNegRisk     bool            `json:"-"`
}

// SignedOrder 已签名订单
type SignedOrder struct {
	Salt          int64  `json:"salt"`           // 数字类型，与 Python SDK 一致
	Maker         string `json:"maker"`
	Signer        string `json:"signer"`
	Taker         string `json:"taker"`
	TokenId       string `json:"tokenId"`
	MakerAmount   string `json:"makerAmount"`
	TakerAmount   string `json:"takerAmount"`
	Expiration    string `json:"expiration"`
	Nonce         string `json:"nonce"`
	FeeRateBps    string `json:"feeRateBps"`
	Side          string `json:"side"`
	SignatureType int    `json:"signatureType"`
	Signature     string `json:"signature"`
}

// PostOrderRequest 提交订单请求
type PostOrderRequest struct {
	Order     *SignedOrder `json:"order"`
	Owner     string       `json:"owner"`
	OrderType OrderType    `json:"orderType"`
}

// OrdersQueryParams 订单查询参数
type OrdersQueryParams struct {
	Market    string `url:"market,omitempty"`
	AssetID   string `url:"asset_id,omitempty"`
	Side      string `url:"side,omitempty"`
	Status    string `url:"status,omitempty"`
	Limit     int    `url:"limit,omitempty"`
	Offset    int    `url:"offset,omitempty"`
}

// OrderResponse 订单响应
type OrderResponse struct {
	Success  bool   `json:"success"`
	OrderID  string `json:"orderID,omitempty"`
	Status   string `json:"status,omitempty"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

// Trade 成交记录
type Trade struct {
	ID              string          `json:"id"`
	Market          string          `json:"market"`
	AssetID         string          `json:"asset_id"`
	Side            OrderSide       `json:"side"`
	Price           decimal.Decimal `json:"price"`
	Size            decimal.Decimal `json:"size"`
	Fee             decimal.Decimal `json:"fee,omitempty"`
	Timestamp       time.Time       `json:"timestamp"`
	Owner           string          `json:"owner,omitempty"`
	OrderID         string          `json:"order_id,omitempty"`
	MatchedOrderID  string          `json:"matched_order_id,omitempty"`
	TradeType       string          `json:"trade_type,omitempty"`  // "MAKER" 或 "TAKER"
	TransactionHash string          `json:"transaction_hash,omitempty"`
}

// TradesQueryParams 交易查询参数
type TradesQueryParams struct {
	Market    string `url:"market,omitempty"`
	AssetID   string `url:"asset_id,omitempty"`
	Maker     string `url:"maker,omitempty"`
	Before    string `url:"before,omitempty"`
	After     string `url:"after,omitempty"`
	Limit     int    `url:"limit,omitempty"`
}

// BalanceAllowance 余额和授权
type BalanceAllowance struct {
	Balance   decimal.Decimal `json:"balance"`
	Allowance decimal.Decimal `json:"allowance"`
}

// AssetType 资产类型
type AssetType string

const (
	// AssetTypeCollateral 抵押品 (USDC)
	AssetTypeCollateral AssetType = "COLLATERAL"
	// AssetTypeConditional 条件代币
	AssetTypeConditional AssetType = "CONDITIONAL"
)

// BalanceAllowanceParams 余额查询参数
type BalanceAllowanceParams struct {
	AssetType AssetType `url:"asset_type"`
	TokenID   string    `url:"token_id,omitempty"`
}

// Position 持仓
type Position struct {
	TokenID     string          `json:"token_id"`
	MarketID    string          `json:"market_id,omitempty"`
	Outcome     string          `json:"outcome"`  // "Yes" 或 "No"
	Size        decimal.Decimal `json:"size"`
	AvgPrice    decimal.Decimal `json:"avg_price,omitempty"`
	Value       decimal.Decimal `json:"value,omitempty"`
}

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	OrderID string `json:"orderID"`
}

// BatchCancelRequest 批量取消请求
type BatchCancelRequest struct {
	OrderIDs []string `json:"orderIDs,omitempty"`
	Market   string   `json:"market,omitempty"`
	AssetID  string   `json:"asset_id,omitempty"`
}

// CancelResponse 取消订单响应
type CancelResponse struct {
	Canceled []string `json:"canceled,omitempty"`
	NotCanceled []string `json:"not_canceled,omitempty"`
}

// TickSize 价格最小变动单位
type TickSize struct {
	TickSize decimal.Decimal `json:"minimum_tick_size"`
}

// PriceInfo 价格信息
type PriceInfo struct {
	TokenID string          `json:"token_id"`
	Price   decimal.Decimal `json:"price"`
}

// Decimal6 USDC 精度 (6 位小数)
const Decimal6 = 1000000
