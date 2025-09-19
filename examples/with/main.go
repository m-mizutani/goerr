package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
)

// Define typed keys for demonstration
var (
	UserIDKey    = goerr.NewTypedKey[string]("user_id")
	RequestIDKey = goerr.NewTypedKey[string]("request_id")
	ComponentKey = goerr.NewTypedKey[string]("component")
)

func main() {
	fmt.Println("=== With Function Example ===")

	// Basic usage example
	basicWithExample()

	// Middleware usage example
	middlewareExample()

	// Standard error usage example
	standardErrorExample()

	// Preservation example
	preservationExample()
}

func basicWithExample() {
	fmt.Println("\n--- Basic With Usage ---")

	// goerr.With(err, options...) adds context without modifying the original error
	originalErr := goerr.New("database connection failed")

	// Add context to the error without modifying the original
	// Original error remains unchanged, new error with additional context is returned
	enrichedErr := goerr.With(originalErr,
		goerr.TV(UserIDKey, "user123"),
		goerr.TV(ComponentKey, "auth-service"),
	)

	fmt.Printf("Original error: %v\n", originalErr)
	fmt.Printf("Enriched error: %v\n", enrichedErr)

	// Verify original error has no additional context
	if userID, ok := goerr.GetTypedValue(originalErr, UserIDKey); ok {
		fmt.Printf("Original has UserID: %s (unexpected)\n", userID)
	} else {
		fmt.Println("Original error has no UserID (expected)")
	}

	// Verify enriched error has the context
	if userID, ok := goerr.GetTypedValue(enrichedErr, UserIDKey); ok {
		fmt.Printf("Enriched error has UserID: %s\n", userID)
	}
}

func middlewareExample() {
	fmt.Println("\n--- Middleware Usage Example ---")

	// Simulate a web request processing chain
	// Middleware can add request context like user ID, request ID without changing business logic errors

	// Business logic error (no request context)
	businessErr := validateUserData("invalid-email")

	// Middleware adds request context
	requestEnrichedErr := addRequestContext(businessErr, "req-12345", "user789")

	// Another middleware adds component information
	componentEnrichedErr := addComponentContext(requestEnrichedErr, "user-validation-service")

	fmt.Printf("Business error: %v\n", businessErr)
	fmt.Printf("Final enriched error: %v\n", componentEnrichedErr)

	// All context is available in the final error
	if reqID, ok := goerr.GetTypedValue(componentEnrichedErr, RequestIDKey); ok {
		fmt.Printf("Request ID: %s\n", reqID)
	}
	if userID, ok := goerr.GetTypedValue(componentEnrichedErr, UserIDKey); ok {
		fmt.Printf("User ID: %s\n", userID)
	}
	if component, ok := goerr.GetTypedValue(componentEnrichedErr, ComponentKey); ok {
		fmt.Printf("Component: %s\n", component)
	}
}

func standardErrorExample() {
	fmt.Println("\n--- Standard Error Usage ---")

	// Works with both goerr.Error and standard Go errors
	stdErr := errors.New("file not found")

	// goerr.With works with standard errors too
	enrichedStdErr := goerr.With(stdErr,
		goerr.TV(ComponentKey, "file-reader"),
		goerr.TV(UserIDKey, "admin"),
	)

	fmt.Printf("Standard error: %v\n", stdErr)
	fmt.Printf("Enriched standard error: %v\n", enrichedStdErr)

	// Context is available even from standard errors
	if component, ok := goerr.GetTypedValue(enrichedStdErr, ComponentKey); ok {
		fmt.Printf("Component from standard error: %s\n", component)
	}
}

func preservationExample() {
	fmt.Println("\n--- Error Preservation Example ---")

	// Demonstrate that With preserves the original error for errors.Is and errors.As
	originalErr := &ValidationError{Field: "email", Reason: "invalid format"}

	enrichedErr := goerr.With(originalErr,
		goerr.TV(UserIDKey, "user456"),
	)

	// errors.Is still works
	if errors.Is(enrichedErr, originalErr) {
		fmt.Println("errors.Is works with enriched error")
	}

	// errors.As still works
	var validationErr *ValidationError
	if errors.As(enrichedErr, &validationErr) {
		fmt.Printf("errors.As works: Field=%s, Reason=%s\n",
			validationErr.Field, validationErr.Reason)
	}

	// Structured logging with slog.LogValuer
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	fmt.Println("\nStructured logging output:")
	logger.Info("Operation failed", slog.Any("error", enrichedErr))
}

// Helper functions

func validateUserData(email string) error {
	if email == "invalid-email" {
		return goerr.New("email validation failed")
	}
	return nil
}

// Middleware function to add request context
func addRequestContext(err error, requestID, userID string) error {
	if err == nil {
		return nil
	}
	// Middleware adds request-specific context without modifying business logic errors
	return goerr.With(err,
		goerr.TV(RequestIDKey, requestID),
		goerr.TV(UserIDKey, userID),
	)
}

// Middleware function to add component context
func addComponentContext(err error, component string) error {
	if err == nil {
		return nil
	}
	// Component middleware adds service identification
	return goerr.With(err,
		goerr.TV(ComponentKey, component),
	)
}

// Custom error type for demonstration
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field %s: %s", e.Field, e.Reason)
}
