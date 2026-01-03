package auth

import (
	"math/big"
	"testing"
)

func TestL1AuthHeadersToMap(t *testing.T) {
	headers := &L1AuthHeaders{
		Address:   "0x1234567890123456789012345678901234567890",
		Signature: "0xabcdef",
		Timestamp: "1234567890",
		Nonce:     "1",
	}

	m := headers.ToMap()

	if m["POLY_ADDRESS"] != headers.Address {
		t.Errorf("POLY_ADDRESS = %s, expected %s", m["POLY_ADDRESS"], headers.Address)
	}
	if m["POLY_SIGNATURE"] != headers.Signature {
		t.Errorf("POLY_SIGNATURE = %s, expected %s", m["POLY_SIGNATURE"], headers.Signature)
	}
	if m["POLY_TIMESTAMP"] != headers.Timestamp {
		t.Errorf("POLY_TIMESTAMP = %s, expected %s", m["POLY_TIMESTAMP"], headers.Timestamp)
	}
	if m["POLY_NONCE"] != headers.Nonce {
		t.Errorf("POLY_NONCE = %s, expected %s", m["POLY_NONCE"], headers.Nonce)
	}
}

func TestL2AuthHeadersToMap(t *testing.T) {
	headers := &L2AuthHeaders{
		Address:    "0x1234567890123456789012345678901234567890",
		APIKey:     "api-key-123",
		Passphrase: "passphrase-456",
		Timestamp:  "1234567890123",
		Signature:  "signature-789",
	}

	m := headers.ToMap()

	if m["POLY_ADDRESS"] != headers.Address {
		t.Errorf("POLY_ADDRESS = %s, expected %s", m["POLY_ADDRESS"], headers.Address)
	}
	if m["POLY_API_KEY"] != headers.APIKey {
		t.Errorf("POLY_API_KEY = %s, expected %s", m["POLY_API_KEY"], headers.APIKey)
	}
	if m["POLY_PASSPHRASE"] != headers.Passphrase {
		t.Errorf("POLY_PASSPHRASE = %s, expected %s", m["POLY_PASSPHRASE"], headers.Passphrase)
	}
	if m["POLY_TIMESTAMP"] != headers.Timestamp {
		t.Errorf("POLY_TIMESTAMP = %s, expected %s", m["POLY_TIMESTAMP"], headers.Timestamp)
	}
	if m["POLY_SIGNATURE"] != headers.Signature {
		t.Errorf("POLY_SIGNATURE = %s, expected %s", m["POLY_SIGNATURE"], headers.Signature)
	}
}

func TestSignatureTypeConstants(t *testing.T) {
	if SignatureTypeEOA != 0 {
		t.Errorf("SignatureTypeEOA = %d, expected 0", SignatureTypeEOA)
	}
	if SignatureTypePolyProxy != 1 {
		t.Errorf("SignatureTypePolyProxy = %d, expected 1", SignatureTypePolyProxy)
	}
	if SignatureTypePolyGnosisSafe != 2 {
		t.Errorf("SignatureTypePolyGnosisSafe = %d, expected 2", SignatureTypePolyGnosisSafe)
	}
}

func TestClobAuthDomain(t *testing.T) {
	if ClobAuthDomain.Name != "ClobAuthDomain" {
		t.Errorf("ClobAuthDomain.Name = %s, expected ClobAuthDomain", ClobAuthDomain.Name)
	}
	if ClobAuthDomain.Version != "1" {
		t.Errorf("ClobAuthDomain.Version = %s, expected 1", ClobAuthDomain.Version)
	}
	if ClobAuthDomain.ChainId.Cmp(big.NewInt(137)) != 0 {
		t.Errorf("ClobAuthDomain.ChainId = %s, expected 137", ClobAuthDomain.ChainId.String())
	}
}

func TestClobAuthTypes(t *testing.T) {
	fields, ok := ClobAuthTypes["ClobAuth"]
	if !ok {
		t.Fatal("ClobAuthTypes should have ClobAuth key")
	}

	expectedFields := map[string]string{
		"address":   "address",
		"timestamp": "string",
		"nonce":     "uint256",
		"message":   "string",
	}

	if len(fields) != len(expectedFields) {
		t.Errorf("ClobAuth has %d fields, expected %d", len(fields), len(expectedFields))
	}

	for _, field := range fields {
		expectedType, ok := expectedFields[field.Name]
		if !ok {
			t.Errorf("Unexpected field: %s", field.Name)
			continue
		}
		if field.Type != expectedType {
			t.Errorf("Field %s has type %s, expected %s", field.Name, field.Type, expectedType)
		}
	}
}

func TestClobAuthMessage(t *testing.T) {
	expected := "This message attests that I control the given wallet"
	if ClobAuthMessage != expected {
		t.Errorf("ClobAuthMessage = %s, expected %s", ClobAuthMessage, expected)
	}
}

func TestPolymarketExchangeDomain(t *testing.T) {
	chainID := 137
	exchangeAddr := "0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e"

	domain := PolymarketExchangeDomain(chainID, exchangeAddr)

	if domain.Name != "Polymarket CTF Exchange" {
		t.Errorf("Domain.Name = %s, expected Polymarket CTF Exchange", domain.Name)
	}
	if domain.Version != "1" {
		t.Errorf("Domain.Version = %s, expected 1", domain.Version)
	}
	if domain.ChainId.Cmp(big.NewInt(int64(chainID))) != 0 {
		t.Errorf("Domain.ChainId = %s, expected %d", domain.ChainId.String(), chainID)
	}
	if domain.VerifyingContract != exchangeAddr {
		t.Errorf("Domain.VerifyingContract = %s, expected %s", domain.VerifyingContract, exchangeAddr)
	}
}

func TestOrderTypes(t *testing.T) {
	fields, ok := OrderTypes["Order"]
	if !ok {
		t.Fatal("OrderTypes should have Order key")
	}

	expectedFields := []string{
		"salt", "maker", "signer", "taker", "tokenId",
		"makerAmount", "takerAmount", "expiration", "nonce",
		"feeRateBps", "side", "signatureType",
	}

	if len(fields) != len(expectedFields) {
		t.Errorf("Order has %d fields, expected %d", len(fields), len(expectedFields))
	}

	fieldNames := make(map[string]bool)
	for _, field := range fields {
		fieldNames[field.Name] = true
	}

	for _, name := range expectedFields {
		if !fieldNames[name] {
			t.Errorf("Missing field: %s", name)
		}
	}
}

func TestCredentialsStruct(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     "dGVzdC1zZWNyZXQ=", // base64 encoded
		Passphrase: "test-passphrase",
	}

	if creds.APIKey != "test-api-key" {
		t.Error("Credentials.APIKey mismatch")
	}
	if creds.Secret != "dGVzdC1zZWNyZXQ=" {
		t.Error("Credentials.Secret mismatch")
	}
	if creds.Passphrase != "test-passphrase" {
		t.Error("Credentials.Passphrase mismatch")
	}
}
