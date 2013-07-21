package ebs

import (
	"github.com/mitchellh/goamz/ec2"
	"log"
	"testing"
	"time"
)

type debuggingUi struct {
}

func (*debuggingUi) Ask(message string) (string, error) {
	panic("not implemented")
}

func (*debuggingUi) Say(message string) {
	log.Println(message)
}

func (*debuggingUi) Message(message string) {
	log.Println(message)
}

func (*debuggingUi) Error(message string) {
	log.Println(message)
}

var (
	ui      = &debuggingUi{}
	timeout = 500 * time.Millisecond
	wait    = 50 * time.Second
)

func TestWaitForAMI(t *testing.T) {
	// Given
	failingImageFetcher := func() (*ec2.ImagesResp, error) {
		response := &ec2.ImagesResp{}
		response.Images = []ec2.Image{ec2.Image{State: "available"}}
		return response, nil
	}

	// When
	err := waitForAMI(ui, failingImageFetcher, timeout, wait)

	// Then
	if err != nil {
		panic("expected waitForAMI to pass successfully")
	}
}

func TestWaitForAMIFailsImmediately(t *testing.T) {
	// Given
	failingImageFetcher := func() (*ec2.ImagesResp, error) {
		err := &ec2.Error{}
		return nil, err
	}

	started := time.Now()

	// When
	err := waitForAMI(ui, failingImageFetcher, timeout, wait)

	// Then
	if err == nil {
		panic("expected waitForAMI to fail")
	}
	if time.Since(started) > wait {
		panic("expected to have failed immediately")
	}
}

func TestWaitForAMIIgnoresInvalidAMIID(t *testing.T) {
	// Given
	failingImageFetcher := func() (*ec2.ImagesResp, error) {
		err := &ec2.Error{Code: "InvalidAMIID.NotFound"}
		return nil, err
	}

	started := time.Now()

	// When
	err := waitForAMI(ui, failingImageFetcher, timeout, wait)

	// Then
	if err == nil {
		panic("expected waitForAMI to fail")
	}
	if time.Since(started) < timeout {
		panic("expected elapsed time > timeout since the reason we failed was InvalidAMIID.NotFound")
	}
}
