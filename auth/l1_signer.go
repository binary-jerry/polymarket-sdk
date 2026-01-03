package auth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	pmcommon "github.com/binary-jerry/polymarket-sdk/common"
)

// L1Signer L1 EIP-712 签名器
type L1Signer struct {
	wallet  *Wallet
	chainID int
}

// NewL1Signer 创建 L1 签名器
func NewL1Signer(privateKeyHex string, chainID int) (*L1Signer, error) {
	// 移除 0x 前缀
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &L1Signer{
		wallet: &Wallet{
			Address:    address,
			PrivateKey: privateKey,
		},
		chainID: chainID,
	}, nil
}

// NewL1SignerFromKey 从私钥创建 L1 签名器
func NewL1SignerFromKey(privateKey *ecdsa.PrivateKey, chainID int) (*L1Signer, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &L1Signer{
		wallet: &Wallet{
			Address:    address,
			PrivateKey: privateKey,
		},
		chainID: chainID,
	}, nil
}

// GetAddress 获取钱包地址
func (s *L1Signer) GetAddress() string {
	return strings.ToLower(s.wallet.Address.Hex())
}

// GetAddressChecksum 获取钱包地址（校验和格式）
func (s *L1Signer) GetAddressChecksum() string {
	return s.wallet.Address.Hex()
}

// SignMessage 签名消息
func (s *L1Signer) SignMessage(message []byte) ([]byte, error) {
	// 添加以太坊签名前缀
	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixedMessage))

	signature, err := crypto.Sign(hash.Bytes(), s.wallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// 调整 v 值
	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

// SignTypedData 签名 EIP-712 类型数据
func (s *L1Signer) SignTypedData(typedData *TypedData) ([]byte, error) {
	// 转换为 go-ethereum 的类型
	types := make(apitypes.Types)
	for name, fields := range typedData.Types {
		apiFields := make([]apitypes.Type, len(fields))
		for i, f := range fields {
			apiFields[i] = apitypes.Type{Name: f.Name, Type: f.Type}
		}
		types[name] = apiFields
	}

	// 添加 EIP712Domain 类型
	types["EIP712Domain"] = []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
	}

	if typedData.Domain.VerifyingContract != "" {
		types["EIP712Domain"] = append(types["EIP712Domain"],
			apitypes.Type{Name: "verifyingContract", Type: "address"})
	}

	domain := apitypes.TypedDataDomain{
		Name:    typedData.Domain.Name,
		Version: typedData.Domain.Version,
		ChainId: (*math.HexOrDecimal256)(typedData.Domain.ChainId),
	}

	if typedData.Domain.VerifyingContract != "" {
		domain.VerifyingContract = typedData.Domain.VerifyingContract
	}

	apiTypedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: typedData.PrimaryType,
		Domain:      domain,
		Message:     typedData.Message,
	}

	// 计算 hash
	domainSeparator, err := apiTypedData.HashStruct("EIP712Domain", apiTypedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain: %w", err)
	}

	messageHash, err := apiTypedData.HashStruct(apiTypedData.PrimaryType, apiTypedData.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message: %w", err)
	}

	// EIP-712 hash: keccak256("\x19\x01" + domainSeparator + messageHash)
	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, messageHash...)
	hash := crypto.Keccak256Hash(rawData)

	// 签名
	signature, err := crypto.Sign(hash.Bytes(), s.wallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign typed data: %w", err)
	}

	// 调整 v 值
	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

// SignClobAuth 签名 CLOB 认证消息
func (s *L1Signer) SignClobAuth(timestamp string, nonce int64) (*L1AuthHeaders, error) {
	typedData := &TypedData{
		Types:       ClobAuthTypes,
		PrimaryType: "ClobAuth",
		Domain:      ClobAuthDomain,
		Message: map[string]interface{}{
			"address":   s.GetAddress(),
			"timestamp": timestamp,
			"nonce":     big.NewInt(nonce),
			"message":   ClobAuthMessage,
		},
	}

	signature, err := s.SignTypedData(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CLOB auth: %w", err)
	}

	return &L1AuthHeaders{
		Address:   s.GetAddress(),
		Signature: hexutil.Encode(signature),
		Timestamp: timestamp,
		Nonce:     fmt.Sprintf("%d", nonce),
	}, nil
}

