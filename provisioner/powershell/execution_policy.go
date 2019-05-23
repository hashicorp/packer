package powershell

import (
	"fmt"
	"reflect"
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

func (ep *ExecutionPolicy) Decode(v interface{}) (err error) {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%#v is not a string", v)
	}
	*ep, err = ExecutionPolicyString(str)
	return err
}

func StringToExecutionPolicyHook(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Int {
		return data, nil
	}

	raw := data.(string)
	return ExecutionPolicyString(raw)
}
