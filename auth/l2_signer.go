package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/binary-jerry/polymarket-sdk/common"
)

// L2Signer L2 HMAC 签名器
type L2Signer struct {
	credentials *Credentials
	address     string
}

// NewL2Signer 创建 L2 签名器
func NewL2Signer(address string, creds *Credentials) *L2Signer {
	return &L2Signer{
		credentials: creds,
		address:     address,
	}
}

// Sign 签名请求
// signature = Base64(HMAC-SHA256(secret, method + path + timestamp + body))
func (s *L2Signer) Sign(method, path, timestamp, body string) (string, error) {
	message := method + path + timestamp + body

	// 解码 Base64 编码的 secret
	secretBytes, err := base64.StdEncoding.DecodeString(s.credentials.Secret)
	if err != nil {
		return "", err
	}

	// 计算 HMAC-SHA256
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	// 返回 Base64 编码的签名
	return base64.StdEncoding.EncodeToString(signature), nil
}

// GetAuthHeaders 获取认证头
func (s *L2Signer) GetAuthHeaders(method, path, body string) (*L2AuthHeaders, error) {
	timestamp := common.TimestampMsStr()

	signature, err := s.Sign(method, path, timestamp, body)
	if err != nil {
		return nil, err
	}

	return &L2AuthHeaders{
		Address:    s.address,
		APIKey:     s.credentials.APIKey,
		Passphrase: s.credentials.Passphrase,
		Timestamp:  timestamp,
		Signature:  signature,
	}, nil
}

// SignRequest 为 HTTP 请求添加认证头
func (s *L2Signer) SignRequest(req *http.Request, body string) error {
	// 获取请求路径（不包含 host）
	path := req.URL.Path
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}

	headers, err := s.GetAuthHeaders(req.Method, path, body)
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("POLY_ADDRESS", headers.Address)
	req.Header.Set("POLY_API_KEY", headers.APIKey)
	req.Header.Set("POLY_PASSPHRASE", headers.Passphrase)
	req.Header.Set("POLY_TIMESTAMP", headers.Timestamp)
	req.Header.Set("POLY_SIGNATURE", headers.Signature)

	return nil
}

// GetCredentials 获取凭证
func (s *L2Signer) GetCredentials() *Credentials {
	return s.credentials
}

// GetAddress 获取地址
func (s *L2Signer) GetAddress() string {
	return s.address
}

// UpdateCredentials 更新凭证
func (s *L2Signer) UpdateCredentials(creds *Credentials) {
	s.credentials = creds
}

// IsValid 检查签名器是否有效
func (s *L2Signer) IsValid() bool {
	return s.credentials != nil &&
		s.credentials.APIKey != "" &&
		s.credentials.Secret != "" &&
		s.credentials.Passphrase != "" &&
		s.address != ""
}
