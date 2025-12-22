package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AEKurt/langfuse-go"
)

// CustomLogger implements the langfuse.Logger interface for debugging
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) LogRequest(method, url string, body interface{}) {
	log.Printf("[%s] REQUEST: %s %s", l.prefix, method, url)
	if body != nil {
		log.Printf("[%s] REQUEST BODY: %+v", l.prefix, body)
	}
}

func (l *CustomLogger) LogResponse(statusCode int, body []byte, err error) {
	if err != nil {
		log.Printf("[%s] RESPONSE ERROR: %v", l.prefix, err)
		return
	}
	log.Printf("[%s] RESPONSE: Status %d, Body: %s", l.prefix, statusCode, string(body))
}

func main() {
	// Initialize client with custom logger
	client, err := langfuse.NewClient(langfuse.Config{
		PublicKey: "pk-lf-your-public-key",
		SecretKey: "sk-lf-your-secret-key",
		Logger: &CustomLogger{
			prefix: "LANGFUSE",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("=== Logger Example ===")
	fmt.Println("All API requests and responses will be logged")
	fmt.Println()

	// Create a trace - logger will capture the request
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name:   "logger-demo",
		UserID: "user-123",
		Metadata: map[string]interface{}{
			"environment": "development",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Trace created: %s\n\n", trace.ID)

	// Create a generation - logger will capture request and response
	startTime := time.Now()
	generation, err := client.CreateGeneration(ctx, langfuse.Generation{
		TraceID: trace.ID,
		Name:    "test-generation",
		Model:   "gpt-4",
		Input: map[string]interface{}{
			"prompt": "Hello",
		},
		StartTime: &startTime,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generation created: %s\n\n", generation.ID)

	// Update generation - logger will capture the update request
	endTime := time.Now()
	_, err = client.UpdateGeneration(ctx, generation.ID, langfuse.GenerationUpdate{
		EndTime: &endTime,
		Output: map[string]interface{}{
			"response": "Hi there!",
		},
		Usage: &langfuse.Usage{
			Input:  5,
			Output: 10,
			Total:  15,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Generation updated")
	fmt.Println()

	fmt.Println("Check the logs above to see all API requests and responses")
	fmt.Println("This is useful for debugging and understanding SDK behavior")
}
