package langfuse

import (
	"net/http"
	"testing"
)

func TestAPIError(t *testing.T) {
	err := &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid request",
		Body:       `{"error": "bad request"}`,
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("APIError.Error() should return non-empty string")
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("APIError.StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
	}
}

func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "APIError",
			err:  &APIError{StatusCode: http.StatusBadRequest},
			want: true,
		},
		{
			name: "non-APIError",
			err:  &testError{msg: "test"},
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAPIError(tt.err); got != tt.want {
				t.Errorf("IsAPIError() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

