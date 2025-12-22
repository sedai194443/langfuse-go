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

	ctx := context.Background()

	// Create a trace
	trace, err := client.CreateTrace(ctx, langfuse.Trace{
		Name:      "agent-workflow",
		UserID:    "user-123",
		SessionID: "session-abc",
		Input:     map[string]interface{}{"query": "Find me a good Italian restaurant nearby"},
		Metadata: map[string]interface{}{
			"account_id": "acc-456",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created trace: %s\n", trace.ID)

	// Set up trace context for child observations
	traceCtx := langfuse.WithTraceContext(ctx, langfuse.TraceContext{
		TraceID: trace.ID,
	})

	// Example 1: Agent - Reasoning blocks that act on tools using LLM guidance
	fmt.Println("\n=== Example 1: Agent Observation ===")
	agent, err := client.StartAgent(traceCtx, "restaurant-finder-agent", map[string]interface{}{
		"task":        "Find Italian restaurant",
		"constraints": []string{"nearby", "good reviews"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started agent: %s\n", agent.ID)

	// Agent uses its context for child observations
	agentCtx := agent.Context()

	// Example 2: Tool - External tool calls (e.g., APIs)
	fmt.Println("\n=== Example 2: Tool Observation ===")
	tool, err := client.StartTool(agentCtx, "location-api", map[string]interface{}{
		"action":   "get_current_location",
		"provider": "gps",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started tool: %s\n", tool.ID)

	// Simulate API call
	time.Sleep(20 * time.Millisecond)
	_ = tool.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
			"city":      "New York",
		},
	})
	_ = tool.End()
	fmt.Println("Tool completed")

	// Example 3: Retriever - Data retrieval (e.g., vector stores, databases)
	fmt.Println("\n=== Example 3: Retriever Observation ===")
	retriever, err := client.StartRetriever(agentCtx, "restaurant-vector-search", map[string]interface{}{
		"query":    "Italian restaurant New York",
		"top_k":    5,
		"provider": "pinecone",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started retriever: %s\n", retriever.ID)

	// Simulate vector search
	time.Sleep(30 * time.Millisecond)
	_ = retriever.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"results": []map[string]interface{}{
				{"name": "Carbone", "rating": 4.8, "distance": "0.5 miles"},
				{"name": "L'Artusi", "rating": 4.6, "distance": "0.8 miles"},
				{"name": "Via Carota", "rating": 4.7, "distance": "1.0 miles"},
			},
			"total_results": 3,
		},
	})
	_ = retriever.End()
	fmt.Println("Retriever completed")

	// Example 4: Chain - Connecting LLM application steps
	fmt.Println("\n=== Example 4: Chain Observation ===")
	chain, err := client.StartChain(agentCtx, "rag-chain", map[string]interface{}{
		"step":    "combine-context-and-query",
		"context": "Restaurant search results",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started chain: %s\n", chain.ID)

	// Example 5: Generation - LLM call within chain
	fmt.Println("\n=== Example 5: Generation Observation ===")
	chainCtx := chain.Context()
	gen, err := client.StartObservation(chainCtx, langfuse.ObservationTypeGeneration, "recommend-restaurant", map[string]interface{}{
		"messages": []map[string]interface{}{
			{"role": "system", "content": "You are a restaurant recommendation assistant."},
			{"role": "user", "content": "Based on these options, recommend the best Italian restaurant: Carbone (4.8), L'Artusi (4.6), Via Carota (4.7)"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started generation: %s\n", gen.ID)

	// Simulate LLM call
	time.Sleep(50 * time.Millisecond)
	model := "gpt-4"
	_ = gen.Update(langfuse.GenerationUpdate{
		Model: &model,
		Output: map[string]interface{}{
			"response": "I recommend Carbone! It has the highest rating (4.8) and is the closest at just 0.5 miles away.",
		},
		Usage: &langfuse.Usage{
			Input:  50,
			Output: 30,
			Total:  80,
			Unit:   "TOKENS",
		},
	})
	_ = gen.End()
	fmt.Println("Generation completed")

	// Complete chain
	_ = chain.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"recommendation": "Carbone",
		},
	})
	_ = chain.End()
	fmt.Println("Chain completed")

	// Example 6: Evaluator - Assessing LLM outputs
	fmt.Println("\n=== Example 6: Evaluator Observation ===")
	evaluator, err := client.StartEvaluator(agentCtx, "response-quality-check", map[string]interface{}{
		"response": "I recommend Carbone!",
		"criteria": []string{"relevance", "helpfulness", "accuracy"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started evaluator: %s\n", evaluator.ID)

	// Simulate evaluation
	time.Sleep(20 * time.Millisecond)
	_ = evaluator.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"scores": map[string]float64{
				"relevance":   0.95,
				"helpfulness": 0.90,
				"accuracy":    0.92,
			},
			"overall_score": 0.92,
			"passed":        true,
		},
	})
	_ = evaluator.End()
	fmt.Println("Evaluator completed")

	// Example 7: Guardrail - Protection against jailbreaks, offensive content
	fmt.Println("\n=== Example 7: Guardrail Observation ===")
	guardrail, err := client.StartGuardrail(agentCtx, "content-safety-filter", map[string]interface{}{
		"content": "I recommend Carbone!",
		"checks":  []string{"pii", "toxicity", "jailbreak"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started guardrail: %s\n", guardrail.ID)

	// Simulate safety check
	time.Sleep(10 * time.Millisecond)
	_ = guardrail.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"checks": map[string]interface{}{
				"pii":       map[string]interface{}{"detected": false},
				"toxicity":  map[string]interface{}{"score": 0.01, "passed": true},
				"jailbreak": map[string]interface{}{"detected": false},
			},
			"overall_safe": true,
		},
	})
	_ = guardrail.End()
	fmt.Println("Guardrail completed")

	// Example 8: Embedding - LLM embedding calls
	fmt.Println("\n=== Example 8: Embedding Observation ===")
	embedding, err := client.StartEmbedding(agentCtx, "query-embedding", "text-embedding-3-small", map[string]interface{}{
		"text": "Find Italian restaurant",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Started embedding: %s\n", embedding.ID)

	// Simulate embedding call
	time.Sleep(15 * time.Millisecond)
	_ = embedding.Update(langfuse.GenerationUpdate{
		Output: map[string]interface{}{
			"dimensions": 1536,
			"model":      "text-embedding-3-small",
		},
		Usage: &langfuse.Usage{
			Input: 4,
			Total: 4,
			Unit:  "TOKENS",
		},
	})
	_ = embedding.End()
	fmt.Println("Embedding completed")

	// Complete the agent
	_ = agent.Update(langfuse.SpanUpdate{
		Output: map[string]interface{}{
			"recommendation": "Carbone",
			"reason":         "Highest rated and closest",
		},
	})
	_ = agent.End()
	fmt.Println("\nAgent completed")

	// Example 9: Using StartAsCurrent* methods for context management
	fmt.Println("\n=== Example 9: StartAsCurrent* Methods ===")

	// Start agent as current - returns context with observation stored
	agentCtx2, agent2, err := client.StartAsCurrentAgent(ctx, "search-agent", map[string]interface{}{
		"query": "best pizza",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = agent2.End() }()

	// Start tool as current within agent context
	toolCtx, tool2, err := client.StartAsCurrentTool(agentCtx2, "search-api", map[string]interface{}{
		"query": "pizza restaurants",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = tool2.End() }()

	// Get current observation from context
	if obs, ok := langfuse.GetCurrentObservation(toolCtx); ok {
		fmt.Printf("Current observation from context: %s (type: %s)\n", obs.ID, obs.Type)
	}

	// Update current span via context
	err = client.UpdateCurrentSpan(toolCtx, map[string]interface{}{
		"results": []string{"Pizza Place 1", "Pizza Place 2"},
	}, nil)
	if err != nil {
		log.Printf("UpdateCurrentSpan error: %v", err)
	}

	// Create score for the trace
	_, err = client.Score(ctx, langfuse.Score{
		TraceID: trace.ID,
		Name:    "task_completion",
		Value:   1.0,
		Comment: "Successfully found restaurant recommendation",
	})
	if err != nil {
		log.Printf("Score error: %v", err)
	}

	fmt.Println("\nSpecialized observation types example completed!")
}
