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

	// Create a trace
	ctx := context.Background()
	now := time.Now()
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name:   "example-trace",
		UserID: "user-123",
		Metadata: map[string]interface{}{
			"environment": "production",
			"version":     "1.0.0",
		},
		Tags:      []string{"example", "demo"},
		Timestamp: &now,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created trace: %s\n", trace.ID)

	// Create a generation (LLM call)
	startTime := time.Now()
	generation, err := client.CreateGeneration(ctx, langfuse.Generation{
		TraceID:   trace.ID,
		Name:      "chat-completion",
		Model:     "gpt-4",
		StartTime: &startTime,
		Input: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "What is the capital of France?"},
			},
		},
		Usage: &langfuse.Usage{
			Input:  10,
			Output: 20,
			Total:  30,
			Unit:   "TOKENS",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created generation: %s\n", generation.ID)

	// Simulate some processing time
	time.Sleep(100 * time.Millisecond)

	// Update the generation with output
	endTime := time.Now()
	_, err = client.UpdateGeneration(ctx, generation.ID, langfuse.GenerationUpdate{
		EndTime: &endTime,
		Output: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "assistant", "content": "The capital of France is Paris."},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated generation with output")

	// Create a span
	spanStart := time.Now()
	span, err := client.CreateSpan(ctx, langfuse.Span{
		TraceID:   trace.ID,
		Name:      "database-query",
		StartTime: &spanStart,
		Input: map[string]interface{}{
			"query":  "SELECT * FROM users WHERE id = ?",
			"params": []interface{}{123},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created span: %s\n", span.ID)

	// Update span with output
	spanEnd := time.Now()
	_, err = client.UpdateSpan(ctx, span.ID, langfuse.SpanUpdate{
		EndTime: &spanEnd,
		Output: map[string]interface{}{
			"rows": 1,
			"data": map[string]interface{}{
				"id":   123,
				"name": "John Doe",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated span with output")

	// Create a score
	_, err = client.Score(ctx, langfuse.Score{
		TraceID: trace.ID,
		Name:    "quality",
		Value:   0.95,
		Comment: "High quality response",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created score")

	fmt.Println("\nExample completed successfully!")
}
