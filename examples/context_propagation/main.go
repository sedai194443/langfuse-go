package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

// simulateServiceCall simulates calling another service
func simulateServiceCall(ctx context.Context, serviceName string) error {
	// Get trace context from incoming context
	traceCtx, ok := langfuse.GetTraceContext(ctx)
	if !ok {
		return fmt.Errorf("no trace context found")
	}

	fmt.Printf("Service '%s' received trace context:\n", serviceName)
	fmt.Printf("  Trace ID: %s\n", traceCtx.TraceID)
	fmt.Printf("  Parent Span ID: %s\n", traceCtx.SpanID)

	// This service would propagate the context to downstream services
	// or create its own observations within the same trace
	return nil
}

func main() {
	client, err := langfuse.NewClient(langfuse.Config{
		PublicKey: "pk-lf-your-public-key",
		SecretKey: "sk-lf-your-secret-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Context Propagation Example ===")
	fmt.Println()

	// Scenario 1: Starting a new trace and propagating it
	fmt.Println("Scenario 1: New Trace Propagation")
	rootSpan, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "api-gateway", map[string]interface{}{
		"endpoint": "/api/users",
		"method":   "GET",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rootSpan.End()

	// Propagate context to downstream service
	if err := simulateServiceCall(rootSpan.Context(), "user-service"); err != nil {
		log.Fatal(err)
	}

	// Create child observation in the propagated context
	childSpan, err := rootSpan.StartChildObservation(
		langfuse.ObservationTypeSpan,
		"database-query",
		map[string]interface{}{
			"query": "SELECT * FROM users",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	childSpan.End()

	fmt.Println()

	// Scenario 2: Joining an existing trace from upstream service
	fmt.Println("Scenario 2: Joining Existing Trace")
	existingTraceID := "abcdef1234567890abcdef1234567890"
	existingParentSpanID := "fedcba0987654321"

	// Create trace context from upstream service
	upstreamTraceCtx := langfuse.TraceContext{
		TraceID:      existingTraceID,
		SpanID:       existingParentSpanID,
		ParentSpanID: "",
	}
	downstreamCtx := langfuse.WithTraceContext(ctx, upstreamTraceCtx)

	// Create observation that joins the existing trace
	downstreamSpan, err := client.StartObservation(
		downstreamCtx,
		langfuse.ObservationTypeSpan,
		"downstream-service",
		map[string]interface{}{
			"service": "payment-service",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer downstreamSpan.End()

	fmt.Printf("Downstream service trace ID: %s (matches upstream: %v)\n",
		downstreamSpan.TraceID,
		downstreamSpan.TraceID == existingTraceID)

	// Scenario 3: Getting current trace ID for logging/correlation
	fmt.Println("\nScenario 3: Trace ID for Logging")
	span, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "logging-demo", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer span.End()

	// Get trace ID for correlation in application logs
	if traceID, ok := langfuse.GetCurrentTraceID(span.Context()); ok {
		fmt.Printf("Application log: [TraceID: %s] Processing request\n", traceID)
	}
	if obsID, ok := langfuse.GetCurrentObservationID(span.Context()); ok {
		fmt.Printf("Application log: [ObservationID: %s] Operation started\n", obsID)
	}

	// Scenario 4: Cross-service propagation with HTTP headers (conceptual)
	fmt.Println("\nScenario 4: HTTP Header Propagation (Conceptual)")
	fmt.Println("In a real microservices setup, you would:")
	fmt.Println("1. Extract trace context from HTTP headers (traceparent, tracestate)")
	fmt.Println("2. Convert to Langfuse TraceContext")
	fmt.Println("3. Use WithTraceContext to propagate")
	fmt.Println("4. Create observations that join the distributed trace")

	// Example: Creating trace context from external trace ID
	externalTraceID := "req_12345"
	langfuseTraceID := langfuse.CreateTraceID(externalTraceID)
	fmt.Printf("\nExternal request ID: %s\n", externalTraceID)
	fmt.Printf("Langfuse trace ID: %s\n", langfuseTraceID)
	fmt.Println("Use this to correlate external systems with Langfuse traces")

	fmt.Println("\nAll context propagation examples completed!")
}
