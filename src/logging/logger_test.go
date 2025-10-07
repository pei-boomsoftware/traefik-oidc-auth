package logging

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	// Save current stdout
	oldStdout := os.Stdout
	
	// Create a pipe to capture output
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Execute function
	f()
	
	// Close write end and restore stdout
	w.Close()
	os.Stdout = oldStdout
	
	// Read captured output
	output, _ := io.ReadAll(r)
	return string(output)
}

func TestCreateLogger(t *testing.T) {
	logger := CreateLogger(LevelInfo)
	
	if logger.MinLevel != LevelInfo {
		t.Errorf("Expected MinLevel to be %s, got %s", LevelInfo, logger.MinLevel)
	}
}

func TestLogLevels(t *testing.T) {
	expectedLevels := map[string]int{
		LevelError: 1,
		LevelWarn:  2,
		LevelInfo:  3,
		LevelDebug: 4,
	}
	
	for level, expectedValue := range expectedLevels {
		if LogLevels[level] != expectedValue {
			t.Errorf("Expected LogLevels[%s] to be %d, got %d", level, expectedValue, LogLevels[level])
		}
	}
}

func TestShouldLog(t *testing.T) {
	// Test case: minLevel=INFO, level=ERROR should log (ERROR=1, INFO=3, 3>=1)
	if !shouldLog(LevelInfo, LevelError) {
		t.Errorf("Expected ERROR to be logged when minLevel is INFO")
	}
	
	// Test case: minLevel=INFO, level=WARN should log (WARN=2, INFO=3, 3>=2)
	if !shouldLog(LevelInfo, LevelWarn) {
		t.Errorf("Expected WARN to be logged when minLevel is INFO")
	}
	
	// Test case: minLevel=INFO, level=INFO should log (INFO=3, INFO=3, 3>=3)
	if !shouldLog(LevelInfo, LevelInfo) {
		t.Errorf("Expected INFO to be logged when minLevel is INFO")
	}
	
	// Test case: minLevel=INFO, level=DEBUG should not log (DEBUG=4, INFO=3, 3<4)
	if shouldLog(LevelInfo, LevelDebug) {
		t.Errorf("Expected DEBUG to NOT be logged when minLevel is INFO")
	}
	
	// Test case: minLevel=ERROR, level=WARN should not log (WARN=2, ERROR=1, 1<2)
	if shouldLog(LevelError, LevelWarn) {
		t.Errorf("Expected WARN to NOT be logged when minLevel is ERROR")
	}
	
	// Test case: minLevel=DEBUG should log everything
	if !shouldLog(LevelDebug, LevelError) {
		t.Errorf("Expected ERROR to be logged when minLevel is DEBUG")
	}
	if !shouldLog(LevelDebug, LevelWarn) {
		t.Errorf("Expected WARN to be logged when minLevel is DEBUG")
	}
	if !shouldLog(LevelDebug, LevelInfo) {
		t.Errorf("Expected INFO to be logged when minLevel is DEBUG")
	}
	if !shouldLog(LevelDebug, LevelDebug) {
		t.Errorf("Expected DEBUG to be logged when minLevel is DEBUG")
	}
}

func TestShouldLogCaseInsensitive(t *testing.T) {
	// Test case insensitive behavior
	if !shouldLog("info", "ERROR") {
		t.Errorf("Expected case insensitive level matching to work")
	}
	
	if !shouldLog("INFO", "error") {
		t.Errorf("Expected case insensitive level matching to work")
	}
	
	if shouldLog("error", "warn") {
		t.Errorf("Expected case insensitive level matching to work")
	}
}

func TestLoggerLog_ShouldLog(t *testing.T) {
	logger := CreateLogger(LevelInfo)
	
	output := captureOutput(func() {
		logger.Log(LevelError, "Test error message: %s", "param")
	})
	
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected output to contain [ERROR]")
	}
	
	if !strings.Contains(output, "[traefik-oidc-auth]") {
		t.Errorf("Expected output to contain [traefik-oidc-auth]")
	}
	
	if !strings.Contains(output, "Test error message: param") {
		t.Errorf("Expected output to contain formatted message")
	}
}

func TestLoggerLog_ShouldNotLog(t *testing.T) {
	logger := CreateLogger(LevelError)
	
	output := captureOutput(func() {
		logger.Log(LevelInfo, "This should not be logged")
	})
	
	if output != "" {
		t.Errorf("Expected no output when log level is below minimum, got: %s", output)
	}
}

func TestLoggerLog_FormatString(t *testing.T) {
	logger := CreateLogger(LevelDebug)
	
	output := captureOutput(func() {
		logger.Log(LevelInfo, "User %s has %d items", "john", 5)
	})
	
	if !strings.Contains(output, "User john has 5 items") {
		t.Errorf("Expected formatted string 'User john has 5 items' in output: %s", output)
	}
}

func TestLoggerLog_TimestampFormat(t *testing.T) {
	logger := CreateLogger(LevelDebug)
	
	output := captureOutput(func() {
		logger.Log(LevelInfo, "Test message")
	})
	
	// Check that output starts with timestamp format YYYY-MM-DD HH:MM:SS
	parts := strings.Split(output, " ")
	if len(parts) < 2 {
		t.Errorf("Expected output to start with timestamp, got: %s", output)
	}
	
	// Basic timestamp format check (YYYY-MM-DD)
	datePart := parts[0]
	if len(datePart) != 10 || datePart[4] != '-' || datePart[7] != '-' {
		t.Errorf("Expected date format YYYY-MM-DD, got: %s", datePart)
	}
	
	// Basic time format check (HH:MM:SS)
	timePart := parts[1]
	if len(timePart) != 8 || timePart[2] != ':' || timePart[5] != ':' {
		t.Errorf("Expected time format HH:MM:SS, got: %s", timePart)
	}
}

func TestLoggerLog_AllLevels(t *testing.T) {
	logger := CreateLogger(LevelDebug)
	
	testCases := []string{LevelError, LevelWarn, LevelInfo, LevelDebug}
	
	for _, level := range testCases {
		output := captureOutput(func() {
			logger.Log(level, "Test message for %s level", level)
		})
		
		expectedTag := "[" + level + "]"
		if !strings.Contains(output, expectedTag) {
			t.Errorf("Expected output to contain %s for level %s", expectedTag, level)
		}
		
		expectedMessage := "Test message for " + level + " level"
		if !strings.Contains(output, expectedMessage) {
			t.Errorf("Expected output to contain message for level %s", level)
		}
	}
}