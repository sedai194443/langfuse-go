package langfuse

import (
	"context"
	"fmt"
	"time"
)

// ObservationType represents the type of observation
type ObservationType string

const (
	// ObservationTypeSpan represents a span observation
	ObservationTypeSpan ObservationType = "span"
	// ObservationTypeGeneration represents a generation observation
	ObservationTypeGeneration ObservationType = "generation"
	// ObservationTypeEvent represents an event observation
	ObservationTypeEvent ObservationType = "event"
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
	case ObservationTypeSpan:
		span := Span{
			ID:                  observationID,
			TraceID:             traceID,
			Name:                name,
			StartTime:           &startTime,
			Input:               input,
			ParentObservationID: parentSpanID,
		}
		resp, err := c.CreateSpan(ctx, span)
		if err != nil {
			return nil, err
		}
		id = resp.ID
	case ObservationTypeGeneration:
		gen := Generation{
			ID:                  observationID,
			TraceID:             traceID,
			Name:                name,
			StartTime:           &startTime,
			Input:               input,
			ParentObservationID: parentSpanID,
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
		return nil, fmt.Errorf("unknown observation type: %s", obsType)
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
	case ObservationTypeSpan:
		if spanUpdate, ok := update.(SpanUpdate); ok {
			_, err := o.client.UpdateSpan(o.ctx, o.ID, spanUpdate)
			return err
		}
		return fmt.Errorf("invalid update type for span")
	case ObservationTypeGeneration:
		if genUpdate, ok := update.(GenerationUpdate); ok {
			_, err := o.client.UpdateGeneration(o.ctx, o.ID, genUpdate)
			return err
		}
		return fmt.Errorf("invalid update type for generation")
	case ObservationTypeEvent:
		// Events typically don't support updates, but we can add metadata
		return nil
	default:
		return fmt.Errorf("unknown observation type: %s", o.Type)
	}
}

// End ends the observation by updating it with end time
func (o *Observation) End() error {
	endTime := time.Now()

	switch o.Type {
	case ObservationTypeSpan:
		return o.Update(SpanUpdate{EndTime: &endTime})
	case ObservationTypeGeneration:
		return o.Update(GenerationUpdate{EndTime: &endTime})
	case ObservationTypeEvent:
		// Events don't have end times
		return nil
	default:
		return fmt.Errorf("unknown observation type: %s", o.Type)
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
