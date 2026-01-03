package auth

import (
	"encoding/base64"
	"testing"
)

// 用于测试的私钥
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

func TestNewCredentialsManager(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)

	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	if manager == nil {
		t.Fatal("NewCredentialsManager() returned nil")
	}
	if manager.GetL1Signer() != signer {
		t.Error("GetL1Signer() should return the same signer")
	}
}

func TestCredentialsManagerGetCredentials(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Initially nil
	if manager.GetCredentials() != nil {
		t.Error("GetCredentials() should return nil initially")
	}
}

func TestCredentialsManagerSetCredentials(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	creds := &Credentials{
		APIKey:     "test-key",
		Secret:     "dGVzdC1zZWNyZXQ=",
		Passphrase: "test-pass",
	}

	manager.SetCredentials(creds)

	if manager.GetCredentials() != creds {
		t.Error("GetCredentials() should return the set credentials")
	}
}

func TestCredentialsManagerHasCredentials(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Initially false
	if manager.HasCredentials() {
		t.Error("HasCredentials() should return false initially")
	}

	// Empty credentials
	manager.SetCredentials(&Credentials{})
	if manager.HasCredentials() {
		t.Error("HasCredentials() should return false for empty credentials")
	}

	// Partial credentials
	manager.SetCredentials(&Credentials{APIKey: "key"})
	if manager.HasCredentials() {
		t.Error("HasCredentials() should return false for partial credentials")
	}

	// Full credentials
	manager.SetCredentials(&Credentials{
		APIKey:     "key",
		Secret:     "secret",
		Passphrase: "pass",
	})
	if !manager.HasCredentials() {
		t.Error("HasCredentials() should return true for complete credentials")
	}
}

func TestCredentialsManagerGetL2Signer(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Without credentials
	_, err := manager.GetL2Signer()
	if err == nil {
		t.Error("GetL2Signer() should fail without credentials")
	}

	// With credentials
	manager.SetCredentials(&Credentials{
		APIKey:     "key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("secret")),
		Passphrase: "pass",
	})

	l2Signer, err := manager.GetL2Signer()
	if err != nil {
		t.Fatalf("GetL2Signer() error: %v", err)
	}

	if l2Signer == nil {
		t.Fatal("GetL2Signer() returned nil")
	}

	if l2Signer.GetAddress() != signer.GetAddress() {
		t.Error("L2Signer address should match L1Signer address")
	}
}

func TestCredentialsManagerGetL1Signer(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	if manager.GetL1Signer() != signer {
		t.Error("GetL1Signer() should return the original signer")
	}
}

func TestCredentialsManagerGetAddress(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	if manager.GetAddress() != signer.GetAddress() {
		t.Error("GetAddress() should match signer address")
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name    string
		creds   *Credentials
		wantErr bool
	}{
		{
			name:    "nil credentials",
			creds:   nil,
			wantErr: true,
		},
		{
			name:    "empty APIKey",
			creds:   &Credentials{APIKey: "", Secret: "s", Passphrase: "p"},
			wantErr: true,
		},
		{
			name:    "empty Secret",
			creds:   &Credentials{APIKey: "k", Secret: "", Passphrase: "p"},
			wantErr: true,
		},
		{
			name:    "empty Passphrase",
			creds:   &Credentials{APIKey: "k", Secret: "s", Passphrase: ""},
			wantErr: true,
		},
		{
			name:    "valid credentials",
			creds:   &Credentials{APIKey: "k", Secret: "s", Passphrase: "p"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentials(tt.creds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCredentialsErrorMessages(t *testing.T) {
	err := ValidateCredentials(nil)
	if err == nil || err.Error() != "credentials is nil" {
		t.Errorf("Expected 'credentials is nil', got %v", err)
	}

	err = ValidateCredentials(&Credentials{})
	if err == nil || err.Error() != "API key is empty" {
		t.Errorf("Expected 'API key is empty', got %v", err)
	}

	err = ValidateCredentials(&Credentials{APIKey: "k"})
	if err == nil || err.Error() != "secret is empty" {
		t.Errorf("Expected 'secret is empty', got %v", err)
	}

	err = ValidateCredentials(&Credentials{APIKey: "k", Secret: "s"})
	if err == nil || err.Error() != "passphrase is empty" {
		t.Errorf("Expected 'passphrase is empty', got %v", err)
	}
}

func TestCredentialsManagerL2SignerUsage(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Set credentials
	secret := base64.StdEncoding.EncodeToString([]byte("test-secret"))
	manager.SetCredentials(&Credentials{
		APIKey:     "test-api-key",
		Secret:     secret,
		Passphrase: "test-passphrase",
	})

	// Get L2 signer
	l2Signer, err := manager.GetL2Signer()
	if err != nil {
		t.Fatalf("GetL2Signer() error: %v", err)
	}

	// Verify L2 signer works
	headers, err := l2Signer.GetAuthHeaders("GET", "/orders", "")
	if err != nil {
		t.Fatalf("GetAuthHeaders() error: %v", err)
	}

	if headers.Address != signer.GetAddress() {
		t.Error("L2 signer address mismatch")
	}
	if headers.APIKey != "test-api-key" {
		t.Error("L2 signer API key mismatch")
	}
	if headers.Passphrase != "test-passphrase" {
		t.Error("L2 signer passphrase mismatch")
	}
}

func TestCredentialsManagerMultipleL2Signers(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Set initial credentials
	manager.SetCredentials(&Credentials{
		APIKey:     "key1",
		Secret:     base64.StdEncoding.EncodeToString([]byte("secret1")),
		Passphrase: "pass1",
	})

	l2Signer1, _ := manager.GetL2Signer()

	// Update credentials
	manager.SetCredentials(&Credentials{
		APIKey:     "key2",
		Secret:     base64.StdEncoding.EncodeToString([]byte("secret2")),
		Passphrase: "pass2",
	})

	l2Signer2, _ := manager.GetL2Signer()

	// Each call should return a new signer with current credentials
	if l2Signer1.GetCredentials().APIKey == l2Signer2.GetCredentials().APIKey {
		t.Error("New L2Signer should use updated credentials")
	}
}

func TestCredentialsManagerConcurrentAccess(t *testing.T) {
	signer, _ := NewL1Signer(testPrivateKey, 137)
	manager := NewCredentialsManager(signer, "https://clob.polymarket.com")

	// Set credentials
	manager.SetCredentials(&Credentials{
		APIKey:     "key",
		Secret:     base64.StdEncoding.EncodeToString([]byte("secret")),
		Passphrase: "pass",
	})

	// Concurrent access (basic test - no synchronization in current impl)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetCredentials()
			_ = manager.HasCredentials()
			_ = manager.GetAddress()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
