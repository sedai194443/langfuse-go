package langfuse

import (
	"github.com/google/uuid"
)

// NewTrace creates a new trace with a generated ID
func NewTrace(name string) Trace {
	return Trace{
		ID:   generateID(),
		Name: name,
	}
}

// NewSpan creates a new span with a generated ID
func NewSpan(traceID, name string) Span {
	return Span{
		ID:      generateID(),
		TraceID: traceID,
		Name:    name,
	}
}

// NewGeneration creates a new generation with a generated ID
func NewGeneration(traceID, name string) Generation {
	return Generation{
		ID:      generateID(),
		TraceID: traceID,
		Name:    name,
	}
}

// NewEvent creates a new event with a generated ID
func NewEvent(traceID, name string) Event {
	return Event{
		ID:      generateID(),
		TraceID: traceID,
		Name:    name,
	}
}

// NewScore creates a new score with a generated ID
func NewScore(traceID, name string, value float64) Score {
	return Score{
		ID:      generateID(),
		TraceID: traceID,
		Name:    name,
		Value:   value,
	}
}

// generateID generates a unique ID
// Note: For W3C Trace Context compliance, use CreateTraceID() or CreateObservationID()
// This function is kept for backward compatibility
func generateID() string {
	return uuid.New().String()
}

