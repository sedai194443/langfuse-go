package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

// Example functions to be wrapped with observe

// Simple function that processes data
func processData(data string) (string, error) {
	time.Sleep(10 * time.Millisecond) // Simulate work
	return fmt.Sprintf("Processed: %s", data), nil
}

// Function with context
func fetchData(ctx context.Context, source string) (map[string]interface{}, error) {
	time.Sleep(20 * time.Millisecond) // Simulate API call
	return map[string]interface{}{
		"data":   fmt.Sprintf("Data from %s", source),
		"source": source,
	}, nil
}

// Function that can error
func riskyOperation(input int) (int, error) {
	time.Sleep(5 * time.Millisecond)
	if input < 0 {
		return 0, fmt.Errorf("input must be positive, got %d", input)
	}
	return input * 2, nil
}

func main() {
	// Initialize the client
	client, err := langfuse.NewClient(langfuse.Config{
		PublicKey: "pk-lf-your-public-key",
		SecretKey: "sk-lf-your-secret-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Observe Wrapper Examples ===\n")

	// Note: The observe wrapper in Go is more limited than Python/JS decorators
	// due to Go's type system. Here are practical alternatives:

	// Example 1: Manual observation wrapping (recommended approach)
	fmt.Println("Example 1: Manual Observation Wrapping")
	manualObservedProcessData := func(data string) (string, error) {
		obs, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "process-data", map[string]interface{}{
			"input": data,
		})
		if err != nil {
			return "", err
		}
		defer obs.End()

		// Execute the actual function
		result, err := processData(data)
		if err != nil {
			obs.Update(langfuse.SpanUpdate{
				StatusMessage: stringPtr(err.Error()),
			})
			return "", err
		}

		// Update with output
		obs.Update(langfuse.SpanUpdate{
			Output: map[string]interface{}{
				"result": result,
			},
		})

		return result, nil
	}

	result, err := manualObservedProcessData("test-data")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Result: %s\n\n", result)

	// Example 2: Helper function for wrapping
	fmt.Println("Example 2: Helper Function Wrapper")
	// Wrap fetchData to match the expected signature
	wrappedFetchData := func(ctx context.Context, source string) (interface{}, error) {
		return fetchData(ctx, source)
	}
	observedFetchData := observeSpan(client, ctx, "fetch-data", wrappedFetchData)
	
	fetchResult, err := observedFetchData(ctx, "API")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fetched: %v\n\n", fetchResult)

	// Example 3: Wrapping with error handling
	fmt.Println("Example 3: Wrapping with Error Handling")
	observedRiskyOp := observeSpanWithError(client, ctx, "risky-operation", riskyOperation)
	
	// Success case
	successResult, err := observedRiskyOp(5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Success result: %d\n", successResult)

	// Error case
	_, err = observedRiskyOp(-1)
	if err != nil {
		fmt.Printf("Expected error: %v\n\n", err)
	}

	// Example 4: Generation observation wrapper
	fmt.Println("Example 4: Generation Observation Wrapper")
	llmCall := func(ctx context.Context, prompt string) (string, error) {
		// Simulate LLM call
		time.Sleep(30 * time.Millisecond)
		return fmt.Sprintf("Response to: %s", prompt), nil
	}

	observedLLMCall := observeGeneration(client, ctx, "llm-call", llmCall)
	
	response, err := observedLLMCall(ctx, "What is Go?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("LLM Response: %s\n\n", response)

	fmt.Println("All observe wrapper examples completed!")
}

// Helper function to wrap a function with span observation
func observeSpan[T any](client *langfuse.Client, ctx context.Context, name string, fn func(context.Context, T) (interface{}, error)) func(context.Context, T) (interface{}, error) {
	return func(ctx context.Context, input T) (interface{}, error) {
		obs, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, name, map[string]interface{}{
			"input": input,
		})
		if err != nil {
			return nil, err
		}
		defer obs.End()

		result, err := fn(obs.Context(), input)
		if err != nil {
			obs.Update(langfuse.SpanUpdate{
				StatusMessage: stringPtr(err.Error()),
			})
			return nil, err
		}

		obs.Update(langfuse.SpanUpdate{
			Output: map[string]interface{}{
				"result": result,
			},
		})

		return result, nil
	}
}

// Helper function to wrap a function with span observation (no context)
func observeSpanWithError[T any, R any](client *langfuse.Client, ctx context.Context, name string, fn func(T) (R, error)) func(T) (R, error) {
	return func(input T) (R, error) {
		var zero R
		obs, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, name, map[string]interface{}{
			"input": input,
		})
		if err != nil {
			return zero, err
		}
		defer obs.End()

		result, err := fn(input)
		if err != nil {
			obs.Update(langfuse.SpanUpdate{
				StatusMessage: stringPtr(err.Error()),
			})
			return zero, err
		}

		obs.Update(langfuse.SpanUpdate{
			Output: map[string]interface{}{
				"result": result,
			},
		})

		return result, nil
	}
}

// Helper function to wrap a function with generation observation
func observeGeneration[T any](client *langfuse.Client, ctx context.Context, name string, fn func(context.Context, T) (string, error)) func(context.Context, T) (string, error) {
	return func(ctx context.Context, input T) (string, error) {
		obs, err := client.StartObservation(ctx, langfuse.ObservationTypeGeneration, name, map[string]interface{}{
			"input": input,
		})
		if err != nil {
			return "", err
		}
		defer obs.End()

		result, err := fn(obs.Context(), input)
		if err != nil {
			obs.Update(langfuse.GenerationUpdate{
				StatusMessage: stringPtr(err.Error()),
			})
			return "", err
		}

		obs.Update(langfuse.GenerationUpdate{
			Output: map[string]interface{}{
				"response": result,
			},
			Usage: &langfuse.Usage{
				Input:  10,
				Output: 20,
				Total:  30,
			},
		})

		return result, nil
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

