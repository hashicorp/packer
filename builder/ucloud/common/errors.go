package common

import (
	"fmt"
)

type NotFoundError struct {
	message string
}

type ExpectedStateError struct {
	message string
}

type NotCompletedError struct {
	message string
}

func (e *ExpectedStateError) Error() string {
	return e.message
}

func (e *NotFoundError) Error() string {
	return e.message
}

func (e *NotCompletedError) Error() string {
	return e.message
}

func NewNotFoundError(product, id string) error {
	return &NotFoundError{fmt.Sprintf("the %s %s is not found", product, id)}
}

func NewExpectedStateError(product, id string) error {
	return &ExpectedStateError{fmt.Sprintf("the %s %s not be expected state", product, id)}
}

func NewNotCompletedError(product string) error {
	return &NotCompletedError{fmt.Sprintf("%s is not completed", product)}
}

func IsNotFoundError(err error) bool {
	if _, ok := err.(*NotFoundError); ok {
		return true
	}
	return false
}

func IsExpectedStateError(err error) bool {
	if _, ok := err.(*ExpectedStateError); ok {
		return true
	}
	return false
}

func IsNotCompleteError(err error) bool {
	if _, ok := err.(*NotCompletedError); ok {
		return true
	}
	return false
}
