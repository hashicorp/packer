// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package sdkerrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWithMessage_StatusErr(t *testing.T) {
	_, ok := status.FromError(WithMessage(errors.New("err"), "msg"))
	assert.False(t, ok)

	expectedStatus := status.New(codes.InvalidArgument, "invalid argument")
	statusWithMessage := WithMessage(expectedStatus.Err(), "msg")

	st, ok := status.FromError(statusWithMessage)
	assert.True(t, ok)
	assert.Equal(t, expectedStatus, st)

	st, ok = status.FromError(WithMessage(statusWithMessage, "extra msg"))
	assert.True(t, ok)
	assert.Equal(t, expectedStatus, st)
}
