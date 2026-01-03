package common

import (
	"math/big"
	"testing"
	"time"
)

func TestTimestampMs(t *testing.T) {
	before := time.Now().UnixMilli()
	ts := TimestampMs()
	after := time.Now().UnixMilli()

	if ts < before || ts > after {
		t.Errorf("TimestampMs() = %d, expected between %d and %d", ts, before, after)
	}
}

func TestTimestampSec(t *testing.T) {
	before := time.Now().Unix()
	ts := TimestampSec()
	after := time.Now().Unix()

	if ts < before || ts > after {
		t.Errorf("TimestampSec() = %d, expected between %d and %d", ts, before, after)
	}
}

func TestTimestampMsStr(t *testing.T) {
	ts := TimestampMsStr()
	if len(ts) < 13 {
		t.Errorf("TimestampMsStr() = %s, expected at least 13 digits", ts)
	}
}

func TestTimestampSecStr(t *testing.T) {
	ts := TimestampSecStr()
	if len(ts) < 10 {
		t.Errorf("TimestampSecStr() = %s, expected at least 10 digits", ts)
	}
}

func TestGenerateRandomHex(t *testing.T) {
	tests := []struct {
		length         int
		expectedLength int
	}{
		{16, 32}, // 16 bytes = 32 hex chars
		{32, 64},
		{1, 2},
	}

	for _, tt := range tests {
		hex, err := GenerateRandomHex(tt.length)
		if err != nil {
			t.Errorf("GenerateRandomHex(%d) error: %v", tt.length, err)
		}
		if len(hex) != tt.expectedLength {
			t.Errorf("GenerateRandomHex(%d) length = %d, expected %d", tt.length, len(hex), tt.expectedLength)
		}
	}

	// Test randomness
	hex1, _ := GenerateRandomHex(16)
	hex2, _ := GenerateRandomHex(16)
	if hex1 == hex2 {
		t.Error("GenerateRandomHex should generate different values")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Errorf("GenerateSalt() error: %v", err)
	}
	if salt == nil {
		t.Error("GenerateSalt() returned nil")
	}
	if salt.Sign() < 0 {
		t.Error("GenerateSalt() should return non-negative value")
	}

	// Test randomness
	salt2, _ := GenerateSalt()
	if salt.Cmp(salt2) == 0 {
		t.Error("GenerateSalt should generate different values")
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Errorf("GenerateNonce() error: %v", err)
	}
	if nonce == nil {
		t.Error("GenerateNonce() returned nil")
	}
}

func TestSaltToString(t *testing.T) {
	tests := []struct {
		salt     *big.Int
		expected string
	}{
		{nil, "0"},
		{big.NewInt(0), "0"},
		{big.NewInt(12345), "12345"},
		{big.NewInt(9999999999), "9999999999"},
	}

	for _, tt := range tests {
		result := SaltToString(tt.salt)
		if result != tt.expected {
			t.Errorf("SaltToString(%v) = %s, expected %s", tt.salt, result, tt.expected)
		}
	}
}

func TestStringToBigInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		ok       bool
	}{
		{"0", 0, true},
		{"12345", 12345, true},
		{"9999999999", 9999999999, true},
		{"invalid", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		result, ok := StringToBigInt(tt.input)
		if ok != tt.ok {
			t.Errorf("StringToBigInt(%s) ok = %v, expected %v", tt.input, ok, tt.ok)
		}
		if ok && result.Int64() != tt.expected {
			t.Errorf("StringToBigInt(%s) = %d, expected %d", tt.input, result.Int64(), tt.expected)
		}
	}
}

func TestHexToBigInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		ok       bool
	}{
		{"0x0", 0, true},
		{"0x10", 16, true},
		{"0xff", 255, true},
		{"ff", 255, true},
		{"0x1234", 4660, true},
		{"invalid", 0, false},
	}

	for _, tt := range tests {
		result, ok := HexToBigInt(tt.input)
		if ok != tt.ok {
			t.Errorf("HexToBigInt(%s) ok = %v, expected %v", tt.input, ok, tt.ok)
		}
		if ok && result.Int64() != tt.expected {
			t.Errorf("HexToBigInt(%s) = %d, expected %d", tt.input, result.Int64(), tt.expected)
		}
	}
}

func TestBigIntToHex(t *testing.T) {
	tests := []struct {
		input    *big.Int
		expected string
	}{
		{nil, "0x0"},
		{big.NewInt(0), "0x0"},
		{big.NewInt(16), "0x10"},
		{big.NewInt(255), "0xff"},
	}

	for _, tt := range tests {
		result := BigIntToHex(tt.input)
		if result != tt.expected {
			t.Errorf("BigIntToHex(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"0x1234abcd", "0x1234abcd"},
		{"1234ABCD", "0x1234ABCD"},
		{"0x1234ABCD", "0x1234ABCD"},
	}

	for _, tt := range tests {
		result := NormalizeAddress(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeAddress(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"0x", false},
		{"0x1234567890123456789012345678901234567890", true},
		{"1234567890123456789012345678901234567890", true},
		{"0x123456789012345678901234567890123456789", false},  // too short
		{"0x12345678901234567890123456789012345678901", false}, // too long
		{"0x123456789012345678901234567890123456789g", false}, // invalid char
	}

	for _, tt := range tests {
		result := IsValidAddress(tt.input)
		if result != tt.expected {
			t.Errorf("IsValidAddress(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestMinMaxInt(t *testing.T) {
	if MinInt(1, 2) != 1 {
		t.Error("MinInt(1, 2) should return 1")
	}
	if MinInt(2, 1) != 1 {
		t.Error("MinInt(2, 1) should return 1")
	}
	if MinInt(1, 1) != 1 {
		t.Error("MinInt(1, 1) should return 1")
	}

	if MaxInt(1, 2) != 2 {
		t.Error("MaxInt(1, 2) should return 2")
	}
	if MaxInt(2, 1) != 2 {
		t.Error("MaxInt(2, 1) should return 2")
	}
	if MaxInt(2, 2) != 2 {
		t.Error("MaxInt(2, 2) should return 2")
	}
}

func TestIsZeroAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{ZeroAddress, true},
		{"0x0000000000000000000000000000000000000000", true},
		{"0000000000000000000000000000000000000000", true},
		{"0x0000000000000000000000000000000000000001", false},
		{"0x1234567890123456789012345678901234567890", false},
	}

	for _, tt := range tests {
		result := IsZeroAddress(tt.input)
		if result != tt.expected {
			t.Errorf("IsZeroAddress(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
