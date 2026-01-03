package auth

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// 用于测试的私钥（请勿在生产环境使用）
const testPrivateKeyHex = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewL1Signer(t *testing.T) {
	// Test with valid private key
	signer, err := NewL1Signer(testPrivateKeyHex, 137)
	if err != nil {
		t.Fatalf("NewL1Signer() error: %v", err)
	}
	if signer == nil {
		t.Fatal("NewL1Signer() returned nil")
	}

	// Test with key without 0x prefix
	signer2, err := NewL1Signer(strings.TrimPrefix(testPrivateKeyHex, "0x"), 137)
	if err != nil {
		t.Fatalf("NewL1Signer() without 0x prefix error: %v", err)
	}
	if signer2.GetAddress() != signer.GetAddress() {
		t.Error("Same key with/without 0x prefix should produce same address")
	}

	// Test with invalid private key
	_, err = NewL1Signer("invalid", 137)
	if err == nil {
		t.Error("NewL1Signer() should fail with invalid key")
	}

	// Test with empty key
	_, err = NewL1Signer("", 137)
	if err == nil {
		t.Error("NewL1Signer() should fail with empty key")
	}
}

func TestNewL1SignerFromKey(t *testing.T) {
	// Generate a private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	signer, err := NewL1SignerFromKey(privateKey, 137)
	if err != nil {
		t.Fatalf("NewL1SignerFromKey() error: %v", err)
	}
	if signer == nil {
		t.Fatal("NewL1SignerFromKey() returned nil")
	}

	// Test with nil key
	_, err = NewL1SignerFromKey(nil, 137)
	if err == nil {
		t.Error("NewL1SignerFromKey() should fail with nil key")
	}
}

func TestL1SignerGetAddress(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	address := signer.GetAddress()
	if !strings.HasPrefix(address, "0x") {
		t.Errorf("Address should start with 0x, got %s", address)
	}
	if len(address) != 42 {
		t.Errorf("Address should be 42 chars, got %d", len(address))
	}
	// Address should be lowercase
	if address != strings.ToLower(address) {
		t.Error("GetAddress() should return lowercase address")
	}
}

func TestL1SignerGetAddressChecksum(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	address := signer.GetAddressChecksum()
	if !strings.HasPrefix(address, "0x") {
		t.Errorf("Checksum address should start with 0x, got %s", address)
	}
	if len(address) != 42 {
		t.Errorf("Checksum address should be 42 chars, got %d", len(address))
	}
}

func TestL1SignerSignMessage(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	message := []byte("Hello, World!")
	signature, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("SignMessage() error: %v", err)
	}

	// Signature should be 65 bytes (r: 32, s: 32, v: 1)
	if len(signature) != 65 {
		t.Errorf("Signature should be 65 bytes, got %d", len(signature))
	}

	// v should be 27 or 28
	v := signature[64]
	if v != 27 && v != 28 {
		t.Errorf("v should be 27 or 28, got %d", v)
	}
}

func TestL1SignerSignTypedData(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	typedData := &TypedData{
		Types:       ClobAuthTypes,
		PrimaryType: "ClobAuth",
		Domain:      ClobAuthDomain,
		Message: map[string]interface{}{
			"address":   signer.GetAddress(),
			"timestamp": "1234567890",
			"nonce":     big.NewInt(0),
			"message":   ClobAuthMessage,
		},
	}

	signature, err := signer.SignTypedData(typedData)
	if err != nil {
		t.Fatalf("SignTypedData() error: %v", err)
	}

	// Signature should be 65 bytes
	if len(signature) != 65 {
		t.Errorf("Signature should be 65 bytes, got %d", len(signature))
	}

	// v should be 27 or 28
	v := signature[64]
	if v != 27 && v != 28 {
		t.Errorf("v should be 27 or 28, got %d", v)
	}
}

func TestL1SignerSignTypedDataWithVerifyingContract(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	domain := PolymarketExchangeDomain(137, "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e")

	// go-ethereum EIP-712 expects addresses as checksummed hex strings
	typedData := &TypedData{
		Types:       OrderTypes,
		PrimaryType: "Order",
		Domain:      domain,
		Message: map[string]interface{}{
			"salt":          big.NewInt(12345),
			"maker":         signer.wallet.Address.Hex(),
			"signer":        signer.wallet.Address.Hex(),
			"taker":         signer.wallet.Address.Hex(),
			"tokenId":       big.NewInt(100),
			"makerAmount":   big.NewInt(1000000),
			"takerAmount":   big.NewInt(500000),
			"expiration":    big.NewInt(0),
			"nonce":         big.NewInt(0),
			"feeRateBps":    big.NewInt(0),
			"side":          big.NewInt(0),
			"signatureType": big.NewInt(0),
		},
	}

	signature, err := signer.SignTypedData(typedData)
	if err != nil {
		t.Fatalf("SignTypedData() with verifyingContract error: %v", err)
	}

	if len(signature) != 65 {
		t.Errorf("Signature should be 65 bytes, got %d", len(signature))
	}
}

func TestL1SignerSignClobAuth(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	headers, err := signer.SignClobAuth("1234567890", 0)
	if err != nil {
		t.Fatalf("SignClobAuth() error: %v", err)
	}

	if headers.Address != signer.GetAddress() {
		t.Errorf("Address = %s, expected %s", headers.Address, signer.GetAddress())
	}
	if headers.Timestamp != "1234567890" {
		t.Errorf("Timestamp = %s, expected 1234567890", headers.Timestamp)
	}
	if headers.Nonce != "0" {
		t.Errorf("Nonce = %s, expected 0", headers.Nonce)
	}
	if !strings.HasPrefix(headers.Signature, "0x") {
		t.Errorf("Signature should start with 0x, got %s", headers.Signature)
	}
}

