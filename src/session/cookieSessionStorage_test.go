package session

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreateCookieSessionStorage(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	if storage == nil {
		t.Errorf("Expected non-nil CookieSessionStorage")
	}
}

func TestCookieSessionStorage_StoreSession(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	now := time.Now()
	sessionState := &SessionState{
		Id:             "test-session-123",
		RefreshedAt:    now,
		AccessToken:    "access-token-xyz",
		IdToken:        "id-token-abc",
		RefreshToken:   "refresh-token-def",
		IsAuthorized:   true,
		TokenExpiresIn: 3600,
	}
	
	ticket, err := storage.StoreSession("test-session-123", sessionState)
	if err != nil {
		t.Fatalf("StoreSession failed: %v", err)
	}
	
	if ticket == "" {
		t.Errorf("Expected non-empty ticket")
	}
	
	// Verify that the ticket is valid JSON
	var retrievedState SessionState
	err = json.Unmarshal([]byte(ticket), &retrievedState)
	if err != nil {
		t.Errorf("Ticket should be valid JSON: %v", err)
	}
	
	// Verify the content matches
	if retrievedState.Id != sessionState.Id {
		t.Errorf("Expected Id '%s', got '%s'", sessionState.Id, retrievedState.Id)
	}
	
	if retrievedState.AccessToken != sessionState.AccessToken {
		t.Errorf("Expected AccessToken '%s', got '%s'", sessionState.AccessToken, retrievedState.AccessToken)
	}
	
	if retrievedState.IsAuthorized != sessionState.IsAuthorized {
		t.Errorf("Expected IsAuthorized %v, got %v", sessionState.IsAuthorized, retrievedState.IsAuthorized)
	}
}

func TestCookieSessionStorage_TryGetSession_Success(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	now := time.Now()
	originalState := &SessionState{
		Id:             "test-session-456",
		RefreshedAt:    now,
		AccessToken:    "access-token-abc",
		IdToken:        "id-token-def",
		RefreshToken:   "refresh-token-ghi",
		IsAuthorized:   false,
		TokenExpiresIn: 1800,
	}
	
	// First store the session to get a valid ticket
	ticket, err := storage.StoreSession("test-session-456", originalState)
	if err != nil {
		t.Fatalf("StoreSession failed: %v", err)
	}
	
	// Now try to get the session back
	retrievedState, err := storage.TryGetSession(ticket)
	if err != nil {
		t.Fatalf("TryGetSession failed: %v", err)
	}
	
	if retrievedState == nil {
		t.Fatalf("Expected non-nil session state")
	}
	
	// Verify all fields match
	if retrievedState.Id != originalState.Id {
		t.Errorf("Expected Id '%s', got '%s'", originalState.Id, retrievedState.Id)
	}
	
	if retrievedState.AccessToken != originalState.AccessToken {
		t.Errorf("Expected AccessToken '%s', got '%s'", originalState.AccessToken, retrievedState.AccessToken)
	}
	
	if retrievedState.IdToken != originalState.IdToken {
		t.Errorf("Expected IdToken '%s', got '%s'", originalState.IdToken, retrievedState.IdToken)
	}
	
	if retrievedState.RefreshToken != originalState.RefreshToken {
		t.Errorf("Expected RefreshToken '%s', got '%s'", originalState.RefreshToken, retrievedState.RefreshToken)
	}
	
	if retrievedState.IsAuthorized != originalState.IsAuthorized {
		t.Errorf("Expected IsAuthorized %v, got %v", originalState.IsAuthorized, retrievedState.IsAuthorized)
	}
	
	if retrievedState.TokenExpiresIn != originalState.TokenExpiresIn {
		t.Errorf("Expected TokenExpiresIn %d, got %d", originalState.TokenExpiresIn, retrievedState.TokenExpiresIn)
	}
	
	// Note: Time comparison needs to be handled carefully due to JSON marshaling/unmarshaling
	if !retrievedState.RefreshedAt.Equal(originalState.RefreshedAt) {
		// Allow for small differences due to JSON serialization
		diff := retrievedState.RefreshedAt.Sub(originalState.RefreshedAt)
		if diff > time.Second || diff < -time.Second {
			t.Errorf("Expected RefreshedAt to be close to %v, got %v", originalState.RefreshedAt, retrievedState.RefreshedAt)
		}
	}
}

