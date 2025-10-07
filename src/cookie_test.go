package src

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestSetChunkedCookiesNonChunked(t *testing.T) {
	config := &Config{
		CookieNamePrefix: "TraefikOidcAuth",
		SessionCookie: &SessionCookieConfig{
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
			SameSite: "default",
			MaxAge:   0,
		},
	}

	rw := newMockResponseWriter()

	setChunkedCookies(config, rw, "TraefikOidcAuth.Session", "some-short-value")

	setCookieHeader := rw.HeaderMap.Get("Set-Cookie")

	if setCookieHeader != "TraefikOidcAuth.Session=some-short-value; Path=/; HttpOnly; Secure" {
		t.Fail()
	}
}

func TestSetChunkedCookiesChunked(t *testing.T) {
	config := &Config{
		CookieNamePrefix: "TraefikOidcAuth",
		SessionCookie: &SessionCookieConfig{
			Path:     "/",
			Domain:   "",
			Secure:   true,
			HttpOnly: true,
			SameSite: "default",
			MaxAge:   0,
		},
	}

	rw := newMockResponseWriter()

	longValue := randomFixedLengthString(4000)

	setChunkedCookies(config, rw, "TraefikOidcAuth.Session", longValue)

	setCookieHeader := rw.HeaderMap.Values("Set-Cookie")

	if len(setCookieHeader) != 3 {
		t.Fail()
	}

	if setCookieHeader[0] != "TraefikOidcAuth.Session.Chunks=2; Path=/; HttpOnly; Secure" {
		t.Fail()
	}
	if setCookieHeader[1] != fmt.Sprintf("TraefikOidcAuth.Session.1=%s; Path=/; HttpOnly; Secure", longValue[:3072]) {
		t.Fail()
	}
	if setCookieHeader[2] != fmt.Sprintf("TraefikOidcAuth.Session.2=%s; Path=/; HttpOnly; Secure", longValue[3072:]) {
		t.Fail()
	}
}

func TestReadChunkedCookieOrdered(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fail()
	}

	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.Chunks",
		Value: "3",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.1",
		Value: "111",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.2",
		Value: "222",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.3",
		Value: "333",
	})

	cookieValue, err := readChunkedCookie(req, "TraefikOidcAuth.Session")
	if err != nil {
		t.Fail()
	}

	if cookieValue != "111222333" {
		t.Fail()
	}
}

func TestReadChunkedCookieUnordered(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fail()
	}

	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.3",
		Value: "333",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.Chunks",
		Value: "3",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.1",
		Value: "111",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.2",
		Value: "222",
	})

	cookieValue, err := readChunkedCookie(req, "TraefikOidcAuth.Session")
	if err != nil {
		t.Fail()
	}

	if cookieValue != "111222333" {
		t.Fail()
	}
}

func TestReadChunkedCookieWithIncompleteChunks(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fail()
	}

	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.Chunks",
		Value: "3",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.1",
		Value: "111",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.2",
		Value: "222",
	})

	cookieValue, err := readChunkedCookie(req, "TraefikOidcAuth.Session")

	// readChunkedCookie should fail
	if err == nil || cookieValue != "" {
		t.Fail()
	}
}

func TestReadChunkedCookieWithNoCount(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fail()
	}

	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.3",
		Value: "333",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.1",
		Value: "111",
	})
	req.AddCookie(&http.Cookie{
		Name:  "TraefikOidcAuth.Session.2",
		Value: "222",
	})

	cookieValue, err := readChunkedCookie(req, "TraefikOidcAuth.Session")

	// readChunkedCookie should fail
	if err == nil || cookieValue != "" {
		t.Fail()
	}
}

type mockResponseWriter struct {
	HeaderMap http.Header
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		HeaderMap: make(http.Header),
	}
}

func (writer *mockResponseWriter) Header() http.Header {
	return writer.HeaderMap
}
func (writer *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (writer *mockResponseWriter) WriteHeader(statusCode int) {
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomFixedLengthString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestParseCookieSameSite(t *testing.T) {
	// Test all supported SameSite values
	testCases := []struct {
		input    string
		expected http.SameSite
	}{
		{"none", http.SameSiteNoneMode},
		{"lax", http.SameSiteLaxMode},
		{"strict", http.SameSiteStrictMode},
		{"default", http.SameSiteDefaultMode},
		{"unknown", http.SameSiteDefaultMode},
		{"", http.SameSiteDefaultMode},
	}
	
	for _, tc := range testCases {
		result := parseCookieSameSite(tc.input)
		if result != tc.expected {
			t.Errorf("parseCookieSameSite(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestMakeCookieExpireImmediately(t *testing.T) {
	cookie := &http.Cookie{
		Name:  "test-cookie",
		Value: "test-value",
	}
	
	// Make cookie expire
	result := makeCookieExpireImmediately(cookie)
	
	// Should be the same cookie object
	if result != cookie {
		t.Errorf("Expected same cookie object")
	}
	
	// Should have negative MaxAge
	if result.MaxAge != -1 {
		t.Errorf("Expected MaxAge -1, got %d", result.MaxAge)
	}
	
	// Should have past expiry time
	if result.Expires.After(time.Now()) {
		t.Errorf("Expected past expiry time, got %v", result.Expires)
	}
}

func TestGetCodeVerifierCookieName(t *testing.T) {
	config := &Config{
		CookieNamePrefix: "TestApp",
	}
	
	result := getCodeVerifierCookieName(config)
	expected := "TestApp.CodeVerifier"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetSessionCookieName(t *testing.T) {
	config := &Config{
		CookieNamePrefix: "TestApp",
	}
	
	result := getSessionCookieName(config)
	expected := "TestApp.Session"
	
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestMakeCookieName(t *testing.T) {
	config := &Config{
		CookieNamePrefix: "MyApp",
	}
	
	testCases := []struct {
		name     string
		expected string
	}{
		{"Session", "MyApp.Session"},
		{"CodeVerifier", "MyApp.CodeVerifier"},
		{"CustomName", "MyApp.CustomName"},
	}
	
	for _, tc := range testCases {
		result := makeCookieName(config, tc.name)
		if result != tc.expected {
			t.Errorf("makeCookieName(%q) = %q, expected %q", tc.name, result, tc.expected)
		}
	}
}


