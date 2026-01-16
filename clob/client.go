package clob

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/binary-jerry/polymarket-sdk/auth"
	"github.com/binary-jerry/polymarket-sdk/common"
)

// Client CLOB API 客户端
type Client struct {
	mu sync.RWMutex

	httpClient   *common.HTTPClient
	config       *Config

	// 认证
	l1Signer     *auth.L1Signer
	l2Signer     *auth.L2Signer
	credentials  *auth.Credentials

	// 订单签名
	orderSigner  *OrderSigner
}

// Config CLOB 模块配置
type Config struct {
	Endpoint             string        // API 端点
	ChainID              int           // 链 ID
	Timeout              time.Duration // 请求超时
	MaxRetries           int           // 最大重试次数
	RetryDelayMs         int           // 重试间隔

	// 合约地址
	ExchangeAddress        string // 标准市场交易合约
	NegRiskExchangeAddress string // NegRisk 市场交易合约
	NegRiskAdapterAddress  string // NegRisk 适配器合约
	CollateralAddress      string // 抵押品合约地址
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Endpoint:               "https://clob.polymarket.com",
		ChainID:                137,
		Timeout:                30 * time.Second,
		MaxRetries:             3,
		RetryDelayMs:           1000,
		ExchangeAddress:        "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e",
		NegRiskExchangeAddress: "0xC5d563A36AE78145C45a50134d48A1215220f80a",
		NegRiskAdapterAddress:  "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
		CollateralAddress:      "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
	}
}

// NewClient 创建 CLOB 客户端
func NewClient(config *Config, privateKey string) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	l1Signer, err := auth.NewL1Signer(privateKey, config.ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	httpConfig := &common.HTTPClientConfig{
		BaseURL:      config.Endpoint,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	}

	orderSigner := NewOrderSigner(
		l1Signer,
		config.ChainID,
		config.ExchangeAddress,
		config.NegRiskExchangeAddress,
		config.NegRiskAdapterAddress,
	)

	return &Client{
		httpClient:  common.NewHTTPClient(httpConfig),
		config:      config,
		l1Signer:    l1Signer,
		orderSigner: orderSigner,
	}, nil
}

// NewClientWithCredentials 使用已有凭证创建客户端
func NewClientWithCredentials(config *Config, privateKey string, creds *auth.Credentials) (*Client, error) {
	client, err := NewClient(config, privateKey)
	if err != nil {
		return nil, err
	}

	client.credentials = creds
	client.l2Signer = auth.NewL2Signer(client.l1Signer.GetAddress(), creds)

	return client, nil
}

// Close 关闭客户端
func (c *Client) Close() {
	// HTTP 客户端无需显式关闭
}

// GetAddress 获取钱包地址
func (c *Client) GetAddress() string {
	return c.l1Signer.GetAddress()
}

// GetCredentials 获取当前凭证
func (c *Client) GetCredentials() *auth.Credentials {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.credentials
}

// SetCredentials 设置凭证
func (c *Client) SetCredentials(creds *auth.Credentials) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.credentials = creds
	c.l2Signer = auth.NewL2Signer(c.l1Signer.GetAddress(), creds)
}

// SetCredentialsWithAddress 设置凭证（指定账户地址）
func (c *Client) SetCredentialsWithAddress(creds *auth.Credentials, address string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.credentials = creds
	c.l2Signer = auth.NewL2Signer(address, creds)
}

// SetFunderAddress 设置代理钱包地址（用于代理钱包模式）
// funderAddress: 代理钱包地址（持有资金的地址）
func (c *Client) SetFunderAddress(funderAddress string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.orderSigner != nil {
		c.orderSigner.SetFunderAddress(funderAddress)
	}
}

// SetSignatureType 设置签名类型
// signatureType: 0=EOA, 1=POLY_PROXY, 2=GNOSIS_SAFE
func (c *Client) SetSignatureType(signatureType int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.orderSigner != nil {
		c.orderSigner.SetSignatureType(signatureType)
	}
}

// GetFunderAddress 获取 Maker 地址（代理钱包或签名钱包）
func (c *Client) GetFunderAddress() string {
	if c.orderSigner != nil {
		return c.orderSigner.GetMakerAddress()
	}
	return c.GetAddress()
}

// CreateOrDeriveAPICredentials 创建或衍生 API 凭证
func (c *Client) CreateOrDeriveAPICredentials(ctx context.Context) (*auth.Credentials, error) {
	manager := auth.NewCredentialsManager(c.l1Signer, c.config.Endpoint)
	creds, err := manager.CreateOrDeriveAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	c.SetCredentials(creds)
	return creds, nil
}

// DeriveAPICredentials 衍生 API 凭证
func (c *Client) DeriveAPICredentials(ctx context.Context, nonce int64) (*auth.Credentials, error) {
	creds, err := c.l1Signer.DeriveAPICredentials(ctx, c.config.Endpoint, nonce)
	if err != nil {
		return nil, err
	}

	c.SetCredentials(creds)
	return creds, nil
}

// ensureCredentials 确保有 API 凭证
func (c *Client) ensureCredentials(ctx context.Context) error {
	c.mu.RLock()
	hasCredentials := c.credentials != nil && c.l2Signer != nil
	c.mu.RUnlock()

	if hasCredentials {
		return nil
	}

	_, err := c.CreateOrDeriveAPICredentials(ctx)
	return err
}

// getL2AuthHeaders 获取 L2 认证头
func (c *Client) getL2AuthHeaders(method, path, body string) (map[string]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.l2Signer == nil {
		return nil, fmt.Errorf("no credentials available, call CreateOrDeriveAPICredentials first")
	}

	headers, err := c.l2Signer.GetAuthHeaders(method, path, body)
	if err != nil {
		return nil, err
	}

	return headers.ToMap(), nil
}

// GetL1Signer 获取 L1 签名器
func (c *Client) GetL1Signer() *auth.L1Signer {
	return c.l1Signer
}

// GetOrderSigner 获取订单签名器
func (c *Client) GetOrderSigner() *OrderSigner {
	return c.orderSigner
}

// GetConfig 获取配置
func (c *Client) GetConfig() *Config {
	return c.config
}
