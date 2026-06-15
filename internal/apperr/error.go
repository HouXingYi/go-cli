package apperr

import "fmt"

const (
	KindConfig  = "config"
	KindNetwork = "network"
	KindAuth    = "auth"
	KindAPI     = "api"
	KindOutput  = "output"
)

type Error struct {
	Kind    string
	Message string
	Advice  string
	Err     error
}

func New(kind, message, advice string) *Error {
	return &Error{
		Kind:    kind,
		Message: message,
		Advice:  advice,
	}
}

func Wrap(kind, message, advice string, err error) *Error {
	return &Error{
		Kind:    kind,
		Message: message,
		Advice:  advice,
		Err:     err,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Hint(err error) string {
	type hinter interface {
		Hint() string
	}

	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.Advice
	}
	if e, ok := err.(hinter); ok {
		return e.Hint()
	}
	return ""
}

func Kind(err error) string {
	if e, ok := err.(*Error); ok {
		return e.Kind
	}
	return ""
}

func (e *Error) Hint() string {
	if e == nil {
		return ""
	}
	return e.Advice
}
