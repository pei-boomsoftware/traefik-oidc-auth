package utils

import (
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestChunkString(t *testing.T) {
	originalText := "abcdefghijklmnopqrstuvwxyz"

	chunks := ChunkString(originalText, 10)

	if len(chunks) != 3 {
		t.Fail()
	}

	value := ""

	for i := 0; i < len(chunks); i++ {
		value += chunks[i]
	}

	if value != originalText {
		t.Fail()
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	secret := "MLFs4TT99kOOq8h3UAVRtYoCTDYXiRcZ"
	originalText := "hello"

	encrypted, err := Encrypt(originalText, secret)
	if err != nil {
		t.Fail()
	}

	decrypted, err := Decrypt(encrypted, secret)
	if err != nil {
		t.Fail()
	}

	if decrypted != originalText {
		t.Fail()
	}
}

func TestDecryptEmptyString(t *testing.T) {
	secret := "MLFs4TT99kOOq8h3UAVRtYoCTDYXiRcZ"

	_, err := Decrypt("", secret)

	// Must return an error
	if err == nil {
		t.Fail()
	}
}

func TestValidateRedirectUri(t *testing.T) {
	validUris := []string{
		"/",
		"https://example.com",
		"https://something.com",
	}

	expectRedirectUriMatch(t, "https://example.com", validUris, true)
	expectRedirectUriMatch(t, "https://malicious.com", validUris, false)
}

func TestValidateRedirectUriWildcards(t *testing.T) {
	validUris := []string{
		"/",
		"https://example.com",
		"https://something.com",
		"*",
	}

	expectRedirectUriMatch(t, "https://malicious.com", validUris, true)

	validUris = []string{
		"https://example.com",
		"https://*.something.com",
		"https://*.something.com/good",
		"https://*.something.com/good/*",
	}

	expectRedirectUriMatch(t, "https://app.something.com", validUris, true)
	expectRedirectUriMatch(t, "https://app.sub.something.com", validUris, false)
	expectRedirectUriMatch(t, "https://app.something.com/login", validUris, false)
	expectRedirectUriMatch(t, "https://app.something.com/good", validUris, true)
	expectRedirectUriMatch(t, "https://app.something.com/good/something", validUris, true)
	expectRedirectUriMatch(t, "https://app.something.com/good/something/bad", validUris, false)
}

func expectRedirectUriMatch(t *testing.T, uri string, validUris []string, shouldMatch bool) {
	matchedUri, err := ValidateRedirectUri(uri, validUris)

	if (shouldMatch && err != nil) || (!shouldMatch && err == nil) {
		t.Fail()
	}

	if (shouldMatch && matchedUri != uri) || (!shouldMatch && matchedUri != "") {
		t.Fail()
	}
}

func TestParseAcceptType(t *testing.T) {
	acceptType := ParseAcceptType("text/html")
	if acceptType.Type != "text/html" {
		t.Fail()
	}
	if acceptType.Weight != 1.0 {
		t.Fail()
	}

	acceptType = ParseAcceptType("text/html;q=0.8")
	if acceptType.Type != "text/html" {
		t.Fail()
	}
	if acceptType.Weight != 0.8 {
		t.Fail()
	}

	acceptType = ParseAcceptType("application/json; q=0.5")
	if acceptType.Type != "application/json" {
		t.Fail()
	}
	if acceptType.Weight != 0.5 {
		t.Fail()
	}

	acceptType = ParseAcceptType("text/html;q=invalid")
	if acceptType.Type != "" {
		t.Fail()
	}
	if acceptType.Weight != 0.0 {
		t.Fail()
	}

	acceptType = ParseAcceptType("*/*")
	if acceptType.Type != "*/*" {
		t.Fail()
	}
	if acceptType.Weight != 1.0 {
		t.Fail()
	}

	acceptType = ParseAcceptType("")
	if acceptType.Type != "" {
		t.Fail()
	}
	if acceptType.Weight != 0.0 {
		t.Fail()
	}
}

func TestParseAcceptHeader(t *testing.T) {
	acceptTypes := ParseAcceptHeader("text/html,application/json")
	if len(acceptTypes) != 2 {
		t.Fail()
	}
	if acceptTypes[0].Type != "text/html" {
		t.Fail()
	}
	if acceptTypes[0].Weight != 1.0 {
		t.Fail()
	}
	if acceptTypes[1].Type != "application/json" {
		t.Fail()
	}
	if acceptTypes[1].Weight != 1.0 {
		t.Fail()
	}

	acceptTypes = ParseAcceptHeader("application/json;q=0.8,text/html;q=0.9")
	if len(acceptTypes) != 2 {
		t.Fail()
	}
	if acceptTypes[0].Type != "text/html" {
		t.Fail()
	}
	if acceptTypes[0].Weight != 0.9 {
		t.Fail()
	}
	if acceptTypes[1].Type != "application/json" {
		t.Fail()
	}
	if acceptTypes[1].Weight != 0.8 {
		t.Fail()
	}

	acceptTypes = ParseAcceptHeader("*/*")
	if len(acceptTypes) != 1 {
		t.Fail()
	}
	if acceptTypes[0].Type != "*/*" {
		t.Fail()
	}
	if acceptTypes[0].Weight != 1.0 {
		t.Fail()
	}
}

func TestIsHtmlRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	if !IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")
	if IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html, application/json")
	if !IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json;q=0.9, text/html;q=0.8")
	if IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json;q=0.8, text/html;q=0.9")
	if !IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "*/*")
	if IsHtmlRequest(req) {
		t.Fail()
	}

	req, _ = http.NewRequest("GET", "/", nil)
	if IsHtmlRequest(req) {
		t.Fail()
	}
}

