package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

func main() {
	client, err := langfuse.NewClient(langfuse.Config{
		PublicKey: "pk-lf-your-public-key",
		SecretKey: "sk-lf-your-secret-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Error Handling Examples ===")
	fmt.Println()

	// Example 1: Handling API errors
	fmt.Println("Example 1: Handling API Errors")
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name: "error-handling-demo",
	})
	if err != nil {
		// Check if it's an API error
		if langfuse.IsAPIError(err) {
			apiErr := err.(*langfuse.APIError)
			fmt.Printf("API Error - Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
			if apiErr.Body != "" {
				fmt.Printf("Response body: %s\n", apiErr.Body)
			}
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
		return
	}
	fmt.Printf("Trace created: %s\n\n", trace.ID)

	// Example 2: Capturing errors in observations
	fmt.Println("Example 2: Capturing Errors in Observations")
	span, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "risky-operation", map[string]interface{}{
		"operation": "data-processing",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer span.End()

	// Simulate an operation that might fail
	err = simulateRiskyOperation()
	if err != nil {
		// Update span with error information
		errMsg := err.Error()
		errorLevel := langfuse.LevelError
		span.Update(langfuse.SpanUpdate{
			StatusMessage: &errMsg,
			Level:         &errorLevel,
			Output: map[string]interface{}{
				"error":   err.Error(),
				"success": false,
			},
		})
		fmt.Printf("Operation failed: %v\n", err)
	} else {
		span.Update(langfuse.SpanUpdate{
			Output: map[string]interface{}{
				"success": true,
			},
		})
		fmt.Println("Operation succeeded")
		fmt.Println()
	}

	// Example 3: Error handling with context cancellation
	fmt.Println("Example 3: Context Cancellation")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Try to create an observation with timeout
	_, err = client.CreateSpan(ctxWithTimeout, langfuse.Span{
		TraceID: trace.ID,
		Name:    "timeout-test",
	})
	if err != nil {
		if ctxWithTimeout.Err() == context.DeadlineExceeded {
			fmt.Println("Request timed out (expected)")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}

	// Example 4: Validating required fields before API calls
	fmt.Println("\nExample 4: Validating Required Fields")
	score := langfuse.Score{
		TraceID: trace.ID,
		// Name is required - this will be validated by the SDK
		Name:  "", // Empty name will cause validation error
		Value: 0.9,
	}
	_, err = client.Score(ctx, score)
	if err != nil {
		fmt.Printf("Validation error (expected): %v\n", err)
	}

	// Example 5: Graceful shutdown
	fmt.Println("\nExample 5: Graceful Shutdown")
	fmt.Println("Always call Shutdown() before exiting to ensure all data is sent")
	defer func() {
		if err := client.Shutdown(); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		} else {
			fmt.Println("Client shutdown successfully")
		}
	}()

	fmt.Println("\nAll error handling examples completed!")
}

func simulateRiskyOperation() error {
	// Simulate a 50% chance of failure
	if time.Now().Unix()%2 == 0 {
		return fmt.Errorf("simulated operation failure")
	}
	return nil
}
