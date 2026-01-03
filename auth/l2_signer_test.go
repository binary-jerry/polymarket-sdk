package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewL2Signer(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("test-secret")),
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234567890123456789012345678901234567890", creds)

	if signer == nil {
		t.Fatal("NewL2Signer() returned nil")
	}
	if signer.GetAddress() != "0x1234567890123456789012345678901234567890" {
		t.Errorf("GetAddress() = %s, expected 0x1234567890123456789012345678901234567890", signer.GetAddress())
	}
}

func TestL2SignerSign(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	signature, err := signer.Sign("GET", "/orders", "1234567890", "")
	if err != nil {
		t.Fatalf("Sign() error: %v", err)
	}

	if signature == "" {
		t.Error("Signature should not be empty")
	}

	// Signature should be base64 encoded
	_, err = base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Errorf("Signature should be valid base64: %v", err)
	}
}

func TestL2SignerSignWithBody(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	body := `{"order_id":"12345"}`
	signature, err := signer.Sign("POST", "/orders", "1234567890", body)
	if err != nil {
		t.Fatalf("Sign() with body error: %v", err)
	}

	if signature == "" {
		t.Error("Signature should not be empty")
	}

	// Different body should produce different signature
	signature2, _ := signer.Sign("POST", "/orders", "1234567890", `{"order_id":"67890"}`)
	if signature == signature2 {
		t.Error("Different body should produce different signature")
	}
}

func TestL2SignerSignInvalidSecret(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     "not-valid-base64!!!",
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	_, err := signer.Sign("GET", "/orders", "1234567890", "")
	if err == nil {
		t.Error("Sign() should fail with invalid base64 secret")
	}
}

func TestL2SignerGetAuthHeaders(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	address := "0x1234567890123456789012345678901234567890"
	signer := NewL2Signer(address, creds)

	headers, err := signer.GetAuthHeaders("GET", "/orders", "")
	if err != nil {
		t.Fatalf("GetAuthHeaders() error: %v", err)
	}

	if headers.Address != address {
		t.Errorf("Address = %s, expected %s", headers.Address, address)
	}
	if headers.APIKey != creds.APIKey {
		t.Errorf("APIKey = %s, expected %s", headers.APIKey, creds.APIKey)
	}
	if headers.Passphrase != creds.Passphrase {
		t.Errorf("Passphrase = %s, expected %s", headers.Passphrase, creds.Passphrase)
	}
	if headers.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
	if headers.Signature == "" {
		t.Error("Signature should not be empty")
	}
}

func TestL2SignerSignRequest(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	address := "0x1234567890123456789012345678901234567890"
	signer := NewL2Signer(address, creds)

	// Create a test request
	req := httptest.NewRequest("GET", "https://api.example.com/orders?status=open", nil)

	err := signer.SignRequest(req, "")
	if err != nil {
		t.Fatalf("SignRequest() error: %v", err)
	}

	// Verify headers are set
	if req.Header.Get("POLY_ADDRESS") != address {
		t.Errorf("POLY_ADDRESS = %s, expected %s", req.Header.Get("POLY_ADDRESS"), address)
	}
	if req.Header.Get("POLY_API_KEY") != creds.APIKey {
		t.Errorf("POLY_API_KEY = %s, expected %s", req.Header.Get("POLY_API_KEY"), creds.APIKey)
	}
	if req.Header.Get("POLY_PASSPHRASE") != creds.Passphrase {
		t.Errorf("POLY_PASSPHRASE = %s, expected %s", req.Header.Get("POLY_PASSPHRASE"), creds.Passphrase)
	}
	if req.Header.Get("POLY_TIMESTAMP") == "" {
		t.Error("POLY_TIMESTAMP should not be empty")
	}
	if req.Header.Get("POLY_SIGNATURE") == "" {
		t.Error("POLY_SIGNATURE should not be empty")
	}
}

func TestL2SignerSignRequestWithQueryString(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	// Create request with query string
	req := httptest.NewRequest("GET", "https://api.example.com/orders?status=open&limit=10", nil)

	err := signer.SignRequest(req, "")
	if err != nil {
		t.Fatalf("SignRequest() with query string error: %v", err)
	}

	// The signature should include the query string
	if req.Header.Get("POLY_SIGNATURE") == "" {
		t.Error("POLY_SIGNATURE should not be empty")
	}
}

