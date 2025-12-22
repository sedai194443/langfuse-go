package langfuse

import (
	"context"
	"fmt"
	"time"
)

// Error message constants
const errUnknownObservationType = "unknown observation type: %s"

// ObservationType represents the type of observation
type ObservationType string

const (
	// ObservationTypeSpan represents a span observation
	ObservationTypeSpan ObservationType = "span"
	// ObservationTypeGeneration represents a generation observation
	ObservationTypeGeneration ObservationType = "generation"
	// ObservationTypeEvent represents an event observation
	ObservationTypeEvent ObservationType = "event"
	// ObservationTypeAgent represents an agent observation (reasoning blocks using LLM guidance)
	ObservationTypeAgent ObservationType = "agent"
	// ObservationTypeTool represents a tool observation (external tool calls, e.g., APIs)
	ObservationTypeTool ObservationType = "tool"
	// ObservationTypeChain represents a chain observation (connecting LLM application steps)
	ObservationTypeChain ObservationType = "chain"
	// ObservationTypeRetriever represents a retriever observation (data retrieval, e.g., vector stores)
	ObservationTypeRetriever ObservationType = "retriever"
	// ObservationTypeEmbedding represents an embedding observation (LLM embedding calls)
	ObservationTypeEmbedding ObservationType = "embedding"
	// ObservationTypeEvaluator represents an evaluator observation (assessing LLM outputs)
	ObservationTypeEvaluator ObservationType = "evaluator"
	// ObservationTypeGuardrail represents a guardrail observation (protection against jailbreaks, etc.)
	ObservationTypeGuardrail ObservationType = "guardrail"
)

// Observation represents an active observation that can be updated
type Observation struct {
	client       *Client
	ctx          context.Context
	Type         ObservationType
	ID           string
	TraceID      string
	ParentSpanID string
}

// StartObservation starts a new observation and returns an Observation handle
// This is the Go equivalent of the context manager pattern
func (c *Client) StartObservation(ctx context.Context, obsType ObservationType, name string, input interface{}) (*Observation, error) {
	// Get trace context from parent context if available
	traceCtx, hasTraceCtx := GetTraceContext(ctx)

	var traceID string
	var parentSpanID string

	if hasTraceCtx {
		traceID = traceCtx.TraceID
		parentSpanID = traceCtx.SpanID
	} else {
		// Create new trace if no context exists
		traceID = CreateTraceID()
	}

	observationID := CreateObservationID()
	startTime := time.Now()

	var id string

	switch obsType {
	case ObservationTypeSpan, ObservationTypeAgent, ObservationTypeTool, ObservationTypeChain,
		ObservationTypeRetriever, ObservationTypeEvaluator, ObservationTypeGuardrail:
		// All span-like types use the Span struct with metadata to indicate subtype
		span := Span{
			ID:                  observationID,
			TraceID:             traceID,
			Name:                name,
			StartTime:           &startTime,
			Input:               input,
			ParentObservationID: parentSpanID,
		}
		// Add observation type metadata for specialized types
		if obsType != ObservationTypeSpan {
			span.Metadata = map[string]interface{}{
				"observation_type": string(obsType),
			}
		}
		resp, err := c.CreateSpan(ctx, span)
		if err != nil {
			return nil, err
		}
		id = resp.ID
	case ObservationTypeGeneration, ObservationTypeEmbedding:
		// Generation-like types (generation and embedding)
		gen := Generation{
			ID:                  observationID,
			TraceID:             traceID,
			Name:                name,
			StartTime:           &startTime,
			Input:               input,
			ParentObservationID: parentSpanID,
		}
		// Add observation type metadata for embedding
		if obsType == ObservationTypeEmbedding {
			gen.Metadata = map[string]interface{}{
				"observation_type": string(obsType),
			}
		}
		resp, err := c.CreateGeneration(ctx, gen)
		if err != nil {
			return nil, err
		}
		id = resp.ID
	case ObservationTypeEvent:
		event := Event{
			ID:                  observationID,
			TraceID:             traceID,
			Name:                name,
			StartTime:           &startTime,
			Input:               input,
			ParentObservationID: parentSpanID,
		}
		resp, err := c.CreateEvent(ctx, event)
		if err != nil {
			return nil, err
		}
		id = resp.ID
	default:
		return nil, fmt.Errorf(errUnknownObservationType, obsType)
	}

	// Create new trace context with this observation as active
	newTraceCtx := TraceContext{
		TraceID:      traceID,
		SpanID:       id,
		ParentSpanID: parentSpanID,
	}
	newCtx := WithTraceContext(ctx, newTraceCtx)

	return &Observation{
		client:       c,
		ctx:          newCtx,
		Type:         obsType,
		ID:           id,
		TraceID:      traceID,
		ParentSpanID: parentSpanID,
	}, nil
}

