package gamma

import (
	"time"

	"github.com/binary-jerry/polymarket-sdk/common"
)

// Client Gamma API 客户端
type Client struct {
	httpClient *common.HTTPClient
	config     *Config
}

// Config Gamma 客户端配置
type Config struct {
	Endpoint     string        // API 端点
	Timeout      time.Duration // 请求超时
	MaxRetries   int           // 最大重试次数
	RetryDelayMs int           // 重试间隔（毫秒）
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint:     "https://gamma-api.polymarket.com",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryDelayMs: 1000,
	}
}

// NewClient 创建 Gamma 客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	httpConfig := &common.HTTPClientConfig{
		BaseURL:      config.Endpoint,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	}

	return &Client{
		httpClient: common.NewHTTPClient(httpConfig),
		config:     config,
	}
}

// Close 关闭客户端
func (c *Client) Close() {
	// HTTP 客户端无需显式关闭
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}
