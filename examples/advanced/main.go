package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

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

	// Example 1: W3C Trace Context - Creating deterministic trace IDs
	fmt.Println("=== Example 1: W3C Trace Context ===")
	externalRequestID := "req_12345"
	deterministicTraceID := langfuse.CreateTraceID(externalRequestID)
	fmt.Printf("Deterministic trace ID from seed '%s': %s\n", externalRequestID, deterministicTraceID)

	// Same seed produces same ID
	sameTraceID := langfuse.CreateTraceID(externalRequestID)
	fmt.Printf("Same seed produces same ID: %s\n", sameTraceID)

	// Example 2: Context Manager Pattern (Go equivalent of Python's 'with' block)
	fmt.Println("\n=== Example 2: Context Manager Pattern ===")
	rootSpan, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "user-request-pipeline", map[string]interface{}{
		"user_query": "Tell me a joke",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = rootSpan.End() }()

	// Update root span with metadata
	_ = rootSpan.Update(langfuse.SpanUpdate{
		Metadata: map[string]interface{}{
			"environment": "production",
			"version":     "1.0.0",
		},
	})

	// Create a child generation within the root span's context
	generation, err := rootSpan.StartChildObservation(
		langfuse.ObservationTypeGeneration,
		"joke-generation",
		map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Tell me a joke"},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Simulate LLM processing
	time.Sleep(50 * time.Millisecond)

	// Update generation with output and usage
	_ = generation.Update(langfuse.GenerationUpdate{
		Model: stringPtr("gpt-4o"),
		Output: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "assistant", "content": "Why did the span cross the road? To get to the other trace!"},
			},
		},
		Usage: &langfuse.Usage{
			Input:  10,
			Output: 20,
			Total:  30,
			Unit:   "TOKENS",
		},
	})
	_ = generation.End()

	// Update root span with final output
	_ = rootSpan.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"final_joke": "Why did the span cross the road? To get to the other trace!",
		},
	})

	// Example 3: Getting current trace and observation IDs from context
	fmt.Println("\n=== Example 3: Getting Current IDs from Context ===")
	obsCtx := generation.Context()
	if traceID, ok := langfuse.GetCurrentTraceID(obsCtx); ok {
		fmt.Printf("Current trace ID: %s\n", traceID)
	}
	if obsID, ok := langfuse.GetCurrentObservationID(obsCtx); ok {
		fmt.Printf("Current observation ID: %s\n", obsID)
	}

	// Example 4: Manual trace context propagation
	fmt.Println("\n=== Example 4: Manual Trace Context Propagation ===")
	existingTraceID := "abcdef1234567890abcdef1234567890"
	existingParentSpanID := "fedcba0987654321"

	// Create trace context manually
	traceCtx := langfuse.TraceContext{
		TraceID:      existingTraceID,
		SpanID:       existingParentSpanID,
		ParentSpanID: "",
	}
	propagatedCtx := langfuse.WithTraceContext(ctx, traceCtx)

	// Create observation in propagated context (joins existing trace)
	downstreamSpan, err := client.StartObservation(
		propagatedCtx,
		langfuse.ObservationTypeSpan,
		"process-downstream-task",
		map[string]interface{}{
			"task": "process-data",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = downstreamSpan.End() }()

	fmt.Printf("Downstream span trace ID: %s (should match: %s)\n", downstreamSpan.TraceID, existingTraceID)

	// Example 5: Nested observations with automatic parent tracking
	fmt.Println("\n=== Example 5: Nested Observations ===")
	parentSpan, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "parent-operation", map[string]interface{}{
		"operation": "data-processing",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = parentSpan.End() }()

	// Child span 1
	child1, err := parentSpan.StartChildObservation(
		langfuse.ObservationTypeSpan,
		"child-operation-1",
		map[string]interface{}{"step": "validation"},
	)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	_ = child1.End()

	// Child span 2
	child2, err := parentSpan.StartChildObservation(
		langfuse.ObservationTypeSpan,
		"child-operation-2",
		map[string]interface{}{"step": "transformation"},
	)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	_ = child2.End()

	// Example 6: Using different observation types
	fmt.Println("\n=== Example 6: Different Observation Types ===")

	// Create a trace first
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name:   "multi-type-trace",
		UserID: "user-123",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a span
	span, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "database-query", map[string]interface{}{
		"query": "SELECT * FROM users",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = span.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{"rows": 42},
	})
	_ = span.End()

	// Create a generation
	gen, err := client.StartObservation(ctx, langfuse.ObservationTypeGeneration, "llm-call", map[string]interface{}{
		"prompt": "What is AI?",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = gen.Update(langfuse.GenerationUpdate{
		Model: stringPtr("gpt-4"),
		Output: map[string]interface{}{
			"response": "AI is artificial intelligence",
		},
		Usage: &langfuse.Usage{
			Input:  5,
			Output: 10,
			Total:  15,
		},
	})
	_ = gen.End()

	// Create an event
	event, err := client.StartObservation(ctx, langfuse.ObservationTypeEvent, "user-action", map[string]interface{}{
		"action": "button_click",
		"button": "submit",
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = event.End()

	// Example 7: Updating trace with user and session info
	fmt.Println("\n=== Example 7: Updating Trace ===")
	_, err = client.UpdateTrace(ctx, trace.ID, langfuse.TraceUpdate{
		UserID:    stringPtr("user-123"),
		SessionID: stringPtr("session-abc"),
		Tags:      []string{"authenticated-user", "premium"},
		Metadata: map[string]interface{}{
			"plan": "premium",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Example 8: Creating scores for observations
	fmt.Println("\n=== Example 8: Creating Scores ===")
	_, err = client.Score(ctx, langfuse.Score{
		TraceID:       trace.ID,
		ObservationID: gen.ID,
		Name:          "quality",
		Value:         0.95,
		Comment:       "High quality response",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAll examples completed successfully!")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
