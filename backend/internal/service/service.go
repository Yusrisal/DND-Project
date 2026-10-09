package service

import (
	"errors"
	"fmt"
)

// These sentinels plus ValidationError are the failures handlers translate into
// status codes; anything else is reported as a 500.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }

func invalid(field, message string) error { return ValidationError{Field: field, Message: message} }