func TestExpandEnvironmentVariableString(t *testing.T) {
	// Test without environment variable syntax
	result := ExpandEnvironmentVariableString("plain-string")
	if result != "plain-string" {
		t.Errorf("Expected 'plain-string', got '%s'", result)
	}
	
	// Test with environment variable that exists
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")
	
	result = ExpandEnvironmentVariableString("${TEST_VAR}")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}
	
	// Test with environment variable that doesn't exist
	result = ExpandEnvironmentVariableString("${NON_EXISTENT_VAR}")
	if result != "${NON_EXISTENT_VAR}" {
		t.Errorf("Expected '${NON_EXISTENT_VAR}', got '%s'", result)
	}
	
	// Test with malformed syntax (no closing brace)
	result = ExpandEnvironmentVariableString("${MALFORMED")
	if result != "${MALFORMED" {
		t.Errorf("Expected '${MALFORMED', got '%s'", result)
	}
	
	// Test with malformed syntax (no opening brace)
	result = ExpandEnvironmentVariableString("MALFORMED}")
	if result != "MALFORMED}" {
		t.Errorf("Expected 'MALFORMED}', got '%s'", result)
	}
}

func TestExpandEnvironmentVariableBoolean(t *testing.T) {
	// Test with true values
	result, err := ExpandEnvironmentVariableBoolean("true", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Errorf("Expected true")
	}
	
	result, err = ExpandEnvironmentVariableBoolean("1", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Errorf("Expected true")
	}
	
	// Test with false values
	result, err = ExpandEnvironmentVariableBoolean("false", true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Errorf("Expected false")
	}
	
	result, err = ExpandEnvironmentVariableBoolean("0", true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Errorf("Expected false")
	}
	
	// Test with default value when empty string
	result, err = ExpandEnvironmentVariableBoolean("", true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Errorf("Expected default value true")
	}
	
	result, err = ExpandEnvironmentVariableBoolean("", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result {
		t.Errorf("Expected default value false")
	}
	
	// Test with invalid boolean value
	_, err = ExpandEnvironmentVariableBoolean("invalid", false)
	if err == nil {
		t.Errorf("Expected error for invalid boolean value")
	}
	
	// Test with environment variable
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	
	result, err = ExpandEnvironmentVariableBoolean("${TEST_BOOL}", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result {
		t.Errorf("Expected true from environment variable")
	}
}

func TestUrlIsAbsolute(t *testing.T) {
	// Test absolute URL
	u, _ := url.Parse("https://example.com/path")
	if !UrlIsAbsolute(u) {
		t.Errorf("Expected https://example.com/path to be absolute")
	}
	
	// Test another absolute URL
	u, _ = url.Parse("http://localhost:8080/api")
	if !UrlIsAbsolute(u) {
		t.Errorf("Expected http://localhost:8080/api to be absolute")
	}
	
	// Test relative URL (no scheme)
	u, _ = url.Parse("/path/to/resource")
	if UrlIsAbsolute(u) {
		t.Errorf("Expected /path/to/resource to be relative")
	}
	
	// Test relative URL (no host)
	u, _ = url.Parse("file:///path/to/file")
	if UrlIsAbsolute(u) {
		t.Errorf("Expected file:///path/to/file to be relative (no host)")
	}
	
	// Test empty URL
	u, _ = url.Parse("")
	if UrlIsAbsolute(u) {
		t.Errorf("Expected empty URL to be relative")
	}
}

func TestParseUrl(t *testing.T) {
	// Test empty URL
	_, err := ParseUrl("")
	if err == nil {
		t.Errorf("Expected error for empty URL")
	}
	
	// Test URL with scheme
	u, err := ParseUrl("https://example.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if u.Scheme != "https" {
		t.Errorf("Expected scheme 'https', got '%s'", u.Scheme)
	}
	if u.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", u.Host)
	}
	
	// Test URL without scheme (should default to https)
	u, err = ParseUrl("example.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if u.Scheme != "https" {
		t.Errorf("Expected default scheme 'https', got '%s'", u.Scheme)
	}
	if u.Host != "example.com" {
		t.Errorf("Expected host 'example.com', got '%s'", u.Host)
	}
	
	// Test HTTP URL
	u, err = ParseUrl("http://localhost:8080")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if u.Scheme != "http" {
		t.Errorf("Expected scheme 'http', got '%s'", u.Scheme)
	}
	
	// Test invalid scheme
	_, err = ParseUrl("ftp://example.com")
	if err == nil {
		t.Errorf("Expected error for invalid scheme")
	}
	
	// Test malformed URL
	_, err = ParseUrl("://invalid")
	if err == nil {
		t.Errorf("Expected error for malformed URL")
	}
}

func TestParseBigInt(t *testing.T) {
	// Test valid base64 encoded big int
	validInput := "AQAB" // Common RSA exponent
	result, err := ParseBigInt(validInput)
	if err != nil {
		t.Fatalf("ParseBigInt failed: %v", err)
	}
	if result == nil {
		t.Errorf("Expected non-nil big.Int")
	}
	
	// Test invalid base64
	_, err = ParseBigInt("invalid-base64-!!!")
	if err == nil {
		t.Errorf("Expected error for invalid base64")
	}
}

func TestParseInt(t *testing.T) {
	// Test valid base64 encoded int
	validInput := "AQAB" // Common RSA exponent
	result, err := ParseInt(validInput)
	if err != nil {
		t.Fatalf("ParseInt failed: %v", err)
	}
	if result == 0 {
		t.Errorf("Expected non-zero int")
	}
	
	// Test invalid base64
	_, err = ParseInt("invalid-base64-!!!")
	if err == nil {
		t.Errorf("Expected error for invalid base64")
	}
}

func TestGetFullHost(t *testing.T) {
	// Test with X-Forwarded-Host header
	req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	req.Header.Set("X-Forwarded-Host", "example.com")
	req.Header.Set("X-Forwarded-Proto", "https")
	
	fullHost := GetFullHost(req)
	if fullHost != "https://example.com" {
		t.Errorf("Expected 'https://example.com', got '%s'", fullHost)
	}
	
	// Test without X-Forwarded-Host (should use req.Host)
	req, _ = http.NewRequest("GET", "http://localhost:8080", nil)
	req.Host = "localhost:8080"
	
	fullHost = GetFullHost(req)
	if fullHost != "http://localhost:8080" {
		t.Errorf("Expected 'http://localhost:8080', got '%s'", fullHost)
	}
}
