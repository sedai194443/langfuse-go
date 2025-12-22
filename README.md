# Langfuse Go SDK

Unofficial Go SDK for [Langfuse](https://langfuse.com) - an open-source LLM engineering platform.

## Installation

```bash
go get github.com/AEKurt/langfuse-go
```

## Quick Start

```go
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
        PublicKey: "your-public-key",
        SecretKey: "your-secret-key",
        // Optional: BaseURL defaults to https://cloud.langfuse.com
        // BaseURL: "https://cloud.langfuse.com",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a trace
    ctx := context.Background()
    now := time.Now()
    trace, err := client.CreateTrace(ctx, langfuse.Trace{
        Name:      "my-trace",
        UserID:    "user-123",
        Metadata:  map[string]interface{}{
            "environment": "production",
        },
        Tags:      []string{"important", "test"},
        Timestamp: &now,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created trace: %s\n", trace.ID)

    // Create a generation (LLM call)
    startTime := time.Now()
    generation, err := client.CreateGeneration(ctx, langfuse.Generation{
        TraceID: trace.ID,
        Name:    "chat-completion",
        Model:   "gpt-4",
        StartTime: &startTime,
        Input: map[string]interface{}{
            "messages": []map[string]interface{}{
                {"role": "user", "content": "Hello!"},
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

    // Update the generation with output
    endTime := time.Now()
    _, err = client.UpdateGeneration(ctx, generation.ID, langfuse.GenerationUpdate{
        EndTime: &endTime,
        Output: map[string]interface{}{
            "messages": []map[string]interface{}{
                {"role": "assistant", "content": "Hi there!"},
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create a span
    spanStart := time.Now()
    span, err := client.CreateSpan(ctx, langfuse.Span{
        TraceID: trace.ID,
        Name:    "database-query",
        StartTime: &spanStart,
        Input: map[string]interface{}{
            "query": "SELECT * FROM users",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Update span with output
    spanEnd := time.Now()
    _, err = client.UpdateSpan(ctx, span.ID, langfuse.SpanUpdate{
        EndTime: &spanEnd,
        Output: map[string]interface{}{
            "rows": 42,
        },
    })
    if err != nil {
        log.Fatal(err)
    }

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
}
```

## API Reference

### Client

#### `NewClient(config Config) (*Client, error)`

Creates a new Langfuse client.

**Parameters:**
- `config.PublicKey` (string, required): Your Langfuse public key
- `config.SecretKey` (string, required): Your Langfuse secret key
- `config.BaseURL` (string, optional): Base URL for Langfuse API (defaults to `https://cloud.langfuse.com`)
- `config.HTTPClient` (*http.Client, optional): Custom HTTP client

### Traces

#### `CreateTrace(ctx context.Context, trace Trace) (*TraceResponse, error)`

Creates a new trace.

#### `UpdateTrace(ctx context.Context, traceID string, trace TraceUpdate) (*TraceResponse, error)`

Updates an existing trace.

### Spans

#### `CreateSpan(ctx context.Context, span Span) (*SpanResponse, error)`

Creates a new span.

#### `UpdateSpan(ctx context.Context, spanID string, span SpanUpdate) (*SpanResponse, error)`

Updates an existing span.

### Generations

#### `CreateGeneration(ctx context.Context, generation Generation) (*GenerationResponse, error)`

Creates a new generation (LLM call).

#### `UpdateGeneration(ctx context.Context, generationID string, generation GenerationUpdate) (*GenerationResponse, error)`

Updates an existing generation.

### Events

#### `CreateEvent(ctx context.Context, event Event) (*EventResponse, error)`

Creates a new event.

### Scores

#### `Score(ctx context.Context, score Score) (*ScoreResponse, error)`

Creates a score for a trace or observation.

## Examples

### Getting Started

See the `examples/getting_started/` directory for a simple step-by-step example.

### Basic Examples

See the `examples/basic/` directory for basic usage examples.

### Advanced Examples

See the `examples/advanced/` directory for advanced features including:
- W3C Trace Context support
- Context manager pattern (Go equivalent of Python's `with` blocks)
- Trace context propagation
- Nested observations
- Different observation types

### Observe Wrapper Examples

See the `examples/observe_wrapper/` directory for examples of wrapping functions with automatic observation tracking.

### Error Handling Examples

See the `examples/error_handling/` directory for examples of handling errors and edge cases.

### Logger Examples

See the `examples/logger/` directory for examples of using the logger interface for debugging.

### Context Propagation Examples

See the `examples/context_propagation/` directory for examples of propagating trace context across services.

## Advanced Features

### W3C Trace Context

Generate W3C-compliant trace and observation IDs:

```go
// Generate deterministic trace ID from external ID
externalID := "req_12345"
traceID := langfuse.CreateTraceID(externalID)

// Generate observation ID
obsID := langfuse.CreateObservationID()

// Get current trace/observation from context
if traceID, ok := langfuse.GetCurrentTraceID(ctx); ok {
    fmt.Printf("Current trace: %s\n", traceID)
}
```

### Context Manager Pattern

Use `StartObservation` for automatic parent-child relationships:

```go
// Start a root span
rootSpan, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "my-operation", input)
if err != nil {
    return err
}
defer rootSpan.End()

// Create child observations
child, err := rootSpan.StartChildObservation(langfuse.ObservationTypeGeneration, "llm-call", prompt)
if err != nil {
    return err
}
defer child.End()

// Update observations
rootSpan.Update(langfuse.SpanUpdate{
    Output: result,
})
```

### Trace Context Propagation

Propagate trace context across services:

```go
// Create trace context
traceCtx := langfuse.TraceContext{
    TraceID: "existing-trace-id",
    SpanID:  "parent-span-id",
}
ctx := langfuse.WithTraceContext(ctx, traceCtx)

// New observations will join the existing trace
span, err := client.StartObservation(ctx, langfuse.ObservationTypeSpan, "downstream-task", input)
```

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

