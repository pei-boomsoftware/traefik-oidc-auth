package session

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateSessionId(t *testing.T) {
	sessionId1 := GenerateSessionId()
	sessionId2 := GenerateSessionId()
	
	// Test that session IDs are generated
	if sessionId1 == "" {
		t.Errorf("Expected non-empty session ID")
	}
	
	if sessionId2 == "" {
		t.Errorf("Expected non-empty session ID")
	}
	
	// Test that session IDs are unique
	if sessionId1 == sessionId2 {
		t.Errorf("Expected unique session IDs, got duplicate: %s", sessionId1)
	}
	
	// Test that session IDs follow UUID format (basic check)
	if len(sessionId1) != 36 {
		t.Errorf("Expected session ID length of 36, got %d", len(sessionId1))
	}
	
	if strings.Count(sessionId1, "-") != 4 {
		t.Errorf("Expected 4 hyphens in UUID format, got %d", strings.Count(sessionId1, "-"))
	}
}

func TestSessionState(t *testing.T) {
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
	
	// Test that all fields are set correctly
	if sessionState.Id != "test-session-123" {
		t.Errorf("Expected Id 'test-session-123', got '%s'", sessionState.Id)
	}
	
	if !sessionState.RefreshedAt.Equal(now) {
		t.Errorf("Expected RefreshedAt to be %v, got %v", now, sessionState.RefreshedAt)
	}
	
	if sessionState.AccessToken != "access-token-xyz" {
		t.Errorf("Expected AccessToken 'access-token-xyz', got '%s'", sessionState.AccessToken)
	}
	
	if sessionState.IdToken != "id-token-abc" {
		t.Errorf("Expected IdToken 'id-token-abc', got '%s'", sessionState.IdToken)
	}
	
	if sessionState.RefreshToken != "refresh-token-def" {
		t.Errorf("Expected RefreshToken 'refresh-token-def', got '%s'", sessionState.RefreshToken)
	}
	
	if !sessionState.IsAuthorized {
		t.Errorf("Expected IsAuthorized to be true")
	}
	
	if sessionState.TokenExpiresIn != 3600 {
		t.Errorf("Expected TokenExpiresIn 3600, got %d", sessionState.TokenExpiresIn)
	}
}

func TestSessionStateDefault(t *testing.T) {
	sessionState := &SessionState{}
	
	// Test default values
	if sessionState.Id != "" {
		t.Errorf("Expected empty Id by default, got '%s'", sessionState.Id)
	}
	
	if !sessionState.RefreshedAt.IsZero() {
		t.Errorf("Expected zero RefreshedAt by default, got %v", sessionState.RefreshedAt)
	}
	
	if sessionState.AccessToken != "" {
		t.Errorf("Expected empty AccessToken by default, got '%s'", sessionState.AccessToken)
	}
	
	if sessionState.IdToken != "" {
		t.Errorf("Expected empty IdToken by default, got '%s'", sessionState.IdToken)
	}
	
	if sessionState.RefreshToken != "" {
		t.Errorf("Expected empty RefreshToken by default, got '%s'", sessionState.RefreshToken)
	}
	
	if sessionState.IsAuthorized {
		t.Errorf("Expected IsAuthorized to be false by default")
	}
	
	if sessionState.TokenExpiresIn != 0 {
		t.Errorf("Expected TokenExpiresIn 0 by default, got %d", sessionState.TokenExpiresIn)
	}
}

// Mock implementation of SessionStorage for testing
type MockSessionStorage struct {
	sessions map[string]*SessionState
}

func NewMockSessionStorage() *MockSessionStorage {
	return &MockSessionStorage{
		sessions: make(map[string]*SessionState),
	}
}

func (m *MockSessionStorage) StoreSession(sessionId string, state *SessionState) (string, error) {
	ticket := "ticket-" + sessionId
	m.sessions[ticket] = state
	return ticket, nil
}

func (m *MockSessionStorage) TryGetSession(sessionTicket string) (*SessionState, error) {
	state, exists := m.sessions[sessionTicket]
	if !exists {
		return nil, nil
	}
	return state, nil
}

func TestMockSessionStorage(t *testing.T) {
	storage := NewMockSessionStorage()
	
	// Test storing and retrieving a session
	sessionState := &SessionState{
		Id:           "test-session",
		AccessToken:  "test-token",
		IsAuthorized: true,
	}
	
	ticket, err := storage.StoreSession("test-session", sessionState)
	if err != nil {
		t.Fatalf("StoreSession failed: %v", err)
	}
	
	if !strings.HasPrefix(ticket, "ticket-") {
		t.Errorf("Expected ticket to start with 'ticket-', got '%s'", ticket)
	}
	
	// Test retrieving the session
	retrievedState, err := storage.TryGetSession(ticket)
	if err != nil {
		t.Fatalf("TryGetSession failed: %v", err)
	}
	
	if retrievedState == nil {
		t.Fatalf("Expected to retrieve session state")
	}
	
	if retrievedState.Id != "test-session" {
		t.Errorf("Expected Id 'test-session', got '%s'", retrievedState.Id)
	}
	
	if retrievedState.AccessToken != "test-token" {
		t.Errorf("Expected AccessToken 'test-token', got '%s'", retrievedState.AccessToken)
	}
	
	if !retrievedState.IsAuthorized {
		t.Errorf("Expected IsAuthorized to be true")
	}
	
	// Test retrieving non-existent session
	nonExistentState, err := storage.TryGetSession("non-existent-ticket")
	if err != nil {
		t.Fatalf("TryGetSession should not error for non-existent session: %v", err)
	}
	
	if nonExistentState != nil {
		t.Errorf("Expected nil for non-existent session")
	}
}

func TestSessionInterface(t *testing.T) {
	// Test that MockSessionStorage implements SessionStorage interface
	var storage SessionStorage = NewMockSessionStorage()
	
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