func TestL2SignerSignRequestWithBody(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	req := httptest.NewRequest("POST", "https://api.example.com/orders", nil)
	body := `{"price":"0.5","size":"100"}`

	err := signer.SignRequest(req, body)
	if err != nil {
		t.Fatalf("SignRequest() with body error: %v", err)
	}

	if req.Header.Get("POLY_SIGNATURE") == "" {
		t.Error("POLY_SIGNATURE should not be empty")
	}
}

func TestL2SignerGetCredentials(t *testing.T) {
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     "dGVzdC1zZWNyZXQ=",
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	returnedCreds := signer.GetCredentials()
	if returnedCreds != creds {
		t.Error("GetCredentials() should return the same credentials")
	}
}

func TestL2SignerGetAddress(t *testing.T) {
	signer := NewL2Signer("0xabcd1234", nil)
	if signer.GetAddress() != "0xabcd1234" {
		t.Errorf("GetAddress() = %s, expected 0xabcd1234", signer.GetAddress())
	}
}

func TestL2SignerUpdateCredentials(t *testing.T) {
	creds1 := &Credentials{APIKey: "key1"}
	creds2 := &Credentials{APIKey: "key2"}

	signer := NewL2Signer("0x1234", creds1)

	if signer.GetCredentials().APIKey != "key1" {
		t.Error("Initial credentials not set")
	}

	signer.UpdateCredentials(creds2)

	if signer.GetCredentials().APIKey != "key2" {
		t.Error("Credentials not updated")
	}
}

func TestL2SignerIsValid(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		creds    *Credentials
		expected bool
	}{
		{
			name:    "valid",
			address: "0x1234",
			creds: &Credentials{
				APIKey:     "key",
				Secret:     "secret",
				Passphrase: "pass",
			},
			expected: true,
		},
		{
			name:     "nil credentials",
			address:  "0x1234",
			creds:    nil,
			expected: false,
		},
		{
			name:    "empty APIKey",
			address: "0x1234",
			creds: &Credentials{
				APIKey:     "",
				Secret:     "secret",
				Passphrase: "pass",
			},
			expected: false,
		},
		{
			name:    "empty Secret",
			address: "0x1234",
			creds: &Credentials{
				APIKey:     "key",
				Secret:     "",
				Passphrase: "pass",
			},
			expected: false,
		},
		{
			name:    "empty Passphrase",
			address: "0x1234",
			creds: &Credentials{
				APIKey:     "key",
				Secret:     "secret",
				Passphrase: "",
			},
			expected: false,
		},
		{
			name:    "empty address",
			address: "",
			creds: &Credentials{
				APIKey:     "key",
				Secret:     "secret",
				Passphrase: "pass",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer := NewL2Signer(tt.address, tt.creds)
			if signer.IsValid() != tt.expected {
				t.Errorf("IsValid() = %v, expected %v", signer.IsValid(), tt.expected)
			}
		})
	}
}

func TestL2SignerDeterministicSignature(t *testing.T) {
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	signer := NewL2Signer("0x1234", creds)

	// Same input should produce same signature
	sig1, _ := signer.Sign("GET", "/orders", "1234567890", "")
	sig2, _ := signer.Sign("GET", "/orders", "1234567890", "")

	if sig1 != sig2 {
		t.Error("Same input should produce deterministic signature")
	}
}

func TestL2SignerIntegration(t *testing.T) {
	// Test that signed request can be verified by a server
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	creds := &Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	}

	address := "0x1234567890123456789012345678901234567890"
	signer := NewL2Signer(address, creds)

	// Create a test server that verifies headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all required headers are present
		if r.Header.Get("POLY_ADDRESS") == "" {
			t.Error("Missing POLY_ADDRESS header")
		}
		if r.Header.Get("POLY_API_KEY") == "" {
			t.Error("Missing POLY_API_KEY header")
		}
		if r.Header.Get("POLY_PASSPHRASE") == "" {
			t.Error("Missing POLY_PASSPHRASE header")
		}
		if r.Header.Get("POLY_TIMESTAMP") == "" {
			t.Error("Missing POLY_TIMESTAMP header")
		}
		if r.Header.Get("POLY_SIGNATURE") == "" {
			t.Error("Missing POLY_SIGNATURE header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create and sign request
	req, _ := http.NewRequest("GET", server.URL+"/orders", nil)
	err := signer.SignRequest(req, "")
	if err != nil {
		t.Fatalf("SignRequest() error: %v", err)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
