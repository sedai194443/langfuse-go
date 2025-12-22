package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestObservationTypes(t *testing.T) {
	// Verify all observation types are defined correctly
	types := []ObservationType{
		ObservationTypeSpan,
		ObservationTypeGeneration,
		ObservationTypeEvent,
		ObservationTypeAgent,
		ObservationTypeTool,
		ObservationTypeChain,
		ObservationTypeRetriever,
		ObservationTypeEmbedding,
		ObservationTypeEvaluator,
		ObservationTypeGuardrail,
	}

	expectedValues := map[ObservationType]string{
		ObservationTypeSpan:       "span",
		ObservationTypeGeneration: "generation",
		ObservationTypeEvent:      "event",
		ObservationTypeAgent:      "agent",
		ObservationTypeTool:       "tool",
		ObservationTypeChain:      "chain",
		ObservationTypeRetriever:  "retriever",
		ObservationTypeEmbedding:  "embedding",
		ObservationTypeEvaluator:  "evaluator",
		ObservationTypeGuardrail:  "guardrail",
	}

	for _, obsType := range types {
		expected, ok := expectedValues[obsType]
		if !ok {
			t.Errorf("unexpected observation type: %s", obsType)
			continue
		}
		if string(obsType) != expected {
			t.Errorf("observation type %s has wrong value: got %s, want %s", obsType, string(obsType), expected)
		}
	}
}

func createTestServer(t *testing.T) (*httptest.Server, *Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Decode request body to get ID
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		id := "generated-id"
		if bodyID, ok := body["id"].(string); ok && bodyID != "" {
			id = bodyID
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	}))

	client, err := NewClient(Config{
		PublicKey: "pk-test",
		SecretKey: "sk-test",
		BaseURL:   server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return server, client
}

func TestStartObservation_AllTypes(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	tests := []struct {
		name    string
		obsType ObservationType
	}{
		{"span", ObservationTypeSpan},
		{"generation", ObservationTypeGeneration},
		{"event", ObservationTypeEvent},
		{"agent", ObservationTypeAgent},
		{"tool", ObservationTypeTool},
		{"chain", ObservationTypeChain},
		{"retriever", ObservationTypeRetriever},
		{"embedding", ObservationTypeEmbedding},
		{"evaluator", ObservationTypeEvaluator},
		{"guardrail", ObservationTypeGuardrail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs, err := client.StartObservation(context.Background(), tt.obsType, "test-"+tt.name, map[string]string{"key": "value"})
			if err != nil {
				t.Fatalf("StartObservation(%s) error = %v", tt.obsType, err)
			}
			if obs == nil {
				t.Fatalf("StartObservation(%s) returned nil", tt.obsType)
			}
			if obs.Type != tt.obsType {
				t.Errorf("StartObservation(%s) Type = %s, want %s", tt.obsType, obs.Type, tt.obsType)
			}
			if obs.ID == "" {
				t.Errorf("StartObservation(%s) ID is empty", tt.obsType)
			}
			if obs.TraceID == "" {
				t.Errorf("StartObservation(%s) TraceID is empty", tt.obsType)
			}
		})
	}
}

func TestStartAgent(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartAgent(context.Background(), "test-agent", map[string]string{"task": "reasoning"})
	if err != nil {
		t.Fatalf("StartAgent() error = %v", err)
	}
	if obs.Type != ObservationTypeAgent {
		t.Errorf("StartAgent() Type = %s, want %s", obs.Type, ObservationTypeAgent)
	}
}

func TestStartTool(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartTool(context.Background(), "weather-api", map[string]string{"city": "NYC"})
	if err != nil {
		t.Fatalf("StartTool() error = %v", err)
	}
	if obs.Type != ObservationTypeTool {
		t.Errorf("StartTool() Type = %s, want %s", obs.Type, ObservationTypeTool)
	}
}

func TestStartChain(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartChain(context.Background(), "rag-chain", map[string]string{"query": "test"})
	if err != nil {
		t.Fatalf("StartChain() error = %v", err)
	}
	if obs.Type != ObservationTypeChain {
		t.Errorf("StartChain() Type = %s, want %s", obs.Type, ObservationTypeChain)
	}
}

