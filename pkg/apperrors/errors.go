package apperrors

import "errors"

// Typed sentinel errors for use across all layers.
// HTTP handlers map these to specific status codes.

var (
	ErrNotFound        = errors.New("not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrConflict        = errors.New("conflict")
	ErrRateLimit       = errors.New("rate limit exceeded")
	ErrValidation      = errors.New("validation error")
	ErrInternal        = errors.New("internal error")
	ErrLLMUnavailable  = errors.New("llm unavailable")
	ErrParseFailure    = errors.New("parse failure")
	ErrJobNotFound     = errors.New("job not found")
)

// AppError wraps a sentinel with a human-readable message and optional code.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func New(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func IsNotFound(err error) bool        { return errors.Is(err, ErrNotFound) }
func IsUnauthorized(err error) bool    { return errors.Is(err, ErrUnauthorized) }
func IsConflict(err error) bool        { return errors.Is(err, ErrConflict) }
func IsRateLimit(err error) bool       { return errors.Is(err, ErrRateLimit) }
func IsValidation(err error) bool      { return errors.Is(err, ErrValidation) }
func IsLLMUnavailable(err error) bool  { return errors.Is(err, ErrLLMUnavailable) }
