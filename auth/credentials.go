package auth

import (
	"context"
	"fmt"
)

// CredentialsManager 凭证管理器
type CredentialsManager struct {
	l1Signer     *L1Signer
	clobEndpoint string
	credentials  *Credentials
}

// NewCredentialsManager 创建凭证管理器
func NewCredentialsManager(l1Signer *L1Signer, clobEndpoint string) *CredentialsManager {
	return &CredentialsManager{
		l1Signer:     l1Signer,
		clobEndpoint: clobEndpoint,
	}
}

// CreateOrDeriveAPIKeys 创建或衍生 API 密钥
// 优先尝试衍生（确定性），失败则创建新的
func (m *CredentialsManager) CreateOrDeriveAPIKeys(ctx context.Context) (*Credentials, error) {
	// 先尝试衍生（使用 nonce=0）
	creds, err := m.l1Signer.DeriveAPICredentials(ctx, m.clobEndpoint, 0)
	if err == nil {
		m.credentials = creds
		return creds, nil
	}

	// 衍生失败，尝试创建新的
	creds, err = m.l1Signer.CreateAPICredentials(ctx, m.clobEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create API credentials: %w", err)
	}

	m.credentials = creds
	return creds, nil
}

// DeriveAPIKey 衍生 API 密钥（确定性）
func (m *CredentialsManager) DeriveAPIKey(ctx context.Context, nonce int64) (*Credentials, error) {
	creds, err := m.l1Signer.DeriveAPICredentials(ctx, m.clobEndpoint, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to derive API key: %w", err)
	}

	m.credentials = creds
	return creds, nil
}

// CreateAPIKey 创建新的 API 密钥
func (m *CredentialsManager) CreateAPIKey(ctx context.Context) (*Credentials, error) {
	creds, err := m.l1Signer.CreateAPICredentials(ctx, m.clobEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	m.credentials = creds
	return creds, nil
}

// GetCredentials 获取当前凭证
func (m *CredentialsManager) GetCredentials() *Credentials {
	return m.credentials
}

// SetCredentials 设置凭证
func (m *CredentialsManager) SetCredentials(creds *Credentials) {
	m.credentials = creds
}

// HasCredentials 检查是否有凭证
func (m *CredentialsManager) HasCredentials() bool {
	return m.credentials != nil &&
		m.credentials.APIKey != "" &&
		m.credentials.Secret != "" &&
		m.credentials.Passphrase != ""
}

// GetL2Signer 获取 L2 签名器
func (m *CredentialsManager) GetL2Signer() (*L2Signer, error) {
	if !m.HasCredentials() {
		return nil, fmt.Errorf("no credentials available")
	}

	return NewL2Signer(m.l1Signer.GetAddress(), m.credentials), nil
}

// GetL1Signer 获取 L1 签名器
func (m *CredentialsManager) GetL1Signer() *L1Signer {
	return m.l1Signer
}

// GetAddress 获取钱包地址
func (m *CredentialsManager) GetAddress() string {
	return m.l1Signer.GetAddress()
}

// ValidateCredentials 验证凭证是否有效
func ValidateCredentials(creds *Credentials) error {
	if creds == nil {
		return fmt.Errorf("credentials is nil")
	}
	if creds.APIKey == "" {
		return fmt.Errorf("API key is empty")
	}
	if creds.Secret == "" {
		return fmt.Errorf("secret is empty")
	}
	if creds.Passphrase == "" {
		return fmt.Errorf("passphrase is empty")
	}
	return nil
}