// Update updates the observation with new data
func (o *Observation) Update(update interface{}) error {
	switch o.Type {
	case ObservationTypeSpan, ObservationTypeAgent, ObservationTypeTool, ObservationTypeChain,
		ObservationTypeRetriever, ObservationTypeEvaluator, ObservationTypeGuardrail:
		// All span-like types use SpanUpdate
		if spanUpdate, ok := update.(SpanUpdate); ok {
			_, err := o.client.UpdateSpan(o.ctx, o.ID, spanUpdate)
			return err
		}
		return fmt.Errorf("invalid update type for span-like observation")
	case ObservationTypeGeneration, ObservationTypeEmbedding:
		// Generation-like types use GenerationUpdate
		if genUpdate, ok := update.(GenerationUpdate); ok {
			_, err := o.client.UpdateGeneration(o.ctx, o.ID, genUpdate)
			return err
		}
		return fmt.Errorf("invalid update type for generation-like observation")
	case ObservationTypeEvent:
		// Events typically don't support updates, but we can add metadata
		return nil
	default:
		return fmt.Errorf(errUnknownObservationType, o.Type)
	}
}

// End ends the observation by updating it with end time
func (o *Observation) End() error {
	endTime := time.Now()

	switch o.Type {
	case ObservationTypeSpan, ObservationTypeAgent, ObservationTypeTool, ObservationTypeChain,
		ObservationTypeRetriever, ObservationTypeEvaluator, ObservationTypeGuardrail:
		// All span-like types use SpanUpdate for ending
		return o.Update(SpanUpdate{EndTime: &endTime})
	case ObservationTypeGeneration, ObservationTypeEmbedding:
		// Generation-like types use GenerationUpdate for ending
		return o.Update(GenerationUpdate{EndTime: &endTime})
	case ObservationTypeEvent:
		// Events don't have end times
		return nil
	default:
		return fmt.Errorf(errUnknownObservationType, o.Type)
	}
}

// Context returns the context with trace information
func (o *Observation) Context() context.Context {
	return o.ctx
}

// StartChildObservation starts a child observation within this observation's context
func (o *Observation) StartChildObservation(obsType ObservationType, name string, input interface{}) (*Observation, error) {
	return o.client.StartObservation(o.ctx, obsType, name, input)
}

// CurrentObservationKey is the context key for storing current observation
type CurrentObservationKey struct{}

// WithCurrentObservation stores the current observation in context
func WithCurrentObservation(ctx context.Context, obs *Observation) context.Context {
	return context.WithValue(ctx, CurrentObservationKey{}, obs)
}

// GetCurrentObservation retrieves the current observation from context
func GetCurrentObservation(ctx context.Context) (*Observation, bool) {
	obs, ok := ctx.Value(CurrentObservationKey{}).(*Observation)
	return obs, ok
}

// UpdateCurrentSpan updates the current span/observation from context
func (c *Client) UpdateCurrentSpan(ctx context.Context, output interface{}, metadata map[string]interface{}) error {
	obs, ok := GetCurrentObservation(ctx)
	if !ok {
		// Try to get from trace context and update via ID
		traceCtx, hasTrace := GetTraceContext(ctx)
		if !hasTrace || traceCtx.SpanID == "" {
			return fmt.Errorf("no current observation in context")
		}
		// Update as span by default
		_, err := c.UpdateSpan(ctx, traceCtx.SpanID, SpanUpdate{
			Output:   output,
			Metadata: metadata,
		})
		return err
	}

	switch obs.Type {
	case ObservationTypeSpan:
		return obs.Update(SpanUpdate{
			Output:   output,
			Metadata: metadata,
		})
	case ObservationTypeGeneration:
		return obs.Update(GenerationUpdate{
			Output:   output,
			Metadata: metadata,
		})
	default:
		return nil
	}
}

