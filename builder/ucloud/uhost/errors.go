package uhost

import (
	"fmt"
)

type NotFoundError struct {
	message string
}

type ExpectedStateError struct {
	message string
}

type NotCompleteError struct {
	message string
}

func (e *ExpectedStateError) Error() string {
	return e.message
}

func (e *NotFoundError) Error() string {
	return e.message
}

func (e *NotCompleteError) Error() string {
	return e.message
}

func newNotFoundError(product, id string) error {
	return &NotFoundError{fmt.Sprintf("the %s %s is not found", product, id)}
}

func newExpectedStateError(product, id string) error {
	return &ExpectedStateError{fmt.Sprintf("the %s %s not be expected state", product, id)}
}

func newNotCompleteError(product string) error {
	return &NotCompleteError{fmt.Sprintf("%s is not completed", product)}
}

func isNotFoundError(err error) bool {
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}

func isExpectedStateError(err error) bool {
	if _, ok := err.(*ExpectedStateError); ok {
		return true
	}
	return false
}

func isNotCompleteError(err error) bool {
	if _, ok := err.(*NotCompleteError); ok {
		return true
	}
	return false
}