func TestStartRetriever(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartRetriever(context.Background(), "vector-search", map[string]string{"query": "semantic search"})
	if err != nil {
		t.Fatalf("StartRetriever() error = %v", err)
	}
	if obs.Type != ObservationTypeRetriever {
		t.Errorf("StartRetriever() Type = %s, want %s", obs.Type, ObservationTypeRetriever)
	}
}

func TestStartEmbedding(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartEmbedding(context.Background(), "embed-text", "text-embedding-3-small", "Hello world")
	if err != nil {
		t.Fatalf("StartEmbedding() error = %v", err)
	}
	if obs.Type != ObservationTypeEmbedding {
		t.Errorf("StartEmbedding() Type = %s, want %s", obs.Type, ObservationTypeEmbedding)
	}
}

func TestStartEvaluator(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartEvaluator(context.Background(), "quality-check", map[string]string{"response": "test response"})
	if err != nil {
		t.Fatalf("StartEvaluator() error = %v", err)
	}
	if obs.Type != ObservationTypeEvaluator {
		t.Errorf("StartEvaluator() Type = %s, want %s", obs.Type, ObservationTypeEvaluator)
	}
}

func TestStartGuardrail(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	obs, err := client.StartGuardrail(context.Background(), "content-filter", map[string]string{"input": "user message"})
	if err != nil {
		t.Fatalf("StartGuardrail() error = %v", err)
	}
	if obs.Type != ObservationTypeGuardrail {
		t.Errorf("StartGuardrail() Type = %s, want %s", obs.Type, ObservationTypeGuardrail)
	}
}

