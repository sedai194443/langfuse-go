package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_CreateGeneration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/generations" {
			t.Errorf("expected /api/public/generations, got %s", r.URL.Path)
		}

		var gen Generation
		if err := json.NewDecoder(r.Body).Decode(&gen); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if gen.Name != "chat-completion" {
			t.Errorf("expected name 'chat-completion', got %s", gen.Name)
		}
		if gen.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got %s", gen.Model)
		}
		if gen.Usage == nil {
			t.Error("expected usage to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(GenerationResponse{ID: "gen-123"})
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
	generation, err := client.CreateGeneration(context.Background(), Generation{
		TraceID:   "trace-123",
		Name:      "chat-completion",
		Model:     "gpt-4",
		StartTime: &startTime,
		Input: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "user", "content": "Hello!"},
			},
		},
		Usage: &Usage{
			Input:  10,
			Output: 20,
			Total:  30,
			Unit:   "TOKENS",
		},
	})
	if err != nil {
		t.Fatalf("CreateGeneration() error = %v", err)
	}
	if generation.ID != "gen-123" {
		t.Errorf("CreateGeneration() ID = %v, want gen-123", generation.ID)
	}
}

func TestClient_UpdateGeneration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/generations/gen-123" {
			t.Errorf("expected /api/public/generations/gen-123, got %s", r.URL.Path)
		}

		var update GenerationUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if update.Output == nil {
			t.Error("expected output to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GenerationResponse{ID: "gen-123"})
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
	generation, err := client.UpdateGeneration(context.Background(), "gen-123", GenerationUpdate{
		EndTime: &endTime,
		Output: map[string]interface{}{
			"messages": []map[string]interface{}{
				{"role": "assistant", "content": "Hi there!"},
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateGeneration() error = %v", err)
	}
	if generation.ID != "gen-123" {
		t.Errorf("UpdateGeneration() ID = %v, want gen-123", generation.ID)
	}
}

func TestClient_CreateGeneration_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
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

	_, err = client.CreateGeneration(context.Background(), Generation{Name: "test"})
	if err == nil {
		t.Error("CreateGeneration() expected error, got nil")
	}
	if !IsAPIError(err) {
		t.Error("CreateGeneration() expected APIError")
	}
	apiErr := err.(*APIError)
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("CreateGeneration() status code = %v, want %v", apiErr.StatusCode, http.StatusUnauthorized)
	}
}