func TestL1SignerSignOrder(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	order := &OrderPayload{
		Salt:          "12345",
		Maker:         signer.GetAddress(),
		Signer:        signer.GetAddress(),
		Taker:         "0x0000000000000000000000000000000000000000",
		TokenID:       "100",
		MakerAmount:   "1000000",
		TakerAmount:   "500000",
		Expiration:    "0",
		Nonce:         "0",
		FeeRateBps:    "0",
		Side:          0,
		SignatureType: 0,
		IsNegRisk:     false,
	}

	exchangeAddr := "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e"
	signature, err := signer.SignOrder(order, exchangeAddr)
	if err != nil {
		t.Fatalf("SignOrder() error: %v", err)
	}

	if !strings.HasPrefix(signature, "0x") {
		t.Errorf("Signature should start with 0x, got %s", signature)
	}
}

func TestL1SignerSignOrderInvalidParams(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)
	exchangeAddr := "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e"

	tests := []struct {
		name  string
		order *OrderPayload
	}{
		{
			name: "invalid salt",
			order: &OrderPayload{
				Salt: "invalid", Maker: signer.GetAddress(), Signer: signer.GetAddress(),
				Taker: "0x0000000000000000000000000000000000000000", TokenID: "100",
				MakerAmount: "1000000", TakerAmount: "500000", Expiration: "0",
				Nonce: "0", FeeRateBps: "0", Side: 0, SignatureType: 0,
			},
		},
		{
			name: "invalid tokenID",
			order: &OrderPayload{
				Salt: "12345", Maker: signer.GetAddress(), Signer: signer.GetAddress(),
				Taker: "0x0000000000000000000000000000000000000000", TokenID: "invalid",
				MakerAmount: "1000000", TakerAmount: "500000", Expiration: "0",
				Nonce: "0", FeeRateBps: "0", Side: 0, SignatureType: 0,
			},
		},
		{
			name: "invalid makerAmount",
			order: &OrderPayload{
				Salt: "12345", Maker: signer.GetAddress(), Signer: signer.GetAddress(),
				Taker: "0x0000000000000000000000000000000000000000", TokenID: "100",
				MakerAmount: "invalid", TakerAmount: "500000", Expiration: "0",
				Nonce: "0", FeeRateBps: "0", Side: 0, SignatureType: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := signer.SignOrder(tt.order, exchangeAddr)
			if err == nil {
				t.Errorf("%s: expected error but got none", tt.name)
			}
		})
	}
}

func TestL1SignerGetChainID(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)
	if signer.GetChainID() != 137 {
		t.Errorf("GetChainID() = %d, expected 137", signer.GetChainID())
	}

	signer2, _ := NewL1Signer(testPrivateKeyHex, 1)
	if signer2.GetChainID() != 1 {
		t.Errorf("GetChainID() = %d, expected 1", signer2.GetChainID())
	}
}

func TestMarshalUnmarshalCredentials(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     "dGVzdC1zZWNyZXQ=",
		Passphrase: "test-passphrase",
	}

	// Marshal
	data, err := MarshalCredentials(creds)
	if err != nil {
		t.Fatalf("MarshalCredentials() error: %v", err)
	}

	// Verify JSON structure
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if m["apiKey"] != creds.APIKey {
		t.Error("apiKey mismatch in JSON")
	}

	// Unmarshal
	creds2, err := UnmarshalCredentials(data)
	if err != nil {
		t.Fatalf("UnmarshalCredentials() error: %v", err)
	}

	if creds2.APIKey != creds.APIKey {
		t.Error("APIKey mismatch after unmarshal")
	}
	if creds2.Secret != creds.Secret {
		t.Error("Secret mismatch after unmarshal")
	}
	if creds2.Passphrase != creds.Passphrase {
		t.Error("Passphrase mismatch after unmarshal")
	}
}

func TestUnmarshalCredentialsInvalid(t *testing.T) {
	_, err := UnmarshalCredentials([]byte("invalid json"))
	if err == nil {
		t.Error("UnmarshalCredentials() should fail with invalid JSON")
	}
}

func TestL1SignerDeterministicSignature(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	message := []byte("Test message for deterministic signing")

	sig1, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("First SignMessage() error: %v", err)
	}

	sig2, err := signer.SignMessage(message)
	if err != nil {
		t.Fatalf("Second SignMessage() error: %v", err)
	}

	// go-ethereum crypto.Sign is deterministic
	if string(sig1) != string(sig2) {
		t.Error("Signatures should be deterministic for the same message")
	}
}

func TestL1SignerWallet(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKeyHex, 137)

	// Verify wallet is properly initialized
	if signer.wallet == nil {
		t.Fatal("Wallet should not be nil")
	}
	if signer.wallet.PrivateKey == nil {
		t.Fatal("Private key should not be nil")
	}
	if signer.wallet.Address == ([20]byte{}) {
		t.Fatal("Address should not be empty")
	}

	// Verify private key can derive public key
	publicKey := signer.wallet.PrivateKey.Public()
	if publicKey == nil {
		t.Fatal("Public key should not be nil")
	}
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Failed to cast to ECDSA public key")
	}

	// Verify address matches
	derivedAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	if derivedAddress != signer.wallet.Address {
		t.Error("Derived address does not match wallet address")
	}
}
