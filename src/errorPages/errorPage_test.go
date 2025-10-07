package errorPages

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sevensolutions/traefik-oidc-auth/src/logging"
)

func TestWriteError_Redirect(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	config := &ErrorPageConfig{
		RedirectTo: "https://example.com/error",
	}
	
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	
	data := map[string]interface{}{
		"statusCode": http.StatusUnauthorized,
		"statusName": "Unauthorized",
		"description": "Access denied",
	}
	
	WriteError(logger, config, recorder, req, data)
	
	if recorder.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, recorder.Code)
	}
	
	location := recorder.Header().Get("Location")
	if location != "https://example.com/error" {
		t.Errorf("Expected location %s, got %s", "https://example.com/error", location)
	}
}

func TestWriteError_HTML(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	config := &ErrorPageConfig{}
	
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "text/html")
	
	data := map[string]interface{}{
		"statusCode": http.StatusUnauthorized,
		"statusName": "Unauthorized", 
		"description": "Access denied",
	}
	
	WriteError(logger, config, recorder, req, data)
	
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	
	contentType := recorder.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}
	
	body := recorder.Body.String()
	if !strings.Contains(body, "Unauthorized") {
		t.Errorf("Expected body to contain 'Unauthorized'")
	}
	
	if !strings.Contains(body, "Access denied") {
		t.Errorf("Expected body to contain 'Access denied'")
	}
}

func TestWriteError_JSON(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	config := &ErrorPageConfig{}
	
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")
	
	data := map[string]interface{}{
		"statusCode": http.StatusUnauthorized,
		"statusName": "Unauthorized",
		"statusType": "https://example.com/errors/unauthorized",
		"description": "Access denied",
	}
	
	WriteError(logger, config, recorder, req, data)
	
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
	
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json+problem" {
		t.Errorf("Expected JSON problem content type, got %s", contentType)
	}
	
	body := recorder.Body.String()
	if !strings.Contains(body, "Unauthorized") {
		t.Errorf("Expected body to contain 'Unauthorized'")
	}
	
	if !strings.Contains(body, "Access denied") {
		t.Errorf("Expected body to contain 'Access denied'")
	}
}

func TestWriteError_CustomTemplate(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	// Create a temporary template file
	tempFile, err := os.CreateTemp("", "error-template-*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	customTemplate := `<html><body><h1>Custom Error: {{ .statusName }}</h1><p>{{ .description }}</p></body></html>`
	_, err = tempFile.WriteString(customTemplate)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()
	
	config := &ErrorPageConfig{
		FilePath: tempFile.Name(),
	}
	
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "text/html")
	
	data := map[string]interface{}{
		"statusCode": http.StatusForbidden,
		"statusName": "Forbidden",
		"description": "Access forbidden",
	}
	
	WriteError(logger, config, recorder, req, data)
	
	if recorder.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
	
	body := recorder.Body.String()
	if !strings.Contains(body, "Custom Error: Forbidden") {
		t.Errorf("Expected body to contain custom template content")
	}
	
	if !strings.Contains(body, "Access forbidden") {
		t.Errorf("Expected body to contain description")
	}
}

func TestWriteProblemDetail(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	problem := ProblemDetails{
		Type:   "https://example.com/errors/test",
		Title:  "Test Error",
		Detail: "This is a test error",
	}
	
	recorder := httptest.NewRecorder()
	
	writeProblemDetail(logger, problem, recorder, http.StatusBadRequest)
	
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json+problem" {
		t.Errorf("Expected JSON problem content type, got %s", contentType)
	}
	
	body := recorder.Body.String()
	if !strings.Contains(body, "Test Error") {
		t.Errorf("Expected body to contain 'Test Error'")
	}
	
	if !strings.Contains(body, "This is a test error") {
		t.Errorf("Expected body to contain 'This is a test error'")
	}
}

func TestRenderPage_DefaultTemplate(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	config := &ErrorPageConfig{}
	
	data := map[string]interface{}{
		"statusCode": http.StatusNotFound,
		"statusName": "Not Found",
		"description": "The requested resource was not found",
	}
	
	html, err := renderPage(logger, config, data)
	if err != nil {
		t.Fatalf("renderPage failed: %v", err)
	}
	
	if !strings.Contains(html, "Not Found") {
		t.Errorf("Expected HTML to contain 'Not Found'")
	}
	
	if !strings.Contains(html, "404") {
		t.Errorf("Expected HTML to contain status code '404'")
	}
	
	if !strings.Contains(html, "The requested resource was not found") {
		t.Errorf("Expected HTML to contain description")
	}
}

func TestRenderPage_InvalidTemplate(t *testing.T) {
	logger := logging.CreateLogger(logging.LevelDebug)
	
	// Create a temporary file with invalid template syntax
	tempFile, err := os.CreateTemp("", "invalid-template-*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	
	invalidTemplate := `<html><body>{{ .invalidSyntax }`
	_, err = tempFile.WriteString(invalidTemplate)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()
	
	config := &ErrorPageConfig{
		FilePath: tempFile.Name(),
	}
	
	data := map[string]interface{}{
		"statusCode": http.StatusInternalServerError,
		"statusName": "Internal Server Error",
		"description": "Something went wrong",
	}
	
	_, err = renderPage(logger, config, data)
	if err == nil {
		t.Errorf("Expected renderPage to fail with invalid template")
	}
}