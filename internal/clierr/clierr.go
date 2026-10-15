package clierr

import "errors"

var (
	ErrUsage             = errors.New("usage error")
	ErrDockerUnavailable = errors.New("docker unavailable")
	ErrTimeout           = errors.New("operation timed out")
	ErrMutation          = errors.New("mutation failed")
)

// Error preserves the user-facing message while exposing a stable category.
type Error struct {
	Kind error
	Err  error
}

func (e *Error) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() []error {
	if e == nil {
		return nil
	}
	return []error{e.Kind, e.Err}
}

func Wrap(kind, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Err: err}
}
