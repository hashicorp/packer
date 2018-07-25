// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
//
// Helper methods for OCI GOSDK Samples
//

package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// LogIfError is equivalent to Println() followed by a call to os.Exit(1) if error is not nil
func LogIfError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

// RetryUntilTrueOrError retries a function until the predicate is true or it reaches a timeout.
// The operation is retried at the give frequency
func RetryUntilTrueOrError(operation func() (interface{}, error), predicate func(interface{}) (bool, error), frequency, timeout <-chan time.Time) error {
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout reached")
		case <-frequency:
			result, err := operation()
			if err != nil {
				return err
			}

			isTrue, err := predicate(result)
			if err != nil {
				return err
			}

			if isTrue {
				return nil
			}
		}
	}
}

// FindLifecycleFieldValue finds lifecycle value inside the struct based on reflection
func FindLifecycleFieldValue(request interface{}) (string, error) {
	val := reflect.ValueOf(request)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "", fmt.Errorf("can not unmarshal to response a pointer to nil structure")
		}
		val = val.Elem()
	}

	var err error
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if err != nil {
			return "", err
		}

		sf := typ.Field(i)

		//unexported
		if sf.PkgPath != "" {
			continue
		}

		sv := val.Field(i)

		if sv.Kind() == reflect.Struct {
			lif, err := FindLifecycleFieldValue(sv.Interface())
			if err == nil {
				return lif, nil
			}
		}
		if !strings.Contains(strings.ToLower(sf.Name), "lifecyclestate") {
			continue
		}
		return sv.String(), nil
	}
	return "", fmt.Errorf("request does not have a lifecycle field")
}

// CheckLifecycleState returns a function that checks for that a struct has the given lifecycle
func CheckLifecycleState(lifecycleState string) func(interface{}) (bool, error) {
	return func(request interface{}) (bool, error) {
		fieldLifecycle, err := FindLifecycleFieldValue(request)
		if err != nil {
			return false, err
		}
		isEqual := fieldLifecycle == lifecycleState
		log.Printf("Current lifecycle state is: %s, waiting for it becomes to: %s", fieldLifecycle, lifecycleState)
		return isEqual, nil
	}
}

// GetRandomString returns a random string with length equals to n
func GetRandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
