package api

// ErrorResponse is the body of an Error served up by Err
type ErrorResponse struct {
	RequestID string    `json:"request_id"`
	Error     ErrorData `json:"error"`
}

type ErrorData struct {
	Message string `json:"message"`
}

func newErr(requestID, msg string) ErrorResponse {
	return ErrorResponse{
		RequestID: requestID,
		Error: ErrorData{
			Message: msg,
		},
	}
}

// HTTP determines if an error exhibits the behavior of an error that provides an
// HTTP status code and the error message itself is safe for returning to a
// client.
type HTTP interface {
	StatusCode() int
}

var _ HTTP = (*HTTPErr)(nil)

// HTTPErr is a error that provides a status code.
type HTTPErr struct {
	error
	Code int
}

// NewHTTPErr generates a new HTTPErr.
func NewHTTPErr(err error, code int) *HTTPErr {
	return &HTTPErr{
		error: err,
		Code:  code,
	}
}

// Unwrap supports the Go 1.13 error semantics.
func (h *HTTPErr) Unwrap() error {
	return h.error
}

// StatusCode provides the status code associated with the error message.
func (h *HTTPErr) StatusCode() int {
	return h.Code
}

// NotFounder determines if an error exhibits the behavior of a resource not found err.
type NotFounder interface {
	NotFound() bool
}

// Exister determines if an error exhibits the behavior of a resource already exists err.
type Exister interface {
	Exists() bool
}

// Conflicter determines if an error exhibits the behavior of a conflict err. This corresponds
// to times when you receive unexpected errors or an error of high severity.
type Conflicter interface {
	Conflict() bool
}

// UserMessager can be used to present a specific message to the user.
type UserMessager interface {
	UserMessage() string
}
