package apperr

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Code + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.Cause }

// New builds an error carrying the standard message for the code.
func New(code string, cause error) *Error {
	return &Error{Code: code, Message: Message(code), Cause: cause}
}

// Withf is for messages that need a runtime value, such as a required size.
func Withf(code string, cause error, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// From converts any error into one that is safe to show a user. Anything that
// is not already an Error becomes INTERNAL, so raw causes never reach the UI.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return New(CodeInternal, err)
}

func Is(err error, code string) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}