func TestCookieSessionStorage_TryGetSession_InvalidJSON(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	// Test with invalid JSON
	invalidTicket := "invalid-json-{broken"
	
	sessionState, err := storage.TryGetSession(invalidTicket)
	if err == nil {
		t.Errorf("Expected error for invalid JSON ticket")
	}
	
	if sessionState != nil {
		t.Errorf("Expected nil session state for invalid ticket")
	}
}

func TestCookieSessionStorage_TryGetSession_EmptyTicket(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	// Test with empty ticket
	sessionState, err := storage.TryGetSession("")
	if err == nil {
		t.Errorf("Expected error for empty ticket")
	}
	
	if sessionState != nil {
		t.Errorf("Expected nil session state for empty ticket")
	}
}

func TestCookieSessionStorage_RoundTrip(t *testing.T) {
	storage := CreateCookieSessionStorage()
	
	// Test multiple round trips
	testCases := []SessionState{
		{
			Id:             "session-1",
			AccessToken:    "token-1",
			IsAuthorized:   true,
			TokenExpiresIn: 3600,
		},
		{
			Id:             "session-2",
			AccessToken:    "token-2",
			IsAuthorized:   false,
			TokenExpiresIn: 1800,
			RefreshToken:   "refresh-2",
		},
		{
			Id:           "session-3",
			AccessToken:  "",
			IsAuthorized: false,
		},
	}
	
	for i, originalState := range testCases {
		// Store session
		ticket, err := storage.StoreSession(originalState.Id, &originalState)
		if err != nil {
			t.Fatalf("StoreSession failed for test case %d: %v", i, err)
		}
		
		// Retrieve session
		retrievedState, err := storage.TryGetSession(ticket)
		if err != nil {
			t.Fatalf("TryGetSession failed for test case %d: %v", i, err)
		}
		
		if retrievedState == nil {
			t.Fatalf("Expected non-nil session state for test case %d", i)
		}
		
		// Compare key fields
		if retrievedState.Id != originalState.Id {
			t.Errorf("Test case %d: Expected Id '%s', got '%s'", i, originalState.Id, retrievedState.Id)
		}
		
		if retrievedState.AccessToken != originalState.AccessToken {
			t.Errorf("Test case %d: Expected AccessToken '%s', got '%s'", i, originalState.AccessToken, retrievedState.AccessToken)
		}
		
		if retrievedState.IsAuthorized != originalState.IsAuthorized {
			t.Errorf("Test case %d: Expected IsAuthorized %v, got %v", i, originalState.IsAuthorized, retrievedState.IsAuthorized)
		}
	}
}

func TestCookieSessionStorage_ImplementsInterface(t *testing.T) {
	// Verify that CookieSessionStorage implements SessionStorage interface
	var storage SessionStorage = CreateCookieSessionStorage()
	
	sessionState := &SessionState{
		Id:          "interface-test",
		AccessToken: "interface-token",
	}
	
	ticket, err := storage.StoreSession("interface-test", sessionState)
	if err != nil {
		t.Fatalf("Interface StoreSession failed: %v", err)
	}
	
	retrievedState, err := storage.TryGetSession(ticket)
	if err != nil {
		t.Fatalf("Interface TryGetSession failed: %v", err)
	}
	
	if retrievedState == nil {
		t.Fatalf("Expected to retrieve session via interface")
	}
	
	if retrievedState.Id != "interface-test" {
		t.Errorf("Expected Id 'interface-test', got '%s'", retrievedState.Id)
	}
}