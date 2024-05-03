// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate enumer -transform snake -trimprefix ExecutionPolicy -type ExecutionPolicy

package powershell

import (
	"reflect"
	"strconv"
)

// ExecutionPolicy setting to run the command(s).
// For the powershell provider the default has historically been to bypass.
type ExecutionPolicy int

const (
	ExecutionPolicyBypass ExecutionPolicy = iota
	ExecutionPolicyAllsigned
	ExecutionPolicyDefault
	ExecutionPolicyRemotesigned
	ExecutionPolicyRestricted
	ExecutionPolicyUndefined
	ExecutionPolicyUnrestricted
	ExecutionPolicyNone // not set
)

func StringToExecutionPolicyHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Int {
		return data, nil
	}

	raw := data.(string)
	// It's possible that the thing being read is not supposed to be an
	// execution policy; if the string provided is actally an int, just return
	// the int.
	i, err := strconv.Atoi(raw)
	if err == nil {
		return i, nil
	}
	// If it can't just be cast to an int, try to parse string into an
	// execution policy.
	return ExecutionPolicyString(raw)
}