// CreateAPICredentials 创建 API 凭证
func (s *L1Signer) CreateAPICredentials(ctx context.Context, clobEndpoint string) (*Credentials, error) {
	timestamp := pmcommon.TimestampSecStr()
	nonce := int64(0)

	headers, err := s.SignClobAuth(timestamp, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CLOB auth: %w", err)
	}

	httpClient := pmcommon.NewHTTPClient(&pmcommon.HTTPClientConfig{
		BaseURL: clobEndpoint,
	})

	var result CreateAPIKeyResponse
	err = httpClient.DoWithAuth(ctx, "POST", "/auth/api-key", nil, headers.ToMap(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &Credentials{
		APIKey:     result.APIKey,
		Secret:     result.Secret,
		Passphrase: result.Passphrase,
	}, nil
}

// DeriveAPICredentials 衍生 API 凭证（确定性）
func (s *L1Signer) DeriveAPICredentials(ctx context.Context, clobEndpoint string, nonce int64) (*Credentials, error) {
	timestamp := pmcommon.TimestampSecStr()

	headers, err := s.SignClobAuth(timestamp, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CLOB auth: %w", err)
	}

	httpClient := pmcommon.NewHTTPClient(&pmcommon.HTTPClientConfig{
		BaseURL: clobEndpoint,
	})

	body := map[string]interface{}{
		"nonce": nonce,
	}

	var result DeriveAPIKeyResponse
	err = httpClient.DoWithAuth(ctx, "POST", "/auth/derive-api-key", body, headers.ToMap(), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to derive API key: %w", err)
	}

	return &Credentials{
		APIKey:     result.APIKey,
		Secret:     result.Secret,
		Passphrase: result.Passphrase,
	}, nil
}

// SignOrder 签名订单
func (s *L1Signer) SignOrder(order *OrderPayload, exchangeAddress string) (string, error) {
	salt, ok := new(big.Int).SetString(order.Salt, 10)
	if !ok {
		return "", fmt.Errorf("invalid salt: %s", order.Salt)
	}

	tokenID, ok := new(big.Int).SetString(order.TokenID, 10)
	if !ok {
		return "", fmt.Errorf("invalid token ID: %s", order.TokenID)
	}

	makerAmount, ok := new(big.Int).SetString(order.MakerAmount, 10)
	if !ok {
		return "", fmt.Errorf("invalid maker amount: %s", order.MakerAmount)
	}

	takerAmount, ok := new(big.Int).SetString(order.TakerAmount, 10)
	if !ok {
		return "", fmt.Errorf("invalid taker amount: %s", order.TakerAmount)
	}

	expiration, ok := new(big.Int).SetString(order.Expiration, 10)
	if !ok {
		return "", fmt.Errorf("invalid expiration: %s", order.Expiration)
	}

	nonce, ok := new(big.Int).SetString(order.Nonce, 10)
	if !ok {
		return "", fmt.Errorf("invalid nonce: %s", order.Nonce)
	}

	feeRateBps, ok := new(big.Int).SetString(order.FeeRateBps, 10)
	if !ok {
		return "", fmt.Errorf("invalid fee rate: %s", order.FeeRateBps)
	}

	domain := PolymarketExchangeDomain(s.chainID, exchangeAddress)

	// go-ethereum EIP-712 expects addresses as checksummed hex strings
	makerAddr := common.HexToAddress(order.Maker).Hex()
	signerAddr := common.HexToAddress(order.Signer).Hex()
	takerAddr := common.HexToAddress(order.Taker).Hex()

	typedData := &TypedData{
		Types:       OrderTypes,
		PrimaryType: "Order",
		Domain:      domain,
		Message: map[string]interface{}{
			"salt":          salt,
			"maker":         makerAddr,
			"signer":        signerAddr,
			"taker":         takerAddr,
			"tokenId":       tokenID,
			"makerAmount":   makerAmount,
			"takerAmount":   takerAmount,
			"expiration":    expiration,
			"nonce":         nonce,
			"feeRateBps":    feeRateBps,
			"side":          big.NewInt(int64(order.Side)),
			"signatureType": big.NewInt(int64(order.SignatureType)),
		},
	}

	signature, err := s.SignTypedData(typedData)
	if err != nil {
		return "", fmt.Errorf("failed to sign order: %w", err)
	}

	return hexutil.Encode(signature), nil
}

// GetChainID 获取链 ID
func (s *L1Signer) GetChainID() int {
	return s.chainID
}

// MarshalCredentials 序列化凭证
func MarshalCredentials(creds *Credentials) ([]byte, error) {
	return json.Marshal(creds)
}

// UnmarshalCredentials 反序列化凭证
func UnmarshalCredentials(data []byte) (*Credentials, error) {
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return &creds, nil
}
