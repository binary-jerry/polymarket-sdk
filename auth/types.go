package auth

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Credentials API 凭证
type Credentials struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`     // Base64 编码
	Passphrase string `json:"passphrase"`
}

// Wallet 钱包信息
type Wallet struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// L1AuthHeaders L1 认证请求头
type L1AuthHeaders struct {
	Address   string // POLY_ADDRESS
	Signature string // POLY_SIGNATURE
	Timestamp string // POLY_TIMESTAMP
	Nonce     string // POLY_NONCE
}

// ToMap 转换为 map
func (h *L1AuthHeaders) ToMap() map[string]string {
	return map[string]string{
		"POLY_ADDRESS":   h.Address,
		"POLY_SIGNATURE": h.Signature,
		"POLY_TIMESTAMP": h.Timestamp,
		"POLY_NONCE":     h.Nonce,
	}
}

// L2AuthHeaders L2 认证请求头
type L2AuthHeaders struct {
	Address    string // POLY_ADDRESS
	APIKey     string // POLY_API_KEY
	Passphrase string // POLY_PASSPHRASE
	Timestamp  string // POLY_TIMESTAMP
	Signature  string // POLY_SIGNATURE
}

// ToMap 转换为 map
func (h *L2AuthHeaders) ToMap() map[string]string {
	return map[string]string{
		"POLY_ADDRESS":    h.Address,
		"POLY_API_KEY":    h.APIKey,
		"POLY_PASSPHRASE": h.Passphrase,
		"POLY_TIMESTAMP":  h.Timestamp,
		"POLY_SIGNATURE":  h.Signature,
	}
}

// SignatureType 签名类型
type SignatureType int

const (
	// SignatureTypeEOA EOA 钱包签名
	SignatureTypeEOA SignatureType = 0
	// SignatureTypePolyProxy Polymarket 代理签名 (Magic/Email 钱包)
	SignatureTypePolyProxy SignatureType = 1
	// SignatureTypePolyGnosisSafe Gnosis Safe 签名
	SignatureTypePolyGnosisSafe SignatureType = 2
)

// TypedDataField EIP-712 类型数据字段
type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// TypedDataDomain EIP-712 域
type TypedDataDomain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainId           *big.Int `json:"chainId"`
	VerifyingContract string `json:"verifyingContract,omitempty"`
}

// TypedData EIP-712 类型数据
type TypedData struct {
	Types       map[string][]TypedDataField `json:"types"`
	PrimaryType string                      `json:"primaryType"`
	Domain      TypedDataDomain             `json:"domain"`
	Message     map[string]interface{}      `json:"message"`
}

// ClobAuthDomain CLOB 认证 EIP-712 域
var ClobAuthDomain = TypedDataDomain{
	Name:    "ClobAuthDomain",
	Version: "1",
	ChainId: big.NewInt(137),
}

// ClobAuthTypes CLOB 认证类型定义
var ClobAuthTypes = map[string][]TypedDataField{
	"ClobAuth": {
		{Name: "address", Type: "address"},
		{Name: "timestamp", Type: "string"},
		{Name: "nonce", Type: "uint256"},
		{Name: "message", Type: "string"},
	},
}

// ClobAuthMessage CLOB 认证消息内容
const ClobAuthMessage = "This message attests that I control the given wallet"

// PolymarketExchangeDomain Polymarket 交易所 EIP-712 域
func PolymarketExchangeDomain(chainID int, exchangeAddress string) TypedDataDomain {
	return TypedDataDomain{
		Name:              "Polymarket CTF Exchange",
		Version:           "1",
		ChainId:           big.NewInt(int64(chainID)),
		VerifyingContract: exchangeAddress,
	}
}

// OrderTypes 订单 EIP-712 类型定义
var OrderTypes = map[string][]TypedDataField{
	"Order": {
		{Name: "salt", Type: "uint256"},
		{Name: "maker", Type: "address"},
		{Name: "signer", Type: "address"},
		{Name: "taker", Type: "address"},
		{Name: "tokenId", Type: "uint256"},
		{Name: "makerAmount", Type: "uint256"},
		{Name: "takerAmount", Type: "uint256"},
		{Name: "expiration", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "feeRateBps", Type: "uint256"},
		{Name: "side", Type: "uint8"},
		{Name: "signatureType", Type: "uint8"},
	},
}

// CreateAPIKeyResponse 创建 API Key 响应
type CreateAPIKeyResponse struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}

// DeriveAPIKeyResponse 衍生 API Key 响应
type DeriveAPIKeyResponse struct {
	APIKey     string `json:"apiKey"`
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
}
