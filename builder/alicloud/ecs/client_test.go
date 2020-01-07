package ecs

import (
	"fmt"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

func TestWaitForExpectedExceedRetryTimes(t *testing.T) {
	c := ClientWrapper{}

	iter := 0
	waitDone := make(chan bool, 1)

	go func() {
		_, _ = c.WaitForExpected(&WaitForExpectArgs{
			RequestFunc: func() (responses.AcsResponse, error) {
				iter++
				return nil, fmt.Errorf("test: let iteration %d failed", iter)
			},
			EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
				if err != nil {
					fmt.Printf("need retry: %s\n", err)
					return WaitForExpectToRetry
				}

				return WaitForExpectSuccess
			},
		})

		waitDone <- true
	}()

	select {
	case <-waitDone:
		if iter != defaultRetryTimes {
			t.Fatalf("WaitForExpected should terminate at the %d iterations", defaultRetryTimes)
		}
	}
}

func TestWaitForExpectedExceedRetryTimeout(t *testing.T) {
	c := ClientWrapper{}

	expectTimeout := 10 * time.Second
	iter := 0
	waitDone := make(chan bool, 1)

	go func() {
		_, _ = c.WaitForExpected(&WaitForExpectArgs{
			RequestFunc: func() (responses.AcsResponse, error) {
				iter++
				return nil, fmt.Errorf("test: let iteration %d failed", iter)
			},
			EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
				if err != nil {
					fmt.Printf("need retry: %s\n", err)
					return WaitForExpectToRetry
				}

				return WaitForExpectSuccess
			},
			RetryTimeout: expectTimeout,
		})

		waitDone <- true
	}()

	timeTolerance := 1 * time.Second
	select {
	case <-waitDone:
		if iter > int(expectTimeout/defaultRetryInterval) {
			t.Fatalf("WaitForExpected should terminate before the %d iterations", int(expectTimeout/defaultRetryInterval))
		}
	case <-time.After(expectTimeout + timeTolerance):
		t.Fatalf("WaitForExpected should terminate within %f seconds", (expectTimeout + timeTolerance).Seconds())
	}
}
