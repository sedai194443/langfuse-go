package langfuse

import (
	"testing"
)

func TestNewTrace(t *testing.T) {
	trace := NewTrace("test-trace")
	if trace.ID == "" {
		t.Error("NewTrace() ID should not be empty")
	}
	if trace.Name != "test-trace" {
		t.Errorf("NewTrace() Name = %v, want test-trace", trace.Name)
	}
}

func TestNewSpan(t *testing.T) {
	span := NewSpan("trace-123", "test-span")
	if span.ID == "" {
		t.Error("NewSpan() ID should not be empty")
	}
	if span.Name != "test-span" {
		t.Errorf("NewSpan() Name = %v, want test-span", span.Name)
	}
	if span.TraceID != "trace-123" {
		t.Errorf("NewSpan() TraceID = %v, want trace-123", span.TraceID)
	}
}

func TestNewGeneration(t *testing.T) {
	gen := NewGeneration("trace-123", "test-generation")
	if gen.ID == "" {
		t.Error("NewGeneration() ID should not be empty")
	}
	if gen.Name != "test-generation" {
		t.Errorf("NewGeneration() Name = %v, want test-generation", gen.Name)
	}
	if gen.TraceID != "trace-123" {
		t.Errorf("NewGeneration() TraceID = %v, want trace-123", gen.TraceID)
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent("trace-123", "test-event")
	if event.ID == "" {
		t.Error("NewEvent() ID should not be empty")
	}
	if event.Name != "test-event" {
		t.Errorf("NewEvent() Name = %v, want test-event", event.Name)
	}
	if event.TraceID != "trace-123" {
		t.Errorf("NewEvent() TraceID = %v, want trace-123", event.TraceID)
	}
}

func TestNewScore(t *testing.T) {
	score := NewScore("trace-123", "quality", 0.95)
	if score.ID == "" {
		t.Error("NewScore() ID should not be empty")
	}
	if score.Name != "quality" {
		t.Errorf("NewScore() Name = %v, want quality", score.Name)
	}
	if score.Value != 0.95 {
		t.Errorf("NewScore() Value = %v, want 0.95", score.Value)
	}
	if score.TraceID != "trace-123" {
		t.Errorf("NewScore() TraceID = %v, want trace-123", score.TraceID)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == id2 {
		t.Error("generateID() should generate unique IDs")
	}
	if len(id1) == 0 {
		t.Error("generateID() should generate non-empty ID")
	}
}

