package stateErrors

import "strings"

type ErrInvalidInput struct {
	error
	message string
	wrapped error
}

func (err ErrInvalidInput) Error() string {
	if err.message != "" && err.wrapped.Error() != "" {
		return strings.Join([]string{err.message, err.wrapped.Error()}, "\n")
	}
	if err.message != "" {
		return err.message
	}
	return err.wrapped.Error()
}

func NewInvalidInput(message string) error {
	return &ErrInvalidInput{message: message}
}

func NewInvalidInputWrapped(err error) error {
	if err != nil {
		return &ErrInvalidInput{wrapped: err}
	}
	return nil
}
