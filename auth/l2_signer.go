package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
// signature = Base64(HMAC-SHA256(secret, timestamp + method + path + body))
// 注意：顺序必须是 timestamp + method + path + body（与 Python SDK 一致）
func (s *L2Signer) Sign(method, path, timestamp, body string) (string, error) {
	message := timestamp + method + path + body

	fmt.Printf(">>> Sign DEBUG:\n")
	fmt.Printf("  Message to sign: %s\n", message)
	fmt.Printf("  Message length: %d\n", len(message))

	// 解码 Base64 编码的 secret
	// Polymarket 使用 URL-safe base64，先尝试 URL-safe 解码，失败则尝试标准解码
	var secretBytes []byte
	var err error
	var decodingMethod string

	// 先尝试 URL-safe base64（带 padding）
	secretBytes, err = base64.URLEncoding.DecodeString(s.credentials.Secret)
	if err != nil {
		// 尝试 URL-safe base64（无 padding）
		secretBytes, err = base64.RawURLEncoding.DecodeString(s.credentials.Secret)
		if err != nil {
			// 最后尝试标准 base64
			secretBytes, err = base64.StdEncoding.DecodeString(s.credentials.Secret)
			if err != nil {
				fmt.Printf("  ERROR: Failed to decode secret with all methods\n")
				return "", err
			}
			decodingMethod = "StdEncoding"
		} else {
			decodingMethod = "RawURLEncoding"
		}
	} else {
		decodingMethod = "URLEncoding"
	}

	fmt.Printf("  Secret decoded using: %s\n", decodingMethod)
	fmt.Printf("  Secret bytes length: %d\n", len(secretBytes))

	// 计算 HMAC-SHA256
	h := hmac.New(sha256.New, secretBytes)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	// 尝试两种编码
	urlSafeSignature := base64.URLEncoding.EncodeToString(signature)
	stdSignature := base64.StdEncoding.EncodeToString(signature)

	fmt.Printf("  Signature (URL-safe): %s\n", urlSafeSignature)
	fmt.Printf("  Signature (Standard): %s\n", stdSignature)
	fmt.Printf("  Using: URL-safe\n")

	// 必须使用 URL-safe Base64 编码（与 Python SDK 一致）
	// 参考：https://github.com/Polymarket/py-clob-client/issues/190
	return base64.URLEncoding.EncodeToString(signature), nil
}

// GetAuthHeaders 获取认证头
func (s *L2Signer) GetAuthHeaders(method, path, body string) (*L2AuthHeaders, error) {
	// Polymarket API 使用秒级时间戳（与 Python SDK 一致）
	// 减去 5 秒以避免时钟偏差导致的认证失败
	// 参考：https://github.com/Polymarket/py-clob-client/issues/190
	timestamp := fmt.Sprintf("%d", common.TimestampSec()-5)

	// 调试日志
	fmt.Printf("\n========== L2 AUTH DEBUG ==========\n")
	fmt.Printf("Method: %s\n", method)
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Timestamp: %s\n", timestamp)
	fmt.Printf("Body length: %d\n", len(body))
	if len(body) < 500 {
		fmt.Printf("Body: %s\n", body)
	} else {
		fmt.Printf("Body (first 500 chars): %s...\n", body[:500])
	}
	fmt.Printf("Address: %s\n", s.address)
	fmt.Printf("API Key: %s\n", s.credentials.APIKey)
	fmt.Printf("Passphrase: %s\n", s.credentials.Passphrase)
	fmt.Printf("Secret: %s\n", s.credentials.Secret)

	signature, err := s.Sign(method, path, timestamp, body)
	if err != nil {
		fmt.Printf("ERROR signing: %v\n", err)
		return nil, err
	}

	fmt.Printf("Signature: %s\n", signature)
	fmt.Printf("===================================\n\n")

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
