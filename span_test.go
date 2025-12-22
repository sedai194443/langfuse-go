package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateSpan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/spans" {
			t.Errorf("expected /api/public/spans, got %s", r.URL.Path)
		}

		var span Span
		if err := json.NewDecoder(r.Body).Decode(&span); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if span.Name != "test-span" {
			t.Errorf("expected name 'test-span', got %s", span.Name)
		}
		if span.TraceID != "trace-123" {
			t.Errorf("expected traceID 'trace-123', got %s", span.TraceID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SpanResponse{ID: "span-123"})
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
	span, err := client.CreateSpan(context.Background(), Span{
		TraceID:   "trace-123",
		Name:      "test-span",
		StartTime: &startTime,
		Input: map[string]interface{}{
			"query": "SELECT * FROM users",
		},
	})
	if err != nil {
		t.Fatalf("CreateSpan() error = %v", err)
	}
	if span.ID != "span-123" {
		t.Errorf("CreateSpan() ID = %v, want span-123", span.ID)
	}
}

func TestClient_UpdateSpan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// UpdateSpan uses POST with ID in body (upsert behavior)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/spans" {
			t.Errorf("expected /api/public/spans, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		// Check that ID is included in the body
		if body["id"] != "span-123" {
			t.Errorf("expected id 'span-123' in body, got %v", body["id"])
		}

		if body["output"] == nil {
			t.Error("expected output to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(SpanResponse{ID: "span-123"})
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

	endTime := time.Now()
	span, err := client.UpdateSpan(context.Background(), "span-123", SpanUpdate{
		EndTime: &endTime,
		Output: map[string]interface{}{
			"rows": 42,
		},
	})
	if err != nil {
		t.Fatalf("UpdateSpan() error = %v", err)
	}
	if span.ID != "span-123" {
		t.Errorf("UpdateSpan() ID = %v, want span-123", span.ID)
	}
}

func TestClient_CreateSpan_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
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

	_, err = client.CreateSpan(context.Background(), Span{Name: "test"})
	if err == nil {
		t.Error("CreateSpan() expected error, got nil")
	}
	if !IsAPIError(err) {
		t.Error("CreateSpan() expected APIError")
	}
}