// StartAsCurrentSpan starts a span and stores it as the current observation in context
func (c *Client) StartAsCurrentSpan(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	// Apply propagated attributes if any
	attrs, hasAttrs := GetPropagatedAttributes(ctx)

	obs, err := c.StartObservation(ctx, ObservationTypeSpan, name, input)
	if err != nil {
		return ctx, nil, err
	}

	// Apply propagated attributes to the span
	if hasAttrs {
		obs.Update(SpanUpdate{
			Metadata: attrs.Metadata,
		})
	}

	// Store observation in context
	newCtx := WithCurrentObservation(obs.ctx, obs)

	return newCtx, obs, nil
}

// StartAsCurrentGeneration starts a generation and stores it as the current observation
func (c *Client) StartAsCurrentGeneration(ctx context.Context, name string, model string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartObservation(ctx, ObservationTypeGeneration, name, input)
	if err != nil {
		return ctx, nil, err
	}

	// Update with model
	obs.Update(GenerationUpdate{
		Model: &model,
	})

	// Store observation in context
	newCtx := WithCurrentObservation(obs.ctx, obs)

	return newCtx, obs, nil
}

// StartAgent starts an agent observation (reasoning blocks using LLM guidance)
func (c *Client) StartAgent(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeAgent, name, input)
}

// StartTool starts a tool observation (external tool calls, e.g., APIs)
func (c *Client) StartTool(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeTool, name, input)
}

// StartChain starts a chain observation (connecting LLM application steps)
func (c *Client) StartChain(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeChain, name, input)
}

// StartRetriever starts a retriever observation (data retrieval, e.g., vector stores)
func (c *Client) StartRetriever(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeRetriever, name, input)
}

// StartEmbedding starts an embedding observation (LLM embedding calls)
func (c *Client) StartEmbedding(ctx context.Context, name string, model string, input interface{}) (*Observation, error) {
	obs, err := c.StartObservation(ctx, ObservationTypeEmbedding, name, input)
	if err != nil {
		return nil, err
	}
	// Update with model
	obs.Update(GenerationUpdate{
		Model: &model,
	})
	return obs, nil
}

// StartEvaluator starts an evaluator observation (assessing LLM outputs)
func (c *Client) StartEvaluator(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeEvaluator, name, input)
}

// StartGuardrail starts a guardrail observation (protection against jailbreaks, etc.)
func (c *Client) StartGuardrail(ctx context.Context, name string, input interface{}) (*Observation, error) {
	return c.StartObservation(ctx, ObservationTypeGuardrail, name, input)
}

// StartAsCurrentAgent starts an agent and stores it as the current observation
func (c *Client) StartAsCurrentAgent(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartAgent(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentTool starts a tool and stores it as the current observation
func (c *Client) StartAsCurrentTool(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartTool(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentChain starts a chain and stores it as the current observation
func (c *Client) StartAsCurrentChain(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartChain(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentRetriever starts a retriever and stores it as the current observation
func (c *Client) StartAsCurrentRetriever(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartRetriever(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentEmbedding starts an embedding and stores it as the current observation
func (c *Client) StartAsCurrentEmbedding(ctx context.Context, name string, model string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartEmbedding(ctx, name, model, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentEvaluator starts an evaluator and stores it as the current observation
func (c *Client) StartAsCurrentEvaluator(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartEvaluator(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}

// StartAsCurrentGuardrail starts a guardrail and stores it as the current observation
func (c *Client) StartAsCurrentGuardrail(ctx context.Context, name string, input interface{}) (context.Context, *Observation, error) {
	obs, err := c.StartGuardrail(ctx, name, input)
	if err != nil {
		return ctx, nil, err
	}
	newCtx := WithCurrentObservation(obs.ctx, obs)
	return newCtx, obs, nil
}
