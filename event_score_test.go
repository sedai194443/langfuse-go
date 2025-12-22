package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/events" {
			t.Errorf("expected /api/public/events, got %s", r.URL.Path)
		}

		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if event.Name != "user-action" {
			t.Errorf("expected name 'user-action', got %s", event.Name)
		}
		if event.TraceID != "trace-123" {
			t.Errorf("expected traceID 'trace-123', got %s", event.TraceID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(EventResponse{ID: "event-123"})
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

	startTime := time.Now()
	event, err := client.CreateEvent(context.Background(), Event{
		TraceID:   "trace-123",
		Name:      "user-action",
		StartTime: &startTime,
		Input: map[string]interface{}{
			"action": "click",
			"button": "submit",
		},
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if event.ID != "event-123" {
		t.Errorf("CreateEvent() ID = %v, want event-123", event.ID)
	}
}

func TestClient_Score(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/scores" {
			t.Errorf("expected /api/public/scores, got %s", r.URL.Path)
		}

		var score Score
		if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if score.Name != "quality" {
			t.Errorf("expected name 'quality', got %s", score.Name)
		}
		if score.Value != 0.95 {
			t.Errorf("expected value 0.95, got %f", score.Value)
		}
		if score.TraceID != "trace-123" {
			t.Errorf("expected traceID 'trace-123', got %s", score.TraceID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(ScoreResponse{ID: "score-123"})
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

	score, err := client.Score(context.Background(), Score{
		TraceID: "trace-123",
		Name:    "quality",
		Value:   0.95,
		Comment: "High quality response",
	})
	if err != nil {
		t.Fatalf("Score() error = %v", err)
	}
	if score.ID != "score-123" {
		t.Errorf("Score() ID = %v, want score-123", score.ID)
	}
}

func TestClient_Score_WithObservationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var score Score
		if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if score.ObservationID != "obs-123" {
			t.Errorf("expected observationID 'obs-123', got %s", score.ObservationID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(ScoreResponse{ID: "score-123"})
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

	_, err = client.Score(context.Background(), Score{
		TraceID:       "trace-123",
		ObservationID: "obs-123",
		Name:          "quality",
		Value:         0.95,
	})
	if err != nil {
		t.Fatalf("Score() error = %v", err)
	}
}
