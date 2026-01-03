package auth

// Signer 签名器接口
type Signer interface {
	// GetAddress 获取钱包地址
	GetAddress() string

	// SignMessage 签名消息
	SignMessage(message []byte) ([]byte, error)

	// SignTypedData 签名 EIP-712 类型数据
	SignTypedData(typedData *TypedData) ([]byte, error)
}

// OrderSigner 订单签名器接口
type OrderSigner interface {
	// SignOrder 签名订单
	SignOrder(order *OrderPayload) (string, error)
}

// OrderPayload 订单载荷（用于签名）
type OrderPayload struct {
	// 基础参数
	Salt        string // 256位随机数
	Maker       string // 订单创建者地址
	Signer      string // 签名者地址
	Taker       string // 接收者地址 (通常为零地址)
	TokenID     string // ERC1155 token ID
	MakerAmount string // 创建者支付金额
	TakerAmount string // 接收者支付金额
	Expiration  string // 过期时间 (Unix timestamp)
	Nonce       string // 链上 nonce
	FeeRateBps  string // 费率 (基点)
	Side        int    // 0=BUY, 1=SELL
	SignatureType int  // 签名类型

	// NegRisk 标识
	IsNegRisk bool
}

// OrderSide 订单方向常量
const (
	OrderSideBuy  = 0
	OrderSideSell = 1
)
