// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkerrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMuliterr(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, Append(nil, nil))
		assert.Nil(t, Errors(nil))
	})
	t.Run("lhs", func(t *testing.T) {
		x := fmt.Errorf("x")
		assert.Equal(t, x, Append(x, nil))
	})
	t.Run("rhs", func(t *testing.T) {
		x := fmt.Errorf("x")
		assert.Equal(t, x, Append(nil, x))
	})
	t.Run("mulit", func(t *testing.T) {
		x1 := fmt.Errorf("x1")
		x2 := fmt.Errorf("x2")
		x12 := Append(x1, x2)
		assert.Equal(t, []error{x1, x2}, Errors(x12))

		x3 := fmt.Errorf("x3")
		x123 := Append(x12, x3)
		assert.Equal(t, []error{x1, x2, x3}, Errors(x123))
		// Check still unchanged
		assert.Equal(t, []error{x1, x2}, Errors(x12))
		assert.Equal(t, "x1\nx2\nx3", x123.Error())
	})
}
