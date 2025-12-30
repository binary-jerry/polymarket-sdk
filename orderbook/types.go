package orderbook

import (
	"github.com/shopspring/decimal"
)

// ConnectionState WebSocket连接状态
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateSubscribing
	StateActive
	StateReconnecting
	StateClosed
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateSubscribing:
		return "Subscribing"
	case StateActive:
		return "Active"
	case StateReconnecting:
		return "Reconnecting"
	case StateClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// EventType 消息事件类型
type EventType string

const (
	EventTypeBook           EventType = "book"
	EventTypePriceChange    EventType = "price_change"
	EventTypeTickSizeChange EventType = "tick_size_change"
	EventTypeLastTradePrice EventType = "last_trade_price"
)

// Side 买卖方向
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// OrderSummary 订单摘要（价格档位）
type OrderSummary struct {
	Price decimal.Decimal
	Size  decimal.Decimal
}

// RawOrderSummary 原始订单摘要（字符串格式）
type RawOrderSummary struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// BookMessage 订单簿完整快照消息
type BookMessage struct {
	EventType EventType         `json:"event_type"`
	AssetID   string            `json:"asset_id"`
	Market    string            `json:"market"`
	Timestamp string            `json:"timestamp"`
	Hash      string            `json:"hash"`
	Bids      []RawOrderSummary `json:"bids"`
	Asks      []RawOrderSummary `json:"asks"`
}

// PriceChange 价格变动
type PriceChange struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Side    Side   `json:"side"`
	Hash    string `json:"hash"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
}

// PriceChangeMessage 价格变动消息
type PriceChangeMessage struct {
	EventType    EventType     `json:"event_type"`
	Market       string        `json:"market"`
	PriceChanges []PriceChange `json:"price_changes"`
	Timestamp    string        `json:"timestamp"`
}

// TickSizeChangeMessage tick size变更消息
type TickSizeChangeMessage struct {
	EventType   EventType `json:"event_type"`
	AssetID     string    `json:"asset_id"`
	Market      string    `json:"market"`
	OldTickSize string    `json:"old_tick_size"`
	NewTickSize string    `json:"new_tick_size"`
	Timestamp   string    `json:"timestamp"`
}

// LastTradePriceMessage 最后成交价消息
type LastTradePriceMessage struct {
	EventType  EventType `json:"event_type"`
	AssetID    string    `json:"asset_id"`
	Market     string    `json:"market"`
	Price      string    `json:"price"`
	Side       Side      `json:"side"`
	Size       string    `json:"size"`
	FeeRateBps string    `json:"fee_rate_bps"`
	Timestamp  string    `json:"timestamp"`
}

// RawMessage 原始消息（用于类型判断）
type RawMessage struct {
	EventType EventType `json:"event_type"`
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	AssetsIDs []string `json:"assets_ids"`
	Type      string   `json:"type"`
}

// OrderBookUpdate 订单簿更新事件（通过channel通知）
type OrderBookUpdate struct {
	TokenID   string
	EventType EventType
	Timestamp int64
}

// BestPrice 最优价格（包含价格和数量）
type BestPrice struct {
	Price     decimal.Decimal
	Size      decimal.Decimal
	Timestamp int64
}

// BBO 最优买卖价（Best Bid and Offer）
type BBO struct {
	BestBid *BestPrice
	BestAsk *BestPrice
}

// ScanResult 扫描结果
type ScanResult struct {
	Orders    []OrderSummary  // 符合条件的订单列表
	TotalSize decimal.Decimal // 总数量
	AvgPrice  decimal.Decimal // 加权平均价格
}

// Config SDK配置
type Config struct {
	// WebSocket端点
	WSEndpoint string
	// 每个连接最大token数量
	MaxTokensPerConn int
	// 重连配置
	ReconnectMinInterval int // 最小重连间隔（毫秒）
	ReconnectMaxInterval int // 最大重连间隔（毫秒）
	ReconnectMaxAttempts int // 最大重连次数，0表示无限
	// 心跳配置
	PingInterval int // ping间隔（秒）
	PongTimeout  int // pong超时（秒）
	// 消息缓冲区大小
	MessageBufferSize int
	// 更新通知channel缓冲区大小
	UpdateChannelSize int
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		WSEndpoint:           "wss://ws-subscriptions-clob.polymarket.com/ws/market",
		MaxTokensPerConn:     50,
		ReconnectMinInterval: 1000,
		ReconnectMaxInterval: 30000,
		ReconnectMaxAttempts: 0, // 无限重连
		PingInterval:         30,
		PongTimeout:          10,
		MessageBufferSize:    1000,
		UpdateChannelSize:    1000,
	}
}
