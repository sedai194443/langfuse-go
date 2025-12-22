package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestClient_CompleteWorkflow tests a complete workflow of creating a trace,
// adding spans and generations, and scoring it
func TestClient_CompleteWorkflow(t *testing.T) {
	var traceID string
	var spanID string
	var genID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/public/traces":
			var trace Trace
			json.NewDecoder(r.Body).Decode(&trace)
			traceID = "trace-123"
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(TraceResponse{ID: traceID})

		case "/api/public/spans":
			var span Span
			json.NewDecoder(r.Body).Decode(&span)
			if span.TraceID != traceID {
				t.Errorf("span.TraceID = %v, want %v", span.TraceID, traceID)
			}
			spanID = "span-123"
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(SpanResponse{ID: spanID})

		case "/api/public/generations":
			var gen Generation
			json.NewDecoder(r.Body).Decode(&gen)
			if gen.TraceID != traceID {
				t.Errorf("gen.TraceID = %v, want %v", gen.TraceID, traceID)
			}
			genID = "gen-123"
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(GenerationResponse{ID: genID})

		case "/api/public/scores":
			var score Score
			json.NewDecoder(r.Body).Decode(&score)
			if score.TraceID != traceID {
				t.Errorf("score.TraceID = %v, want %v", score.TraceID, traceID)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ScoreResponse{ID: "score-123"})

		case "/api/public/spans/span-123":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SpanResponse{ID: spanID})

		case "/api/public/generations/gen-123":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(GenerationResponse{ID: genID})

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{
		PublicKey: "pk-test",
		SecretKey: "sk-test",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create trace
	now := time.Now()
	trace, err := client.CreateTrace(ctx, Trace{
		Name:      "workflow-test",
		UserID:    "user-123",
		Timestamp: &now,
	})
	if err != nil {
		t.Fatalf("CreateTrace() error = %v", err)
	}

	// Create span
	spanStart := time.Now()
	span, err := client.CreateSpan(ctx, Span{
		TraceID:   trace.ID,
		Name:      "database-query",
		StartTime: &spanStart,
		Input: map[string]interface{}{
			"query": "SELECT * FROM users",
		},
	})
	if err != nil {
		t.Fatalf("CreateSpan() error = %v", err)
	}

	// Update span
	spanEnd := time.Now()
	_, err = client.UpdateSpan(ctx, span.ID, SpanUpdate{
		EndTime: &spanEnd,
		Output: map[string]interface{}{
			"rows": 42,
		},
	})
	if err != nil {
		t.Fatalf("UpdateSpan() error = %v", err)
	}

	// Create generation
	genStart := time.Now()
	generation, err := client.CreateGeneration(ctx, Generation{
		TraceID:   trace.ID,
		Name:      "llm-call",
		Model:     "gpt-4",
		StartTime: &genStart,
		Input: map[string]interface{}{
			"prompt": "Hello",
		},
		Usage: &Usage{
			Input:  10,
			Output: 20,
			Total:  30,
		},
	})
	if err != nil {
		t.Fatalf("CreateGeneration() error = %v", err)
	}

	// Update generation
	genEnd := time.Now()
	_, err = client.UpdateGeneration(ctx, generation.ID, GenerationUpdate{
		EndTime: &genEnd,
		Output: map[string]interface{}{
			"response": "Hi there!",
		},
	})
	if err != nil {
		t.Fatalf("UpdateGeneration() error = %v", err)
	}

	// Create score
	_, err = client.Score(ctx, Score{
		TraceID: trace.ID,
		Name:    "quality",
		Value:   0.95,
	})
	if err != nil {
		t.Fatalf("Score() error = %v", err)
	}
}
