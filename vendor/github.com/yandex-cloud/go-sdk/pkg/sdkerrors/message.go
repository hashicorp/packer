// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkerrors

import (
	"fmt"
)

type errWithMessage struct {
	err     error
	message string
}

func (e *errWithMessage) Error() string {
	return e.message + ": " + e.err.Error()
}

func (e *errWithMessage) Cause() error {
	return e.err
}

func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &errWithMessage{err, message}
}

func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &errWithMessage{err, fmt.Sprintf(format, args...)}
}
