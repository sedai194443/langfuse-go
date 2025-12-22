package langfuse

import (
	"errors"
	"fmt"
)

// APIError represents an error returned by the Langfuse API
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("langfuse API error (status %d): %s - %s", e.StatusCode, e.Message, e.Body)
	}
	return fmt.Sprintf("langfuse API error (status %d): %s", e.StatusCode, e.Message)
}

// Unwrap returns nil as APIError is a leaf error type
func (e *APIError) Unwrap() error {
	return nil
}

// IsAPIError checks if an error is an APIError
// It uses errors.As to properly handle wrapped errors
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}
