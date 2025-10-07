package predicate

import (
	"testing"
)

func TestGetStringMapValue_StringToStringMap(t *testing.T) {
	testMap := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "",
	}
	
	// Test existing key
	result, err := GetStringMapValue(testMap, "key1")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	if result != "value1" {
		t.Errorf("Expected 'value1', got '%v'", result)
	}
	
	// Test another existing key
	result, err = GetStringMapValue(testMap, "key2")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	if result != "value2" {
		t.Errorf("Expected 'value2', got '%v'", result)
	}
	
	// Test key with empty value
	result, err = GetStringMapValue(testMap, "key3")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	if result != "" {
		t.Errorf("Expected empty string, got '%v'", result)
	}
	
	// Test non-existent key
	result, err = GetStringMapValue(testMap, "nonexistent")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	if result != "" {
		t.Errorf("Expected empty string for non-existent key, got '%v'", result)
	}
}

func TestGetStringMapValue_StringToStringSliceMap(t *testing.T) {
	testMap := map[string][]string{
		"headers": {"Authorization", "Content-Type"},
		"methods": {"GET", "POST", "PUT"},
		"empty":   {},
	}
	
	// Test existing key with multiple values
	result, err := GetStringMapValue(testMap, "headers")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	headers, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}
	
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}
	
	expectedHeaders := []string{"Authorization", "Content-Type"}
	for i, expected := range expectedHeaders {
		if i >= len(headers) || headers[i] != expected {
			t.Errorf("Expected header[%d] to be '%s', got '%s'", i, expected, headers[i])
		}
	}
	
	// Test another existing key
	result, err = GetStringMapValue(testMap, "methods")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	methods, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}
	
	if len(methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(methods))
	}
	
	// Test key with empty slice
	result, err = GetStringMapValue(testMap, "empty")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	empty, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}
	
	if len(empty) != 0 {
		t.Errorf("Expected empty slice, got %v", empty)
	}
	
	// Test non-existent key
	result, err = GetStringMapValue(testMap, "nonexistent")
	if err != nil {
		t.Fatalf("GetStringMapValue failed: %v", err)
	}
	
	if result == nil {
		t.Errorf("Expected empty slice for non-existent key, got nil")
	}
}

func TestGetStringMapValue_EmptyMaps(t *testing.T) {
	// Test empty string map
	emptyStringMap := map[string]string{}
	result, err := GetStringMapValue(emptyStringMap, "any")
	if err != nil {
		t.Fatalf("GetStringMapValue failed for empty string map: %v", err)
	}
	
	if result != "" {
		t.Errorf("Expected empty string for empty map, got '%v'", result)
	}
	
	// Test empty string slice map
	emptySliceMap := map[string][]string{}
	result, err = GetStringMapValue(emptySliceMap, "any")
	if err != nil {
		t.Fatalf("GetStringMapValue failed for empty slice map: %v", err)
	}
	
	slice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}
	
	if slice != nil {
		t.Errorf("Expected nil slice for empty map, got %v", slice)
	}
}

func TestGetStringMapValue_InvalidKeyType(t *testing.T) {
	testMap := map[string]string{
		"key1": "value1",
	}
	
	// Test with non-string key
	_, err := GetStringMapValue(testMap, 123)
	if err == nil {
		t.Errorf("Expected error for non-string key")
	}
	
	expectedError := "only string keys are supported"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetStringMapValue_InvalidMapType(t *testing.T) {
	// Test with unsupported map type
	invalidMap := map[string]int{
		"key1": 42,
	}
	
	result, err := GetStringMapValue(invalidMap, "key1")
	if err == nil {
		t.Errorf("Expected error for unsupported map type")
	}
	
	// Should return nil for unsupported map types
	if result != nil {
		t.Errorf("Expected nil for unsupported map type, got %v", result)
	}
}

func TestGetStringMapValue_NilMap(t *testing.T) {
	// Test with nil string map
	var nilStringMap map[string]string
	result, err := GetStringMapValue(nilStringMap, "any")
	if err != nil {
		t.Fatalf("GetStringMapValue failed for nil string map: %v", err)
	}
	
	if result != "" {
		t.Errorf("Expected empty string for nil map, got '%v'", result)
	}
	
	// Test with nil string slice map
	var nilSliceMap map[string][]string
	result, err = GetStringMapValue(nilSliceMap, "any")
	if err != nil {
		t.Fatalf("GetStringMapValue failed for nil slice map: %v", err)
	}
	
	slice, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}
	
	if slice != nil {
		t.Errorf("Expected nil slice for nil map, got %v", slice)
	}
}

func TestGetStringMapValue_OtherTypes(t *testing.T) {
	// Test with completely different type
	notAMap := "this is not a map"
	result, err := GetStringMapValue(notAMap, "key")
	if err == nil {
		t.Errorf("Expected error for non-map type")
	}
	
	if result != nil {
		t.Errorf("Expected nil for non-map type, got %v", result)
	}
	
	// Test with slice instead of map
	slice := []string{"item1", "item2"}
	result, err = GetStringMapValue(slice, "key")
	if err == nil {
		t.Errorf("Expected error for slice type")
	}
	
	if result != nil {
		t.Errorf("Expected nil for slice type, got %v", result)
	}
}