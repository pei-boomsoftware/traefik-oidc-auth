package oidc

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestEncodeState(t *testing.T) {
	state := &OidcState{
		Action:      "login",
		RedirectUrl: "https://example.com/callback",
	}
	
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatalf("EncodeState failed: %v", err)
	}
	
	if encoded == "" {
		t.Errorf("Expected non-empty encoded state")
	}
	
	// Verify it's valid base64
	_, err = base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Errorf("Encoded state should be valid base64: %v", err)
	}
}

func TestDecodeState(t *testing.T) {
	originalState := &OidcState{
		Action:      "logout",
		RedirectUrl: "https://example.com/home",
	}
	
	// First encode it
	encoded, err := EncodeState(originalState)
	if err != nil {
		t.Fatalf("EncodeState failed: %v", err)
	}
	
	// Then decode it back
	decodedState, err := DecodeState(encoded)
	if err != nil {
		t.Fatalf("DecodeState failed: %v", err)
	}
	
	if decodedState == nil {
		t.Fatalf("Expected non-nil decoded state")
	}
	
	if decodedState.Action != originalState.Action {
		t.Errorf("Expected Action '%s', got '%s'", originalState.Action, decodedState.Action)
	}
	
	if decodedState.RedirectUrl != originalState.RedirectUrl {
		t.Errorf("Expected RedirectUrl '%s', got '%s'", originalState.RedirectUrl, decodedState.RedirectUrl)
	}
}

func TestEncodeDecodeState_RoundTrip(t *testing.T) {
	testCases := []OidcState{
		{
			Action:      "login",
			RedirectUrl: "https://example.com/dashboard",
		},
		{
			Action:      "logout",
			RedirectUrl: "https://example.com/goodbye",
		},
		{
			Action:      "",
			RedirectUrl: "https://example.com/empty-action",
		},
		{
			Action:      "special-action",
			RedirectUrl: "",
		},
		{
			Action:      "test with spaces",
			RedirectUrl: "https://example.com/path with spaces",
		},
		{
			Action:      "unicode-test-🚀",
			RedirectUrl: "https://example.com/unicode-🌟",
		},
	}
	
	for i, originalState := range testCases {
		// Encode
		encoded, err := EncodeState(&originalState)
		if err != nil {
			t.Errorf("Test case %d: EncodeState failed: %v", i, err)
			continue
		}
		
		// Decode
		decodedState, err := DecodeState(encoded)
		if err != nil {
			t.Errorf("Test case %d: DecodeState failed: %v", i, err)
			continue
		}
		
		// Compare
		if decodedState.Action != originalState.Action {
			t.Errorf("Test case %d: Expected Action '%s', got '%s'", i, originalState.Action, decodedState.Action)
		}
		
		if decodedState.RedirectUrl != originalState.RedirectUrl {
			t.Errorf("Test case %d: Expected RedirectUrl '%s', got '%s'", i, originalState.RedirectUrl, decodedState.RedirectUrl)
		}
	}
}

func TestDecodeState_InvalidBase64(t *testing.T) {
	invalidBase64 := "invalid-base64-string-with-invalid-chars-!!!"
	
	decodedState, err := DecodeState(invalidBase64)
	if err == nil {
		t.Errorf("Expected error for invalid base64")
	}
	
	if decodedState != nil {
		t.Errorf("Expected nil state for invalid base64")
	}
}

func TestDecodeState_InvalidJSON(t *testing.T) {
	// Create invalid JSON base64 encoded
	invalidJSON := "invalid-json-{broken"
	invalidJSONBase64 := base64.RawURLEncoding.EncodeToString([]byte(invalidJSON))
	
	decodedState, err := DecodeState(invalidJSONBase64)
	if err == nil {
		t.Errorf("Expected error for invalid JSON")
	}
	
	if decodedState != nil {
		t.Errorf("Expected nil state for invalid JSON")
	}
}

func TestDecodeState_EmptyString(t *testing.T) {
	decodedState, err := DecodeState("")
	if err == nil {
		t.Errorf("Expected error for empty string")
	}
	
	if decodedState != nil {
		t.Errorf("Expected nil state for empty string")
	}
}

func TestOidcState_JSONSerialization(t *testing.T) {
	state := &OidcState{
		Action:      "test-action",
		RedirectUrl: "https://test.example.com/callback",
	}
	
	// Test JSON marshaling
	jsonBytes, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}
	
	jsonString := string(jsonBytes)
	if !strings.Contains(jsonString, "test-action") {
		t.Errorf("Expected JSON to contain action")
	}
	
	if !strings.Contains(jsonString, "https://test.example.com/callback") {
		t.Errorf("Expected JSON to contain redirect URL")
	}
	
	// Test JSON unmarshaling
	var unmarshaled OidcState
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}
	
	if unmarshaled.Action != state.Action {
		t.Errorf("Expected Action '%s', got '%s'", state.Action, unmarshaled.Action)
	}
	
	if unmarshaled.RedirectUrl != state.RedirectUrl {
		t.Errorf("Expected RedirectUrl '%s', got '%s'", state.RedirectUrl, unmarshaled.RedirectUrl)
	}
}

func TestEncodeState_ValidBase64URL(t *testing.T) {
	state := &OidcState{
		Action:      "test",
		RedirectUrl: "https://example.com/test",
	}
	
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatalf("EncodeState failed: %v", err)
	}
	
	// Verify it uses URL-safe base64 (no padding)
	if strings.Contains(encoded, "=") {
		t.Errorf("Expected no padding in URL-safe base64")
	}
	
	if strings.Contains(encoded, "+") || strings.Contains(encoded, "/") {
		t.Errorf("Expected URL-safe base64 characters only")
	}
}