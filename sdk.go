package polymarket

import (
	"context"
	"fmt"

	"github.com/binary-jerry/polymarket-sdk/auth"
	"github.com/binary-jerry/polymarket-sdk/clob"
	"github.com/binary-jerry/polymarket-sdk/gamma"
	"github.com/binary-jerry/polymarket-sdk/orderbook"
)

// SDK Polymarket 统一 SDK
type SDK struct {
	config *Config

	// 公开模块
	OrderBook *orderbook.SDK // 订单簿 (WebSocket)
	Markets   *gamma.Client  // 市场查询 (Gamma API)
	Trading   *clob.Client   // 交易 (CLOB API)

	// 内部
	l1Signer *auth.L1Signer
}

// NewSDK 创建完整 SDK 实例（需要私钥）
func NewSDK(config *Config, privateKey string) (*SDK, error) {
	if config == nil {
		config = DefaultConfig()
	}
	config.Validate()

	// 创建 L1 签名器
	l1Signer, err := auth.NewL1Signer(privateKey, ChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// 创建 OrderBook SDK
	obConfig := &orderbook.Config{
		WSEndpoint:           config.WSEndpoint,
		MaxTokensPerConn:     config.MaxTokensPerConn,
		ReconnectMinInterval: config.ReconnectMinInterval,
		ReconnectMaxInterval: config.ReconnectMaxInterval,
		ReconnectMaxAttempts: config.ReconnectMaxAttempts,
		PingInterval:         config.PingInterval,
		PongTimeout:          config.PongTimeout,
		MessageBufferSize:    config.MessageBufferSize,
		UpdateChannelSize:    config.UpdateChannelSize,
	}
	obSDK := orderbook.NewSDK(obConfig)

	// 创建 Gamma 客户端
	gammaConfig := &gamma.Config{
		Endpoint:     config.GammaEndpoint,
		Timeout:      config.HTTPTimeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	}
	gammaClient := gamma.NewClient(gammaConfig)

	// 创建 CLOB 客户端
	clobConfig := &clob.Config{
		Endpoint:               config.CLOBEndpoint,
		ChainID:                ChainID,
		Timeout:                config.HTTPTimeout,
		MaxRetries:             config.MaxRetries,
		RetryDelayMs:           config.RetryDelayMs,
		ExchangeAddress:        config.CTFExchangeAddress,
		NegRiskExchangeAddress: config.NegRiskCTFExchangeAddress,
		NegRiskAdapterAddress:  config.NegRiskAdapterAddress,
		CollateralAddress:      config.CollateralAddress,
	}
	clobClient, err := clob.NewClient(clobConfig, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CLOB client: %w", err)
	}

	return &SDK{
		config:    config,
		OrderBook: obSDK,
		Markets:   gammaClient,
		Trading:   clobClient,
		l1Signer:  l1Signer,
	}, nil
}

// NewPublicSDK 创建仅公开接口的 SDK（无需私钥）
func NewPublicSDK(config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}
	config.Validate()

	// 创建 OrderBook SDK
	obConfig := &orderbook.Config{
		WSEndpoint:           config.WSEndpoint,
		MaxTokensPerConn:     config.MaxTokensPerConn,
		ReconnectMinInterval: config.ReconnectMinInterval,
		ReconnectMaxInterval: config.ReconnectMaxInterval,
		ReconnectMaxAttempts: config.ReconnectMaxAttempts,
		PingInterval:         config.PingInterval,
		PongTimeout:          config.PongTimeout,
		MessageBufferSize:    config.MessageBufferSize,
		UpdateChannelSize:    config.UpdateChannelSize,
	}
	obSDK := orderbook.NewSDK(obConfig)

	// 创建 Gamma 客户端
	gammaConfig := &gamma.Config{
		Endpoint:     config.GammaEndpoint,
		Timeout:      config.HTTPTimeout,
		MaxRetries:   config.MaxRetries,
		RetryDelayMs: config.RetryDelayMs,
	}
	gammaClient := gamma.NewClient(gammaConfig)

	return &SDK{
		config:    config,
		OrderBook: obSDK,
		Markets:   gammaClient,
	}
}

// NewTradingSDK 创建带交易功能的 SDK（需要私钥和凭证）
func NewTradingSDK(config *Config, privateKey string, creds *auth.Credentials) (*SDK, error) {
	sdk, err := NewSDK(config, privateKey)
	if err != nil {
		return nil, err
	}

	if creds != nil {
		sdk.Trading.SetCredentials(creds)
	}

	return sdk, nil
}

// Close 关闭 SDK
func (s *SDK) Close() {
	if s.OrderBook != nil {
		s.OrderBook.Close()
	}
	if s.Markets != nil {
		s.Markets.Close()
	}
	if s.Trading != nil {
		s.Trading.Close()
	}
}

// GetAddress 获取钱包地址
func (s *SDK) GetAddress() string {
	if s.l1Signer != nil {
		return s.l1Signer.GetAddress()
	}
	return ""
}

// GetConfig 获取配置
func (s *SDK) GetConfig() *Config {
	return s.config
}

// CreateOrDeriveAPICredentials 创建或衍生 API 凭证
func (s *SDK) CreateOrDeriveAPICredentials(ctx context.Context) (*auth.Credentials, error) {
	if s.Trading == nil {
		return nil, fmt.Errorf("trading client not initialized, use NewSDK with private key")
	}
	return s.Trading.CreateOrDeriveAPICredentials(ctx)
}

// GetCredentials 获取当前凭证
func (s *SDK) GetCredentials() *auth.Credentials {
	if s.Trading == nil {
		return nil
	}
	return s.Trading.GetCredentials()
}

// SetCredentials 设置凭证
func (s *SDK) SetCredentials(creds *auth.Credentials) {
	if s.Trading != nil {
		s.Trading.SetCredentials(creds)
	}
}

// SetCredentialsWithAddress 设置凭证（指定账户地址）
func (s *SDK) SetCredentialsWithAddress(creds *auth.Credentials, address string) {
	if s.Trading != nil {
		s.Trading.SetCredentialsWithAddress(creds, address)
	}
}

// SetFunderAddress 设置代理钱包地址（用于代理钱包模式）
// funderAddress: 代理钱包地址（持有资金的地址）
func (s *SDK) SetFunderAddress(funderAddress string) {
	if s.Trading != nil {
		s.Trading.SetFunderAddress(funderAddress)
	}
}

// SetSignatureType 设置签名类型
// signatureType: 0=EOA, 1=POLY_PROXY, 2=GNOSIS_SAFE
func (s *SDK) SetSignatureType(signatureType int) {
	if s.Trading != nil {
		s.Trading.SetSignatureType(signatureType)
	}
}

// IsTradingEnabled 是否启用交易功能
func (s *SDK) IsTradingEnabled() bool {
	return s.Trading != nil && s.l1Signer != nil
}
