package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jasonhancock/go-logger"
)

// RequestIDFunc is a function that can be used to retrieve the request ID from
// a context.
type RequestIDFunc func(context.Context) string

// ClientIPFunc is a function that can be used to retrieve the client IP from
// a context.
type ClientIPFunc func(context.Context) string

// Responder writes API responses.
type Responder struct {
	log             *logger.L
	logResponseBody bool
	requestIDFunc   RequestIDFunc
	clientIPFunc    ClientIPFunc
}

// NewResponder makes a new Responder object.
func NewResponder(logger *logger.L, opts ...ResponderOption) *Responder {
	o := options{
		requestIDFunc: func(context.Context) string { return "" },
		clientIPFunc:  func(context.Context) string { return "" },
	}
	for _, opt := range opts {
		opt(&o)
	}

	return &Responder{
		log:             logger,
		logResponseBody: o.logResponseBodies,
		requestIDFunc:   o.requestIDFunc,
		clientIPFunc:    o.clientIPFunc,
	}
}

// With responds with the specified data.
func (r *Responder) With(w http.ResponseWriter, req *http.Request, status int, data any) {
	var buf bytes.Buffer
	// cannot write to buf if data is nil, in case of StatusNoContent, this write will fail
	// so we need an escape hatch here.
	if data != nil {
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "\t")
		err := enc.Encode(data)
		if err != nil {
			err = fmt.Errorf("failed to encode response object: %w", err)
			r.Err(w, req, err)
			return
		}
	}

	log := r.log.With("request_id", r.requestIDFunc(req.Context()))

	if r.logResponseBody {
		log.Debug("api_response",
			"status", status,
			"body", buf.String(),
		)
	} else {
		log.Debug("api_response", "status", status)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		log.LogError(
			"api_response_copy_error",
			fmt.Errorf("failed to copy response bytes: %w", err),
		)
	}
}

// Err responds with an error that corresponds to the behavior the type is illustrating.
func (r *Responder) Err(w http.ResponseWriter, req *http.Request, err error) {
	reqID := r.requestIDFunc(req.Context())
	logMsg := true
	defer func() {
		if logMsg {
			r.log.Err("api_response_error",
				"method", req.Method,
				"path", req.URL.Path,
				"query", req.URL.Query().Encode(),
				"error", err.Error(),
				"client_ip", r.clientIPFunc(req.Context()),
				"request_id", reqID,
			)
		}
	}()

	if httpErr, ok := err.(HTTP); ok {
		r.With(w, req, httpErr.StatusCode(), newErr(reqID, err.Error()))
		return
	}

	if nfErr, ok := err.(NotFounder); ok && nfErr.NotFound() {
		logMsg = false
		r.With(w, req, http.StatusNotFound, newErr(reqID, "resource not found"))
		return
	}

	if exErr, ok := err.(Exister); ok && exErr.Exists() {
		logMsg = false
		r.With(w, req, http.StatusUnprocessableEntity, newErr(reqID, "resource exists"))
		return
	}

	if cErr, ok := err.(Conflicter); ok && cErr.Conflict() {
		logMsg = false
		r.With(w, req, http.StatusUnprocessableEntity, newErr(reqID, "unprocessible entity"))
		return
	}

	r.With(w, req, http.StatusInternalServerError, newErr(reqID, "internal server error"))
}

type options struct {
	logResponseBodies bool
	requestIDFunc     RequestIDFunc
	clientIPFunc      ClientIPFunc
}

// ResponderOption is used to customize the API responder.
type ResponderOption func(*options)

// WithLogResponseBodies enables the logging of response bodies.
func WithLogResponseBodies(enabled bool) ResponderOption {
	return func(o *options) {
		o.logResponseBodies = enabled
	}
}

// WithRequestIDFunc sets the function to retrieve the request ID.
func WithRequestIDFunc(fn RequestIDFunc) ResponderOption {
	return func(o *options) {
		if fn == nil {
			return
		}
		o.requestIDFunc = fn
	}
}

// WithClientIPFunc sets the function to retrieve the client IP.
func WithClientIPFunc(fn ClientIPFunc) ResponderOption {
	return func(o *options) {
		if fn == nil {
			return
		}
		o.clientIPFunc = fn
	}
}
