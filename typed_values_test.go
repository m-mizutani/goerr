package goerr_test

import (
	"fmt"
	"testing"

	"github.com/m-mizutani/goerr/v2"
)

// Test structures for testing different types
type testConfig struct {
	Name  string
	Value int
}

func TestNewTypedKey(t *testing.T) {
	tests := []struct {
		name     string
		keyName  string
		expected string
	}{
		{
			name:     "string key",
			keyName:  "user_id",
			expected: "user_id",
		},
		{
			name:     "empty key",
			keyName:  "",
			expected: "",
		},
		{
			name:     "complex key name",
			keyName:  "complex_key_name_123",
			expected: "complex_key_name_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := goerr.NewTypedKey[string](tt.keyName)
			if key.Name() != tt.expected {
				t.Errorf("NewTypedKey().Name() = %v, want %v", key.Name(), tt.expected)
			}
			if key.String() != tt.expected {
				t.Errorf("NewTypedKey().String() = %v, want %v", key.String(), tt.expected)
			}
		})
	}
}

func TestTypedValueOption(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "string value",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[string]("user_id")
				err := goerr.New("test error", goerr.TypedValue(key, "user123"))

				value, ok := goerr.GetTypedValue(err, key)
				if !ok {
					t.Error("GetTypedValue() returned false, want true")
				}
				if value != "user123" {
					t.Errorf("GetTypedValue() = %v, want %v", value, "user123")
				}
			},
		},
		{
			name: "int value",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[int]("count")
				err := goerr.New("test error", goerr.TypedValue(key, 42))

				value, ok := goerr.GetTypedValue(err, key)
				if !ok {
					t.Error("GetTypedValue() returned false, want true")
				}
				if value != 42 {
					t.Errorf("GetTypedValue() = %v, want %v", value, 42)
				}
			},
		},
		{
			name: "struct value",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[testConfig]("config")
				config := testConfig{Name: "test", Value: 123}
				err := goerr.New("test error", goerr.TypedValue(key, config))

				value, ok := goerr.GetTypedValue(err, key)
				if !ok {
					t.Error("GetTypedValue() returned false, want true")
				}
				if value.Name != config.Name || value.Value != config.Value {
					t.Errorf("GetTypedValue() = %v, want %v", value, config)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestTVAlias(t *testing.T) {
	key := goerr.NewTypedKey[string]("user_id")
	err := goerr.New("test error", goerr.TV(key, "user123"))

	value, ok := goerr.GetTypedValue(err, key)
	if !ok {
		t.Error("GetTypedValue() returned false, want true")
	}
	if value != "user123" {
		t.Errorf("GetTypedValue() = %v, want %v", value, "user123")
	}
}

func TestGetTypedValue(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "value exists",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[string]("user_id")
				err := goerr.New("test error", goerr.TV(key, "user123"))

				value, ok := goerr.GetTypedValue(err, key)
				if !ok {
					t.Error("GetTypedValue() returned false, want true")
				}
				if value != "user123" {
					t.Errorf("GetTypedValue() = %v, want %v", value, "user123")
				}
			},
		},
		{
			name: "value not exists",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[string]("user_id")
				err := goerr.New("test error")

				value, ok := goerr.GetTypedValue(err, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != "" {
					t.Errorf("GetTypedValue() = %v, want zero value", value)
				}
			},
		},
		{
			name: "nil error",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[string]("user_id")

				value, ok := goerr.GetTypedValue(nil, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != "" {
					t.Errorf("GetTypedValue() = %v, want zero value", value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestTypedValueTypes(t *testing.T) {
	t.Run("multiple types in same error", func(t *testing.T) {
		userIDKey := goerr.NewTypedKey[string]("user_id")
		countKey := goerr.NewTypedKey[int]("count")
		configKey := goerr.NewTypedKey[*testConfig]("config")

		config := &testConfig{Name: "test", Value: 456}
		err := goerr.New("test error",
			goerr.TV(userIDKey, "user123"),
			goerr.TV(countKey, 42),
			goerr.TV(configKey, config),
		)

		// Test string value
		userID, ok := goerr.GetTypedValue(err, userIDKey)
		if !ok {
			t.Error("GetTypedValue(userIDKey) returned false, want true")
		}
		if userID != "user123" {
			t.Errorf("GetTypedValue(userIDKey) = %v, want %v", userID, "user123")
		}

		// Test int value
		count, ok := goerr.GetTypedValue(err, countKey)
		if !ok {
			t.Error("GetTypedValue(countKey) returned false, want true")
		}
		if count != 42 {
			t.Errorf("GetTypedValue(countKey) = %v, want %v", count, 42)
		}

		// Test pointer value
		configValue, ok := goerr.GetTypedValue(err, configKey)
		if !ok {
			t.Error("GetTypedValue(configKey) returned false, want true")
		}
		if configValue != config {
			t.Errorf("GetTypedValue(configKey) = %v, want %v", configValue, config)
		}
	})
}

func TestTypedValuePropagation(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	requestIDKey := goerr.NewTypedKey[int64]("request_id")

	t.Run("value propagation from wrapped error", func(t *testing.T) {
		baseErr := goerr.New("base error", goerr.TV(userIDKey, "user123"))
		wrappedErr := goerr.Wrap(baseErr, "wrapped error", goerr.TV(requestIDKey, int64(456)))

		// Should get value from base error
		userID, ok := goerr.GetTypedValue(wrappedErr, userIDKey)
		if !ok {
			t.Error("GetTypedValue(userIDKey) returned false, want true")
		}
		if userID != "user123" {
			t.Errorf("GetTypedValue(userIDKey) = %v, want %v", userID, "user123")
		}

		// Should get value from wrapped error
		requestID, ok := goerr.GetTypedValue(wrappedErr, requestIDKey)
		if !ok {
			t.Error("GetTypedValue(requestIDKey) returned false, want true")
		}
		if requestID != int64(456) {
			t.Errorf("GetTypedValue(requestIDKey) = %v, want %v", requestID, int64(456))
		}
	})
}

func TestTypedValueOverride(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")

	t.Run("value override in wrapped error", func(t *testing.T) {
		baseErr := goerr.New("base error", goerr.TV(userIDKey, "original_user"))
		wrappedErr := goerr.Wrap(baseErr, "wrapped error", goerr.TV(userIDKey, "new_user"))

		// Should get the overridden value from the upper level
		userID, ok := goerr.GetTypedValue(wrappedErr, userIDKey)
		if !ok {
			t.Error("GetTypedValue() returned false, want true")
		}
		if userID != "new_user" {
			t.Errorf("GetTypedValue() = %v, want %v", userID, "new_user")
		}
	})
}

func TestTypedValueMerging(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	requestIDKey := goerr.NewTypedKey[int64]("request_id")
	sessionKey := goerr.NewTypedKey[string]("session_id")

	t.Run("different keys merge correctly", func(t *testing.T) {
		baseErr := goerr.New("base error",
			goerr.TV(userIDKey, "user123"),
			goerr.TV(sessionKey, "session789"),
		)
		wrappedErr := goerr.Wrap(baseErr, "wrapped error",
			goerr.TV(requestIDKey, int64(456)),
		)

		// All values should be accessible
		userID, ok := goerr.GetTypedValue(wrappedErr, userIDKey)
		if !ok || userID != "user123" {
			t.Errorf("GetTypedValue(userIDKey) = %v, %v, want %v, true", userID, ok, "user123")
		}

		requestID, ok := goerr.GetTypedValue(wrappedErr, requestIDKey)
		if !ok || requestID != int64(456) {
			t.Errorf("GetTypedValue(requestIDKey) = %v, %v, want %v, true", requestID, ok, int64(456))
		}

		sessionID, ok := goerr.GetTypedValue(wrappedErr, sessionKey)
		if !ok || sessionID != "session789" {
			t.Errorf("GetTypedValue(sessionKey) = %v, %v, want %v, true", sessionID, ok, "session789")
		}
	})
}

func TestBackwardCompatibility(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")

	t.Run("string keys and typed keys separated", func(t *testing.T) {
		err := goerr.New("test error",
			goerr.V("old_key", "old_value"),  // string key
			goerr.TV(userIDKey, "new_value"), // typed key
		)

		// String values should only be accessible via Values()
		values := goerr.Values(err)
		if values == nil {
			t.Fatal("Values() returned nil")
		}

		if values["old_key"] != "old_value" {
			t.Errorf("Values()['old_key'] = %v, want %v", values["old_key"], "old_value")
		}

		// String values should NOT contain typed values
		if _, exists := values["user_id"]; exists {
			t.Error("Values() should not contain typed values")
		}

		// Typed values should only be accessible via TypedValues()
		typedValues := goerr.TypedValues(err)
		if typedValues == nil {
			t.Fatal("TypedValues() returned nil")
		}

		if typedValues["user_id"] != "new_value" {
			t.Errorf("TypedValues()['user_id'] = %v, want %v", typedValues["user_id"], "new_value")
		}

		// Typed values should NOT contain string values
		if _, exists := typedValues["old_key"]; exists {
			t.Error("TypedValues() should not contain string values")
		}

		// Typed value should be accessible via GetTypedValue
		typedValue, ok := goerr.GetTypedValue(err, userIDKey)
		if !ok {
			t.Error("GetTypedValue() returned false, want true")
		}
		if typedValue != "new_value" {
			t.Errorf("GetTypedValue() = %v, want %v", typedValue, "new_value")
		}
	})
}

func TestMixedKeyTypes(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	requestIDKey := goerr.NewTypedKey[int64]("request_id")

	t.Run("mixed string and typed keys in error chain", func(t *testing.T) {
		baseErr := goerr.New("base error",
			goerr.V("legacy_key", "legacy_value"),
			goerr.TV(userIDKey, "user123"),
		)
		wrappedErr := goerr.Wrap(baseErr, "wrapped error",
			goerr.V("another_legacy", 123),
			goerr.TV(requestIDKey, int64(456)),
		)

		// String values should be accessible via Values()
		values := goerr.Values(wrappedErr)
		if values == nil {
			t.Fatal("Values() returned nil")
		}

		expectedStringValues := map[string]any{
			"legacy_key":     "legacy_value",
			"another_legacy": 123,
		}

		for key, expectedValue := range expectedStringValues {
			if values[key] != expectedValue {
				t.Errorf("Values()['%s'] = %v, want %v", key, values[key], expectedValue)
			}
		}

		// Typed values should be accessible via TypedValues()
		typedValues := goerr.TypedValues(wrappedErr)
		if typedValues == nil {
			t.Fatal("TypedValues() returned nil")
		}

		expectedTypedValues := map[string]any{
			"user_id":    "user123",
			"request_id": int64(456),
		}

		for key, expectedValue := range expectedTypedValues {
			if typedValues[key] != expectedValue {
				t.Errorf("TypedValues()['%s'] = %v, want %v", key, typedValues[key], expectedValue)
			}
		}

		// Typed values should be accessible via GetTypedValue
		userID, ok := goerr.GetTypedValue(wrappedErr, userIDKey)
		if !ok || userID != "user123" {
			t.Errorf("GetTypedValue(userIDKey) = %v, %v, want %v, true", userID, ok, "user123")
		}

		requestID, ok := goerr.GetTypedValue(wrappedErr, requestIDKey)
		if !ok || requestID != int64(456) {
			t.Errorf("GetTypedValue(requestIDKey) = %v, %v, want %v, true", requestID, ok, int64(456))
		}
	})
}

func TestValuesMethod(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	countKey := goerr.NewTypedKey[int]("count")

	t.Run("TypedValues() method excludes string values", func(t *testing.T) {
		err := goerr.New("test error",
			goerr.TV(userIDKey, "user123"),
			goerr.TV(countKey, 42),
		)

		// Values() should be empty since no string keys were used
		values := goerr.Values(err)
		if values == nil {
			t.Fatal("Values() returned nil")
		}
		if len(values) != 0 {
			t.Errorf("Values() should be empty, got %v", values)
		}

		// TypedValues() should contain the typed values
		typedValues := goerr.TypedValues(err)
		if typedValues == nil {
			t.Fatal("TypedValues() returned nil")
		}

		if typedValues["user_id"] != "user123" {
			t.Errorf("TypedValues()['user_id'] = %v, want %v", typedValues["user_id"], "user123")
		}

		if typedValues["count"] != 42 {
			t.Errorf("TypedValues()['count'] = %v, want %v", typedValues["count"], 42)
		}

		// Type assertions should work for TypedValues() results
		if userID, ok := typedValues["user_id"].(string); !ok || userID != "user123" {
			t.Errorf("Type assertion for typedValues['user_id'] failed: %v, %v", userID, ok)
		}

		if count, ok := typedValues["count"].(int); !ok || count != 42 {
			t.Errorf("Type assertion for typedValues['count'] failed: %v, %v", count, ok)
		}
	})
}

func TestTypedValueNotFound(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	requestIDKey := goerr.NewTypedKey[int64]("request_id")

	t.Run("key not found in error", func(t *testing.T) {
		err := goerr.New("test error", goerr.TV(userIDKey, "user123"))

		// Request different key that doesn't exist
		value, ok := goerr.GetTypedValue(err, requestIDKey)
		if ok {
			t.Error("GetTypedValue() returned true, want false")
		}
		if value != 0 {
			t.Errorf("GetTypedValue() = %v, want zero value (0)", value)
		}
	})

	t.Run("key not found in error chain", func(t *testing.T) {
		baseErr := goerr.New("base error", goerr.TV(userIDKey, "user123"))
		wrappedErr := goerr.Wrap(baseErr, "wrapped error")

		// Request key that doesn't exist in entire chain
		value, ok := goerr.GetTypedValue(wrappedErr, requestIDKey)
		if ok {
			t.Error("GetTypedValue() returned true, want false")
		}
		if value != 0 {
			t.Errorf("GetTypedValue() = %v, want zero value (0)", value)
		}
	})

	t.Run("non-goerr error", func(t *testing.T) {
		err := fmt.Errorf("standard error")

		value, ok := goerr.GetTypedValue(err, userIDKey)
		if ok {
			t.Error("GetTypedValue() returned true, want false")
		}
		if value != "" {
			t.Errorf("GetTypedValue() = %v, want zero value", value)
		}
	})
}

func TestTypedValueTypeConflict(t *testing.T) {
	t.Run("same key name different types", func(t *testing.T) {
		stringKey := goerr.NewTypedKey[string]("same_key")
		intKey := goerr.NewTypedKey[int]("same_key")

		// Set string value
		err := goerr.New("test error", goerr.TV(stringKey, "string_value"))

		// Try to get as int (should fail due to type assertion)
		value, ok := goerr.GetTypedValue(err, intKey)
		if ok {
			t.Error("GetTypedValue() returned true, want false for type mismatch")
		}
		if value != 0 {
			t.Errorf("GetTypedValue() = %v, want zero value (0)", value)
		}

		// Should still work with correct type
		strValue, ok := goerr.GetTypedValue(err, stringKey)
		if !ok {
			t.Error("GetTypedValue() returned false, want true")
		}
		if strValue != "string_value" {
			t.Errorf("GetTypedValue() = %v, want %v", strValue, "string_value")
		}
	})

	t.Run("type conflict in error chain", func(t *testing.T) {
		stringKey := goerr.NewTypedKey[string]("conflict_key")
		intKey := goerr.NewTypedKey[int]("conflict_key")

		// Base error with string value
		baseErr := goerr.New("base error", goerr.TV(stringKey, "base_value"))
		// Wrapped error with int value (same key name)
		wrappedErr := goerr.Wrap(baseErr, "wrapped error", goerr.TV(intKey, 42))

		// Should get int value (upper level priority)
		intValue, ok := goerr.GetTypedValue(wrappedErr, intKey)
		if !ok {
			t.Error("GetTypedValue(intKey) returned false, want true")
		}
		if intValue != 42 {
			t.Errorf("GetTypedValue(intKey) = %v, want %v", intValue, 42)
		}

		// Should NOT get string value due to type conflict at upper level
		stringValue, ok := goerr.GetTypedValue(wrappedErr, stringKey)
		if ok {
			t.Error("GetTypedValue(stringKey) returned true, want false due to type conflict at upper level")
		}
		if stringValue != "" {
			t.Errorf("GetTypedValue(stringKey) = %v, want zero value", stringValue)
		}
	})
}

func TestTypedValueZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name: "zero value for string",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[string]("missing")
				err := goerr.New("test error")

				value, ok := goerr.GetTypedValue(err, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != "" {
					t.Errorf("GetTypedValue() = %q, want empty string", value)
				}
			},
		},
		{
			name: "zero value for int",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[int]("missing")
				err := goerr.New("test error")

				value, ok := goerr.GetTypedValue(err, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != 0 {
					t.Errorf("GetTypedValue() = %v, want 0", value)
				}
			},
		},
		{
			name: "zero value for pointer",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[*testConfig]("missing")
				err := goerr.New("test error")

				value, ok := goerr.GetTypedValue(err, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != nil {
					t.Errorf("GetTypedValue() = %v, want nil", value)
				}
			},
		},
		{
			name: "zero value for slice",
			testFunc: func(t *testing.T) {
				key := goerr.NewTypedKey[[]string]("missing")
				err := goerr.New("test error")

				value, ok := goerr.GetTypedValue(err, key)
				if ok {
					t.Error("GetTypedValue() returned true, want false")
				}
				if value != nil {
					t.Errorf("GetTypedValue() = %v, want nil slice", value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}

func TestTypedValueClone(t *testing.T) {
	userIDKey := goerr.NewTypedKey[string]("user_id")
	countKey := goerr.NewTypedKey[int]("count")

	t.Run("values clone correctly in error copy", func(t *testing.T) {
		originalErr := goerr.New("original error",
			goerr.TV(userIDKey, "user123"),
			goerr.TV(countKey, 42),
		)

		// Use Wrap which internally uses copy() method
		wrappedErr := originalErr.Wrap(nil, goerr.TV(userIDKey, "user456"))

		// Original error should still have original values
		if userID, ok := goerr.GetTypedValue(originalErr, userIDKey); !ok || userID != "user123" {
			t.Errorf("Original error userID = %v, %v, want %v, true", userID, ok, "user123")
		}

		if count, ok := goerr.GetTypedValue(originalErr, countKey); !ok || count != 42 {
			t.Errorf("Original error count = %v, %v, want %v, true", count, ok, 42)
		}

		// Wrapped error should have modified userID but same count
		if userID, ok := goerr.GetTypedValue(wrappedErr, userIDKey); !ok || userID != "user456" {
			t.Errorf("Wrapped error userID = %v, %v, want %v, true", userID, ok, "user456")
		}

		if count, ok := goerr.GetTypedValue(wrappedErr, countKey); !ok || count != 42 {
			t.Errorf("Wrapped error count = %v, %v, want %v, true", count, ok, 42)
		}
	})

	t.Run("modifying cloned values does not affect original", func(t *testing.T) {
		configKey := goerr.NewTypedKey[map[string]string]("config")
		originalConfig := map[string]string{"key": "value"}

		originalErr := goerr.New("original error",
			goerr.TV(configKey, originalConfig),
		)

		// Get the config from original error
		retrievedConfig, ok := goerr.GetTypedValue(originalErr, configKey)
		if !ok {
			t.Fatal("Failed to retrieve config from original error")
		}

		// Modify the retrieved config (this modifies the actual map reference)
		retrievedConfig["key"] = "modified"

		// Check if original error's config is also modified (it should be, since we only do shallow copy)
		checkConfig, ok := goerr.GetTypedValue(originalErr, configKey)
		if !ok {
			t.Fatal("Failed to retrieve config for verification")
		}

		if checkConfig["key"] != "modified" {
			t.Error("Expected shallow copy behavior - map reference should be shared")
		}
	})
}

func ExampleNewTypedKey() {
	// Define typed keys at the package level for reuse.
	var UserIDKey = goerr.NewTypedKey[string]("user_id")
	var RequestIDKey = goerr.NewTypedKey[int]("request_id")

	// Attach typed values when creating an error.
	err := goerr.New("request failed",
		goerr.TV(UserIDKey, "blue"),
		goerr.TV(RequestIDKey, 12345),
	)

	// Retrieve the typed value later.
	if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
		// The retrieved value has the correct type (string), no assertion needed.
		fmt.Printf("User ID: %s\n", userID)
	}

	if reqID, ok := goerr.GetTypedValue(err, RequestIDKey); ok {
		fmt.Printf("Request ID: %d\n", reqID)
	}

	// Output:
	// User ID: blue
	// Request ID: 12345
}
