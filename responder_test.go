package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jasonhancock/go-logger"
	"github.com/stretchr/testify/require"
)

func TestResponderUserMessage(t *testing.T) {
	tests := []struct {
		desc     string
		err      error
		expected string
	}{
		{"not found", &notFounder{}, "resource not found"},
		{"not found custom message", &notFounderMessage{}, "not found custom message"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			r, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)

			responder := NewResponder(logger.Default())
			w := httptest.NewRecorder()
			responder.Err(w, r, tt.err)

			require.Equal(t, http.StatusNotFound, w.Code)
			var decoded ErrorResponse
			require.NoError(t, json.NewDecoder(w.Body).Decode(&decoded))
			require.Equal(t, tt.expected, decoded.Error.Message)
		})
	}
}

type notFounder struct{}

func (e *notFounder) Error() string  { return "not found" }
func (e *notFounder) NotFound() bool { return true }

type notFounderMessage struct{}

func (e *notFounderMessage) Error() string       { return "not found" }
func (e *notFounderMessage) UserMessage() string { return "not found custom message" }
func (e *notFounderMessage) NotFound() bool      { return true }
