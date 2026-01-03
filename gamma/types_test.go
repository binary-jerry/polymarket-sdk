package gamma

import (
	"testing"
)

func TestMarketGetOutcomePrices(t *testing.T) {
	tests := []struct {
		name          string
		outcomePrices string
		expectedLen   int
		expectError   bool
	}{
		{
			name:          "empty",
			outcomePrices: "",
			expectedLen:   0,
			expectError:   false,
		},
		{
			name:          "json array",
			outcomePrices: `["0.5","0.5"]`,
			expectedLen:   2,
			expectError:   false,
		},
		{
			name:          "comma separated",
			outcomePrices: "0.3,0.7",
			expectedLen:   2,
			expectError:   false,
		},
		{
			name:          "invalid number",
			outcomePrices: "invalid",
			expectedLen:   0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Market{OutcomePrices: tt.outcomePrices}
			prices, err := m.GetOutcomePrices()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(prices) != tt.expectedLen {
				t.Errorf("Got %d prices, expected %d", len(prices), tt.expectedLen)
			}
		})
	}
}

func TestMarketGetClobTokenIDs(t *testing.T) {
	tests := []struct {
		name        string
		clobTokenIds string
		expectedLen int
	}{
		{
			name:         "empty",
			clobTokenIds: "",
			expectedLen:  0,
		},
		{
			name:         "json array",
			clobTokenIds: `["token1","token2"]`,
			expectedLen:  2,
		},
		{
			name:         "comma separated",
			clobTokenIds: "token1,token2",
			expectedLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Market{ClobTokenIds: tt.clobTokenIds}
			ids := m.GetClobTokenIDs()
			if len(ids) != tt.expectedLen {
				t.Errorf("Got %d token IDs, expected %d", len(ids), tt.expectedLen)
			}
		})
	}
}

func TestMarketGetEndDate(t *testing.T) {
	tests := []struct {
		name       string
		endDate    string
		endDateIso string
		expectZero bool
	}{
		{
			name:       "both empty",
			endDate:    "",
			endDateIso: "",
			expectZero: true,
		},
		{
			name:       "endDateIso set",
			endDate:    "",
			endDateIso: "2024-12-31T23:59:59Z",
			expectZero: false,
		},
		{
			name:       "endDate set",
			endDate:    "2024-12-31T23:59:59Z",
			endDateIso: "",
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Market{EndDate: tt.endDate, EndDateIso: tt.endDateIso}
			date, err := m.GetEndDate()
			if err != nil && !tt.expectZero {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectZero && !date.IsZero() {
				t.Error("Expected zero time")
			}
			if !tt.expectZero && date.IsZero() {
				t.Error("Expected non-zero time")
			}
		})
	}
}

func TestMarketIsActive(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		closed   bool
		archived bool
		expected bool
	}{
		{
			name:     "all active",
			active:   true,
			closed:   false,
			archived: false,
			expected: true,
		},
		{
			name:     "closed",
			active:   true,
			closed:   true,
			archived: false,
			expected: false,
		},
		{
			name:     "archived",
			active:   true,
			closed:   false,
			archived: true,
			expected: false,
		},
		{
			name:     "not active",
			active:   false,
			closed:   false,
			archived: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Market{Active: tt.active, Closed: tt.closed, Archived: tt.archived}
			if m.IsActive() != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", m.IsActive(), tt.expected)
			}
		})
	}
}

func TestMarketIsNegRisk(t *testing.T) {
	m := &Market{NegRisk: true}
	if !m.IsNegRisk() {
		t.Error("IsNegRisk() should return true")
	}

	m.NegRisk = false
	if m.IsNegRisk() {
		t.Error("IsNegRisk() should return false")
	}
}

func TestMarketGetYesNoToken(t *testing.T) {
	m := &Market{
		Tokens: []Token{
			{TokenID: "yes-token", Outcome: "Yes"},
			{TokenID: "no-token", Outcome: "No"},
		},
	}

	yesToken := m.GetYesToken()
	if yesToken == nil {
		t.Error("GetYesToken() should not return nil")
	}
	if yesToken.TokenID != "yes-token" {
		t.Errorf("GetYesToken().TokenID = %s, expected yes-token", yesToken.TokenID)
	}

	noToken := m.GetNoToken()
	if noToken == nil {
		t.Error("GetNoToken() should not return nil")
	}
	if noToken.TokenID != "no-token" {
		t.Errorf("GetNoToken().TokenID = %s, expected no-token", noToken.TokenID)
	}

	// Test with empty tokens
	m.Tokens = nil
	if m.GetYesToken() != nil {
		t.Error("GetYesToken() should return nil for empty tokens")
	}
	if m.GetNoToken() != nil {
		t.Error("GetNoToken() should return nil for empty tokens")
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := BoolPtr(true)
	if truePtr == nil || *truePtr != true {
		t.Error("BoolPtr(true) should return pointer to true")
	}

	falsePtr := BoolPtr(false)
	if falsePtr == nil || *falsePtr != false {
		t.Error("BoolPtr(false) should return pointer to false")
	}
}

func TestSplitString(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"", ",", nil},
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{"a, b, c", ",", []string{"a", "b", "c"}},
		{"single", ",", []string{"single"}},
	}

	for _, tt := range tests {
		result := splitString(tt.input, tt.sep)
		if len(result) != len(tt.expected) {
			t.Errorf("splitString(%q, %q) length = %d, expected %d", tt.input, tt.sep, len(result), len(tt.expected))
		}
	}
}
