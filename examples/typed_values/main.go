package main

import (
	"fmt"

	"github.com/m-mizutani/goerr/v2"
)

// Define typed keys (usually at package level)
var (
	UserIDKey     = goerr.NewTypedKey[string]("user_id")
	RequestIDKey  = goerr.NewTypedKey[int64]("request_id")
	HTTPStatusKey = goerr.NewTypedKey[int]("http_status")
	ConfigKey     = goerr.NewTypedKey[*ServerConfig]("config")
)

// Example configuration struct
type ServerConfig struct {
	Host string
	Port int
}

func main() {
	fmt.Println("=== Typed Values Example ===")

	// Basic usage example
	basicExample()

	// Web application example
	webAppExample()

	// Error chain example
	errorChainExample()

	// Mixed key types example
	mixedKeysExample()
}

func basicExample() {
	fmt.Println("\n--- Basic Usage ---")

	// Create error with typed values
	err := goerr.New("operation failed",
		goerr.TV(UserIDKey, "user123"),
		goerr.TV(RequestIDKey, int64(456)),
	)

	// Retrieve values in a type-safe manner (no type assertions needed)
	if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
		fmt.Printf("User ID: %s\n", userID)
	}

	if requestID, ok := goerr.GetTypedValue(err, RequestIDKey); ok {
		fmt.Printf("Request ID: %d\n", requestID)
	}

	// This would fail at compile time:
	// goerr.TV(UserIDKey, 123) // cannot use 123 as string value
}

func webAppExample() {
	fmt.Println("\n--- Web Application Example ---")

	// Simulate a web request handler
	err := handleRequest("user789", "req_001")
	if err != nil {
		fmt.Printf("Error: %v\n", err)

		// Extract structured information for logging/monitoring
		if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
			fmt.Printf("Failed for user: %s\n", userID)
		}

		if status, ok := goerr.GetTypedValue(err, HTTPStatusKey); ok {
			fmt.Printf("HTTP Status: %d\n", status)
		}
	}
}

func handleRequest(userID, requestID string) error {
	// Simulate user validation failure
	if err := validateUser(userID); err != nil {
		return goerr.Wrap(err, "request handling failed",
			goerr.TV(UserIDKey, userID),
			goerr.TV(RequestIDKey, parseRequestID(requestID)),
			goerr.TV(HTTPStatusKey, 400),
		)
	}
	return nil
}

func validateUser(userID string) error {
	// Simulate validation failure
	return goerr.New("user validation failed",
		goerr.TV(UserIDKey, userID),
		goerr.TV(HTTPStatusKey, 401),
	)
}

func parseRequestID(_ string) int64 {
	// Simplified parsing
	return 12345
}

func errorChainExample() {
	fmt.Println("\n--- Error Chain Example ---")

	config := &ServerConfig{Host: "localhost", Port: 8080}

	// Create a chain of errors with typed values
	baseErr := goerr.New("database connection failed",
		goerr.TV(ConfigKey, config),
	)

	middleErr := goerr.Wrap(baseErr, "service initialization failed",
		goerr.TV(UserIDKey, "admin"),
	)

	topErr := goerr.Wrap(middleErr, "server startup failed",
		goerr.TV(HTTPStatusKey, 500),
	)

	// Values propagate through the error chain
	if cfg, ok := goerr.GetTypedValue(topErr, ConfigKey); ok {
		fmt.Printf("Failed config - Host: %s, Port: %d\n", cfg.Host, cfg.Port)
	}

	if userID, ok := goerr.GetTypedValue(topErr, UserIDKey); ok {
		fmt.Printf("User involved: %s\n", userID)
	}

	if status, ok := goerr.GetTypedValue(topErr, HTTPStatusKey); ok {
		fmt.Printf("HTTP Status: %d\n", status)
	}
}

func mixedKeysExample() {
	fmt.Println("\n--- Mixed Keys Example (Backward Compatibility) ---")

	// Mix old string-based keys with new typed keys
	err := goerr.New("mixed error",
		goerr.V("legacy_key", "legacy_value"), // Old API
		goerr.TV(UserIDKey, "user456"),        // New API
	)

	// Both approaches work
	values := goerr.Values(err)
	fmt.Printf("Legacy value: %v\n", values["legacy_key"])
	fmt.Printf("User ID from Values(): %v\n", values["user_id"])

	// Type-safe access still works
	if userID, ok := goerr.GetTypedValue(err, UserIDKey); ok {
		fmt.Printf("User ID (type-safe): %s\n", userID)
	}
}