func TestStartAsCurrentAgent(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentAgent(context.Background(), "test-agent", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentAgent() error = %v", err)
	}
	if obs.Type != ObservationTypeAgent {
		t.Errorf("StartAsCurrentAgent() Type = %s, want %s", obs.Type, ObservationTypeAgent)
	}

	// Verify observation is stored in context
	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentAgent() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentAgent() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentTool(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentTool(context.Background(), "test-tool", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentTool() error = %v", err)
	}
	if obs.Type != ObservationTypeTool {
		t.Errorf("StartAsCurrentTool() Type = %s, want %s", obs.Type, ObservationTypeTool)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentTool() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentTool() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentChain(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentChain(context.Background(), "test-chain", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentChain() error = %v", err)
	}
	if obs.Type != ObservationTypeChain {
		t.Errorf("StartAsCurrentChain() Type = %s, want %s", obs.Type, ObservationTypeChain)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentChain() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentChain() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentRetriever(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentRetriever(context.Background(), "test-retriever", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentRetriever() error = %v", err)
	}
	if obs.Type != ObservationTypeRetriever {
		t.Errorf("StartAsCurrentRetriever() Type = %s, want %s", obs.Type, ObservationTypeRetriever)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentRetriever() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentRetriever() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentEmbedding(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentEmbedding(context.Background(), "test-embedding", "text-embedding-3-small", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentEmbedding() error = %v", err)
	}
	if obs.Type != ObservationTypeEmbedding {
		t.Errorf("StartAsCurrentEmbedding() Type = %s, want %s", obs.Type, ObservationTypeEmbedding)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentEmbedding() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentEmbedding() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentEvaluator(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentEvaluator(context.Background(), "test-evaluator", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentEvaluator() error = %v", err)
	}
	if obs.Type != ObservationTypeEvaluator {
		t.Errorf("StartAsCurrentEvaluator() Type = %s, want %s", obs.Type, ObservationTypeEvaluator)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentEvaluator() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentEvaluator() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestStartAsCurrentGuardrail(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	ctx, obs, err := client.StartAsCurrentGuardrail(context.Background(), "test-guardrail", nil)
	if err != nil {
		t.Fatalf("StartAsCurrentGuardrail() error = %v", err)
	}
	if obs.Type != ObservationTypeGuardrail {
		t.Errorf("StartAsCurrentGuardrail() Type = %s, want %s", obs.Type, ObservationTypeGuardrail)
	}

	retrieved, ok := GetCurrentObservation(ctx)
	if !ok {
		t.Error("StartAsCurrentGuardrail() observation not stored in context")
	}
	if retrieved.ID != obs.ID {
		t.Errorf("StartAsCurrentGuardrail() retrieved ID = %s, want %s", retrieved.ID, obs.ID)
	}
}

func TestObservation_Update_SpanLikeTypes(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	spanLikeTypes := []ObservationType{
		ObservationTypeSpan,
		ObservationTypeAgent,
		ObservationTypeTool,
		ObservationTypeChain,
		ObservationTypeRetriever,
		ObservationTypeEvaluator,
		ObservationTypeGuardrail,
	}

	for _, obsType := range spanLikeTypes {
		t.Run(string(obsType), func(t *testing.T) {
			obs, err := client.StartObservation(context.Background(), obsType, "test", nil)
			if err != nil {
				t.Fatalf("StartObservation() error = %v", err)
			}

			err = obs.Update(SpanUpdate{
				Output: map[string]string{"result": "success"},
			})
			if err != nil {
				t.Errorf("Update() error = %v", err)
			}
		})
	}
}

func TestObservation_Update_GenerationLikeTypes(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	genLikeTypes := []ObservationType{
		ObservationTypeGeneration,
		ObservationTypeEmbedding,
	}

	for _, obsType := range genLikeTypes {
		t.Run(string(obsType), func(t *testing.T) {
			obs, err := client.StartObservation(context.Background(), obsType, "test", nil)
			if err != nil {
				t.Fatalf("StartObservation() error = %v", err)
			}

			model := "gpt-4"
			err = obs.Update(GenerationUpdate{
				Model:  &model,
				Output: map[string]string{"result": "success"},
			})
			if err != nil {
				t.Errorf("Update() error = %v", err)
			}
		})
	}
}

func TestObservation_End_AllTypes(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	allTypes := []ObservationType{
		ObservationTypeSpan,
		ObservationTypeGeneration,
		ObservationTypeEvent,
		ObservationTypeAgent,
		ObservationTypeTool,
		ObservationTypeChain,
		ObservationTypeRetriever,
		ObservationTypeEmbedding,
		ObservationTypeEvaluator,
		ObservationTypeGuardrail,
	}

	for _, obsType := range allTypes {
		t.Run(string(obsType), func(t *testing.T) {
			obs, err := client.StartObservation(context.Background(), obsType, "test", nil)
			if err != nil {
				t.Fatalf("StartObservation() error = %v", err)
			}

			err = obs.End()
			if err != nil {
				t.Errorf("End() error = %v", err)
			}
		})
	}
}

func TestObservation_StartChildObservation(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	parent, err := client.StartAgent(context.Background(), "parent-agent", nil)
	if err != nil {
		t.Fatalf("StartAgent() error = %v", err)
	}

	child, err := parent.StartChildObservation(ObservationTypeTool, "child-tool", nil)
	if err != nil {
		t.Fatalf("StartChildObservation() error = %v", err)
	}

	if child.Type != ObservationTypeTool {
		t.Errorf("StartChildObservation() Type = %s, want %s", child.Type, ObservationTypeTool)
	}
	if child.TraceID != parent.TraceID {
		t.Errorf("StartChildObservation() TraceID = %s, want %s", child.TraceID, parent.TraceID)
	}
	if child.ParentSpanID != parent.ID {
		t.Errorf("StartChildObservation() ParentSpanID = %s, want %s", child.ParentSpanID, parent.ID)
	}
}

func TestObservation_InvalidUpdateType(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	// Span-like type with GenerationUpdate should fail
	spanObs, _ := client.StartAgent(context.Background(), "test-agent", nil)
	err := spanObs.Update(GenerationUpdate{})
	if err == nil {
		t.Error("Update() with wrong type should return error")
	}

	// Generation-like type with SpanUpdate should fail
	genObs, _ := client.StartObservation(context.Background(), ObservationTypeGeneration, "test-gen", nil)
	err = genObs.Update(SpanUpdate{})
	if err == nil {
		t.Error("Update() with wrong type should return error")
	}
}

func TestStartObservation_UnknownType(t *testing.T) {
	server, client := createTestServer(t)
	defer server.Close()

	_, err := client.StartObservation(context.Background(), ObservationType("unknown"), "test", nil)
	if err == nil {
		t.Error("StartObservation() with unknown type should return error")
	}
}

