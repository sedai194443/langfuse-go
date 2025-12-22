package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

func main() {
	// Step 1: Initialize the Langfuse client
	client, err := langfuse.NewClient(langfuse.Config{
		PublicKey: "pk-lf-your-public-key", // Get from Langfuse dashboard
		SecretKey: "sk-lf-your-secret-key", // Get from Langfuse dashboard
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Step 2: Create a trace (represents a single user request/session)
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name:   "my-first-trace",
		UserID: "user-123",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created trace: %s\n", trace.ID)

	// Step 3: Create a generation (LLM call) within the trace
	startTime := time.Now()
	generation, err := client.CreateGeneration(ctx, langfuse.Generation{
		TraceID: trace.ID,
		Name:    "chat-completion",
		Model:   "gpt-4",
		Input: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello!"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created generation: %s\n", generation.ID)

	// Step 4: Update the generation with output
	endTime := time.Now()
	_, err = client.UpdateGeneration(ctx, generation.ID, langfuse.GenerationUpdate{
		EndTime: &endTime,
		Output: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "assistant", "content": "Hi there!"},
			},
		},
		Usage: &langfuse.Usage{
			Input:  10,
			Output: 20,
			Total:  30,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated generation with output")

	// Step 5: (Optional) Create a score to rate the response
	_, err = client.Score(ctx, langfuse.Score{
		TraceID: trace.ID,
		Name:    "user-satisfaction",
		Value:   0.9,
		Comment: "User was happy with the response",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created score")

	fmt.Println("\n✅ Successfully sent data to Langfuse!")
	fmt.Println("Check your Langfuse dashboard to see the trace.")
}

