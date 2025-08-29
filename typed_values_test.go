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

	t.Run("string keys and typed keys coexist", func(t *testing.T) {
		err := goerr.New("test error",
			goerr.V("old_key", "old_value"),  // string key
			goerr.TV(userIDKey, "new_value"), // typed key
		)

		// Both values should be accessible via Values()
		values := goerr.Values(err)
		if values == nil {
			t.Fatal("Values() returned nil")
		}

		if values["old_key"] != "old_value" {
			t.Errorf("Values()['old_key'] = %v, want %v", values["old_key"], "old_value")
		}

		if values["user_id"] != "new_value" {
			t.Errorf("Values()['user_id'] = %v, want %v", values["user_id"], "new_value")
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

		// All values should be accessible via Values()
		values := goerr.Values(wrappedErr)
		if values == nil {
			t.Fatal("Values() returned nil")
		}

		expected := map[string]any{
			"legacy_key":     "legacy_value",
			"user_id":        "user123",
			"another_legacy": 123,
			"request_id":     int64(456),
		}

		for key, expectedValue := range expected {
			if values[key] != expectedValue {
				t.Errorf("Values()['%s'] = %v, want %v", key, values[key], expectedValue)
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

	t.Run("Values() method includes typed values", func(t *testing.T) {
		err := goerr.New("test error",
			goerr.TV(userIDKey, "user123"),
			goerr.TV(countKey, 42),
		)

		values := goerr.Values(err)
		if values == nil {
			t.Fatal("Values() returned nil")
		}

		// Typed values should be accessible as regular map values
		if values["user_id"] != "user123" {
			t.Errorf("Values()['user_id'] = %v, want %v", values["user_id"], "user123")
		}

		if values["count"] != 42 {
			t.Errorf("Values()['count'] = %v, want %v", values["count"], 42)
		}

		// Type assertions should work for Values() results
		if userID, ok := values["user_id"].(string); !ok || userID != "user123" {
			t.Errorf("Type assertion for values['user_id'] failed: %v, %v", userID, ok)
		}

		if count, ok := values["count"].(int); !ok || count != 42 {
			t.Errorf("Type assertion for values['count'] failed: %v, %v", count, ok)
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
