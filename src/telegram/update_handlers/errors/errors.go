package customErrors

import (
	"errors"
	"strings"
)

type ErrInvalidInput struct {
	message string
	wrapped error
}

func (err ErrInvalidInput) Error() string {
	if err.message != "" && err.wrapped != nil {
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

var ErrNoLabworks = errors.New("no labworks found")
