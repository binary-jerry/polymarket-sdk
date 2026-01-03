package polymarket

import "time"

// ChainID Polygon 主网链 ID
const ChainID = 137

// API 端点常量
const (
	// GammaEndpoint Gamma API 端点（市场数据）
	GammaEndpoint = "https://gamma-api.polymarket.com"

	// CLOBEndpoint CLOB API 端点（交易）
	CLOBEndpoint = "https://clob.polymarket.com"

	// WSEndpoint WebSocket 端点（订单簿）
	WSEndpoint = "wss://ws-subscriptions-clob.polymarket.com/ws/market"
)

// 合约地址常量 (Polygon Mainnet)
const (
	// CTFExchangeAddress 标准市场交易合约
	CTFExchangeAddress = "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e"

	// NegRiskCTFExchangeAddress NegRisk 市场交易合约
	NegRiskCTFExchangeAddress = "0xC5d563A36AE78145C45a50134d48A1215220f80a"

	// NegRiskAdapterAddress NegRisk 适配器合约
	NegRiskAdapterAddress = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"

	// CollateralAddress USDC 抵押品合约
	CollateralAddress = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"

	// ConditionalTokensAddress 条件代币合约
	ConditionalTokensAddress = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"
)

// Config SDK 全局配置
type Config struct {
	// API 端点配置
	GammaEndpoint string // Gamma API 端点
	CLOBEndpoint  string // CLOB API 端点
	WSEndpoint    string // WebSocket 端点

	// HTTP 配置
	HTTPTimeout   time.Duration // HTTP 请求超时
	MaxRetries    int           // 最大重试次数
	RetryDelayMs  int           // 重试间隔（毫秒）

	// WebSocket 配置（订单簿）
	MaxTokensPerConn     int // 每个连接最大 token 数
	ReconnectMinInterval int // 最小重连间隔（毫秒）
	ReconnectMaxInterval int // 最大重连间隔（毫秒）
	ReconnectMaxAttempts int // 最大重连次数，0 表示无限
	PingInterval         int // ping 间隔（秒）
	PongTimeout          int // pong 超时（秒）
	MessageBufferSize    int // 消息缓冲区大小
	UpdateChannelSize    int // 更新通知 channel 大小

	// 合约地址配置
	CTFExchangeAddress        string // 标准市场交易合约
	NegRiskCTFExchangeAddress string // NegRisk 市场交易合约
	NegRiskAdapterAddress     string // NegRisk 适配器合约
	CollateralAddress         string // 抵押品合约地址
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		// API 端点
		GammaEndpoint: GammaEndpoint,
		CLOBEndpoint:  CLOBEndpoint,
		WSEndpoint:    WSEndpoint,

		// HTTP 配置
		HTTPTimeout:  30 * time.Second,
		MaxRetries:   3,
		RetryDelayMs: 1000,

		// WebSocket 配置
		MaxTokensPerConn:     50,
		ReconnectMinInterval: 1000,
		ReconnectMaxInterval: 30000,
		ReconnectMaxAttempts: 0, // 无限重连
		PingInterval:         30,
		PongTimeout:          10,
		MessageBufferSize:    1000,
		UpdateChannelSize:    1000,

		// 合约地址
		CTFExchangeAddress:        CTFExchangeAddress,
		NegRiskCTFExchangeAddress: NegRiskCTFExchangeAddress,
		NegRiskAdapterAddress:     NegRiskAdapterAddress,
		CollateralAddress:         CollateralAddress,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.GammaEndpoint == "" {
		c.GammaEndpoint = GammaEndpoint
	}
	if c.CLOBEndpoint == "" {
		c.CLOBEndpoint = CLOBEndpoint
	}
	if c.WSEndpoint == "" {
		c.WSEndpoint = WSEndpoint
	}
	if c.HTTPTimeout == 0 {
		c.HTTPTimeout = 30 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryDelayMs == 0 {
		c.RetryDelayMs = 1000
	}
	if c.MaxTokensPerConn == 0 {
		c.MaxTokensPerConn = 50
	}
	if c.ReconnectMinInterval == 0 {
		c.ReconnectMinInterval = 1000
	}
	if c.ReconnectMaxInterval == 0 {
		c.ReconnectMaxInterval = 30000
	}
	if c.PingInterval == 0 {
		c.PingInterval = 30
	}
	if c.PongTimeout == 0 {
		c.PongTimeout = 10
	}
	if c.MessageBufferSize == 0 {
		c.MessageBufferSize = 1000
	}
	if c.UpdateChannelSize == 0 {
		c.UpdateChannelSize = 1000
	}
	if c.CTFExchangeAddress == "" {
		c.CTFExchangeAddress = CTFExchangeAddress
	}
	if c.NegRiskCTFExchangeAddress == "" {
		c.NegRiskCTFExchangeAddress = NegRiskCTFExchangeAddress
	}
	if c.NegRiskAdapterAddress == "" {
		c.NegRiskAdapterAddress = NegRiskAdapterAddress
	}
	if c.CollateralAddress == "" {
		c.CollateralAddress = CollateralAddress
	}
	return nil
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}
