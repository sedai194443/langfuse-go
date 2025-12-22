package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AEKurt/langfuse-go"
)

func main() {
	// Initialize the async client with batch processing
	// This is recommended for production use as it queues events
	// and sends them in batches for better performance
	client, err := langfuse.NewAsyncClient(
		langfuse.Config{
			PublicKey: "pk-lf-your-public-key",
			SecretKey: "sk-lf-your-secret-key",
		},
		langfuse.BatchConfig{
			MaxBatchSize:  50,                      // Flush when 50 events are queued
			FlushInterval: 2 * time.Second,         // Or flush every 2 seconds
			MaxRetries:    3,                       // Retry failed requests up to 3 times
			RetryDelay:    500 * time.Millisecond,  // Initial retry delay (exponential backoff)
			QueueSize:     10000,                   // Max events in queue
			ShutdownTimeout: 30 * time.Second,      // Max wait time for shutdown
			OnError: func(err error, events []langfuse.BatchEvent) {
				// Handle failed batches (e.g., log to error monitoring)
				log.Printf("Failed to send %d events: %v", len(events), err)
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Always shutdown gracefully to flush pending events
	defer func() {
		fmt.Println("Shutting down... flushing pending events")
		if err := client.Shutdown(); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
		fmt.Println("Shutdown complete")
	}()

	// Example 1: Basic async trace creation
	fmt.Println("=== Example 1: Basic Async Operations ===")
	
	// CreateTraceAsync returns immediately - the event is queued
	traceID, err := client.CreateTraceAsync(langfuse.Trace{
		Name:      "async-example-trace",
		UserID:    "user-123",
		SessionID: "session-abc",
		Input:     map[string]interface{}{"query": "What is async processing?"},
		Metadata: map[string]interface{}{
			"environment": "production",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queued trace: %s\n", traceID)

	// Create a span async
	spanID, err := client.CreateSpanAsync(langfuse.Span{
		TraceID: traceID,
		Name:    "process-request",
		Input:   map[string]interface{}{"step": "validation"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queued span: %s\n", spanID)

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Update span async
	endTime := time.Now()
	err = client.UpdateSpanAsync(spanID, langfuse.SpanUpdate{
		EndTime: &endTime,
		Output:  map[string]interface{}{"result": "validated"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Queued span update")

	// Create generation async
	genID, err := client.CreateGenerationAsync(langfuse.Generation{
		TraceID: traceID,
		Name:    "llm-completion",
		Model:   "gpt-4",
		Input: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Explain async processing"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queued generation: %s\n", genID)

	// Update generation async
	model := "gpt-4"
	err = client.UpdateGenerationAsync(genID, langfuse.GenerationUpdate{
		Model: &model,
		Output: map[string]interface{}{
			"response": "Async processing queues work for later execution...",
		},
		Usage: &langfuse.Usage{
			Input:  15,
			Output: 25,
			Total:  40,
			Unit:   "TOKENS",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Queued generation update")

	// Create event async
	eventID, err := client.CreateEventAsync(langfuse.Event{
		TraceID: traceID,
		Name:    "user-feedback",
		Input:   map[string]interface{}{"action": "thumbs_up"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queued event: %s\n", eventID)

	// Create score async
	scoreID, err := client.ScoreAsync(langfuse.Score{
		TraceID: traceID,
		Name:    "user_satisfaction",
		Value:   0.9,
		Comment: "User gave positive feedback",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Queued score: %s\n", scoreID)

	// Check queue length
	fmt.Printf("Current queue length: %d\n", client.QueueLength())

	// Example 2: High-volume concurrent writes
	fmt.Println("\n=== Example 2: High-Volume Concurrent Writes ===")
	
	var wg sync.WaitGroup
	numTraces := 100

	start := time.Now()
	for i := 0; i < numTraces; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			
			// Create trace
			tid, _ := client.CreateTraceAsync(langfuse.Trace{
				Name:   fmt.Sprintf("concurrent-trace-%d", i),
				UserID: fmt.Sprintf("user-%d", i%10),
				Input:  map[string]interface{}{"index": i},
			})
			
			// Create span
			sid, _ := client.CreateSpanAsync(langfuse.Span{
				TraceID: tid,
				Name:    "concurrent-span",
				Input:   map[string]interface{}{"index": i},
			})
			
			// Update span
			_ = client.UpdateSpanAsync(sid, langfuse.SpanUpdate{
				Output: map[string]interface{}{"completed": true},
			})
		}(i)
	}
	wg.Wait()
	
	elapsed := time.Since(start)
	fmt.Printf("Queued %d traces with spans in %v\n", numTraces, elapsed)
	fmt.Printf("Queue length after concurrent writes: %d\n", client.QueueLength())

	// Example 3: Manual flush
	fmt.Println("\n=== Example 3: Manual Flush ===")
	
	// Sometimes you want to ensure events are sent immediately
	// (e.g., before sending a response to the user)
	fmt.Println("Flushing queue...")
	if err := client.Flush(); err != nil {
		log.Printf("Flush error: %v", err)
	}
	fmt.Printf("Queue length after flush: %d\n", client.QueueLength())

	// Example 4: Using batch processor directly (advanced)
	fmt.Println("\n=== Example 4: Direct BatchProcessor Access ===")
	
	bp := client.BatchProcessor()
	
	// You can enqueue raw batch events
	err = bp.Enqueue(langfuse.BatchEvent{
		Type: langfuse.BatchEventTypeTrace,
		Body: langfuse.Trace{
			Name:  "direct-batch-trace",
			Input: map[string]interface{}{"method": "direct"},
		},
	})
	if err != nil {
		log.Printf("Direct enqueue error: %v", err)
	}
	fmt.Println("Queued event directly via BatchProcessor")

	// Example 5: Sync fallback (when you need the response)
	fmt.Println("\n=== Example 5: Sync Fallback ===")
	
	// AsyncClient embeds Client, so you can still use sync methods
	// when you need the actual response
	syncTrace, err := client.CreateTrace(context.TODO(), langfuse.Trace{
		Name:  "sync-trace",
		Input: map[string]interface{}{"mode": "synchronous"},
	})
	if err != nil {
		log.Printf("Sync create error: %v", err)
	} else {
		fmt.Printf("Created trace synchronously: %s\n", syncTrace.ID)
	}

	fmt.Println("\nAsync batch example completed!")
}

