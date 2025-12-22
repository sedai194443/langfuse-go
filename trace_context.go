package langfuse

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
)

// TraceContextKey is the context key for storing trace context
type TraceContextKey struct{}

// PropagatedAttributesKey is the context key for storing propagated attributes
type PropagatedAttributesKey struct{}

// TraceContext holds trace and observation IDs for context propagation
type TraceContext struct {
	TraceID      string
	SpanID       string // Current observation ID
	ParentSpanID string
}

// PropagatedAttributes holds attributes that propagate to all child spans
type PropagatedAttributes struct {
	SessionID string
	UserID    string
	Tags      []string
	Metadata  map[string]interface{}
}

// createTraceID generates a W3C-compliant trace ID (32-char hex, 16 bytes)
func createTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to deterministic generation if rand fails
		return createDeterministicTraceID(fmt.Sprintf("%d", len(bytes)))
	}
	return hex.EncodeToString(bytes)
}

// createDeterministicTraceID generates a deterministic trace ID from a seed
// This is useful for correlating external IDs with Langfuse traces
func createDeterministicTraceID(seed string) string {
	hash := fnv.New128a()
	hash.Write([]byte(seed))
	bytes := hash.Sum(nil)
	return hex.EncodeToString(bytes)
}

// createObservationID generates a W3C-compliant observation ID (16-char hex, 8 bytes)
func createObservationID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to deterministic generation if rand fails
		return createDeterministicObservationID(fmt.Sprintf("%d", len(bytes)))
	}
	return hex.EncodeToString(bytes)
}

// createDeterministicObservationID generates a deterministic observation ID from a seed
func createDeterministicObservationID(seed string) string {
	hash := fnv.New64a()
	hash.Write([]byte(seed))
	bytes := hash.Sum(nil)
	return hex.EncodeToString(bytes)
}

// CreateTraceID generates a W3C-compliant trace ID
// If seed is provided, the ID is deterministic (same seed = same ID)
// This is useful for correlating external IDs with Langfuse traces
func CreateTraceID(seed ...string) string {
	if len(seed) > 0 && seed[0] != "" {
		return createDeterministicTraceID(seed[0])
	}
	return createTraceID()
}

// CreateObservationID generates a W3C-compliant observation ID
// If seed is provided, the ID is deterministic (same seed = same ID)
func CreateObservationID(seed ...string) string {
	if len(seed) > 0 && seed[0] != "" {
		return createDeterministicObservationID(seed[0])
	}
	return createObservationID()
}

// WithTraceContext adds trace context to a Go context
func WithTraceContext(ctx context.Context, traceCtx TraceContext) context.Context {
	return context.WithValue(ctx, TraceContextKey{}, traceCtx)
}

// GetTraceContext retrieves trace context from a Go context
func GetTraceContext(ctx context.Context) (TraceContext, bool) {
	traceCtx, ok := ctx.Value(TraceContextKey{}).(TraceContext)
	return traceCtx, ok
}

// GetCurrentTraceID gets the current trace ID from context
func GetCurrentTraceID(ctx context.Context) (string, bool) {
	traceCtx, ok := GetTraceContext(ctx)
	if !ok {
		return "", false
	}
	return traceCtx.TraceID, traceCtx.TraceID != ""
}

// GetCurrentObservationID gets the current observation ID from context
func GetCurrentObservationID(ctx context.Context) (string, bool) {
	traceCtx, ok := GetTraceContext(ctx)
	if !ok {
		return "", false
	}
	return traceCtx.SpanID, traceCtx.SpanID != ""
}

// WithPropagatedAttributes adds propagated attributes to context
// These attributes will be automatically applied to all child spans/generations
func WithPropagatedAttributes(ctx context.Context, attrs PropagatedAttributes) context.Context {
	return context.WithValue(ctx, PropagatedAttributesKey{}, attrs)
}

// GetPropagatedAttributes retrieves propagated attributes from context
func GetPropagatedAttributes(ctx context.Context) (PropagatedAttributes, bool) {
	attrs, ok := ctx.Value(PropagatedAttributesKey{}).(PropagatedAttributes)
	return attrs, ok
}

// MergePropagatedAttributes merges existing propagated attributes with new ones
// New values override existing ones, tags and metadata are merged
func MergePropagatedAttributes(ctx context.Context, newAttrs PropagatedAttributes) context.Context {
	existing, ok := GetPropagatedAttributes(ctx)
	if !ok {
		return WithPropagatedAttributes(ctx, newAttrs)
	}

	merged := PropagatedAttributes{
		SessionID: newAttrs.SessionID,
		UserID:    newAttrs.UserID,
	}

	// Use existing if new is empty
	if merged.SessionID == "" {
		merged.SessionID = existing.SessionID
	}
	if merged.UserID == "" {
		merged.UserID = existing.UserID
	}

	// Merge tags
	tagSet := make(map[string]bool)
	for _, t := range existing.Tags {
		tagSet[t] = true
	}
	for _, t := range newAttrs.Tags {
		tagSet[t] = true
	}
	for t := range tagSet {
		merged.Tags = append(merged.Tags, t)
	}

	// Merge metadata
	merged.Metadata = make(map[string]interface{})
	for k, v := range existing.Metadata {
		merged.Metadata[k] = v
	}
	for k, v := range newAttrs.Metadata {
		merged.Metadata[k] = v
	}

	return WithPropagatedAttributes(ctx, merged)
}
