package packer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestProvisionHook_Impl(t *testing.T) {
	var raw interface{}
	raw = &ProvisionHook{}
	if _, ok := raw.(packersdk.Hook); !ok {
		t.Fatalf("must be a Hook")
	}
}

func TestProvisionHook(t *testing.T) {
	pA := &packersdk.MockProvisioner{}
	pB := &packersdk.MockProvisioner{}

	ui := testUi()
	var comm packersdk.Communicator = new(packersdk.MockCommunicator)
	var data interface{} = nil

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{pA, nil, ""},
			{pB, nil, ""},
		},
	}

	hook.Run(context.Background(), "foo", ui, comm, data)

	if !pA.ProvCalled {
		t.Error("provision should be called on pA")
	}

	if !pB.ProvCalled {
		t.Error("provision should be called on pB")
	}
}

func TestProvisionHook_nilComm(t *testing.T) {
	pA := &packersdk.MockProvisioner{}
	pB := &packersdk.MockProvisioner{}

	ui := testUi()
	var comm packersdk.Communicator = nil
	var data interface{} = nil

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{pA, nil, ""},
			{pB, nil, ""},
		},
	}

	err := hook.Run(context.Background(), "foo", ui, comm, data)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestProvisionHook_cancel(t *testing.T) {
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	p := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			topCtxCancel()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{p, nil, ""},
		},
	}

	err := hook.Run(topCtx, "foo", nil, new(packersdk.MockCommunicator), nil)
	if err == nil {
		t.Fatal("should have err")
	}
}

// TODO(mitchellh): Test that they're run in the proper order

func TestPausedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(PausedProvisioner)
}

func TestPausedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	prov.Prepare(42)
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestPausedProvisionerProvision(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	prov.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestPausedProvisionerProvision_cancel(t *testing.T) {
	pausingCtx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	startTime := time.Now()
	waitTime := 50 * time.Millisecond

	ui := new(packersdk.MockUi)
	prov := pausedTestProvisionor(startTime, waitTime)
	err := prov.Provision(pausingCtx, ui, new(packersdk.MockCommunicator), make(map[string]interface{}))
	if err == nil {
		t.Fatalf("No error occurred")
	}

	expectedError := "context deadline exceeded"
	if err.Error() != expectedError {
		t.Fatalf("Unexpected error, expected: %v, got: %v", expectedError, err)
	}
}

func pausedTestProvisionor(startTime time.Time, waitTime time.Duration) *PausedProvisioner {
	return &PausedProvisioner{
		PauseBefore: waitTime,
		Provisioner: &packersdk.MockProvisioner{
			ProvFunc: func(context.Context) error {
				timeSinceStartTime := time.Since(startTime)
				if timeSinceStartTime < waitTime {
					return fmt.Errorf("Spent not enough time waiting: %s", timeSinceStartTime)
				}
				return nil
			},
		},
	}
}

// TODO update this test as well to make sure that updates are not happening when less than 10 seconds
func TestPausedProvisionerProvision_waits(t *testing.T) {
	startTime := time.Now()
	waitTime := 50 * time.Millisecond

	ui := new(packersdk.MockUi)
	prov := pausedTestProvisionor(startTime, waitTime)
	err := prov.Provision(context.Background(), ui, new(packersdk.MockCommunicator), make(map[string]interface{}))

	if err != nil {
		t.Fatalf("prov failed: %v", err)
	}

	expectedMessages := []string{
		fmt.Sprintf("Pausing %s before the next provisioner...", waitTime),
	}

	if ui.SayMessages[0].Message != expectedMessages[0] {
		t.Fatalf("expected: %s, got: %s", expectedMessages[0], ui.SayMessages[0].Message)
	}

	if len(ui.SayMessages) != len(expectedMessages) {
		t.Fatalf(
			"length of expected messages different than messages sent. Expected: %v, got: %v, messages: %v",
			len(ui.SayMessages),
			len(expectedMessages),
			ui.SayMessages)
	}
}

func TestPausedProvisionerProvision_waits_MachineReadableUi(t *testing.T) {
	startTime := time.Now()
	waitTime := 15 * time.Second

	var buffer bytes.Buffer
	prov := pausedTestProvisionor(startTime, waitTime)
	ui := MachineReadableUi{
		Writer: &buffer,
	}
	err := prov.Provision(context.Background(), &ui, new(packersdk.MockCommunicator), make(map[string]interface{}))

	if err != nil {
		t.Fatalf("prov failed: %v", err)
	}

	bufferString := buffer.String()
	splitByLine := strings.Split(bufferString, "\n")

	if len(splitByLine) != 2 {
		t.Fatalf(fmt.Sprintf("Too many lines in buffer: %v, expected: %v,  buffer: %v, ", len(splitByLine), 2, bufferString))
	}

	// First message should be the default message
	splitByComma := strings.Split(splitByLine[0], ",")
	// because of the messages coming in via ui.Say function, the message is the 4th entry per line in buffer
	message := splitByComma[4]
	firstMessage := "Pausing 15s before the next provisioner..."
	if message != firstMessage {
		t.Fatalf(fmt.Sprintf("First message does not match what is expected. Expected: %v, got: %v", firstMessage, message))
	}

	if splitByLine[1] != "" {
		t.Fatalf(fmt.Sprintf("Second message in buffer is not empty:  %v", splitByLine[1]))
	}
}

func TestPausedProvisionerProvision_waits_with_updates(t *testing.T) {
	startTime := time.Now()
	waitTime := 30 * time.Second

	prov := pausedTestProvisionor(startTime, waitTime)
	ui := new(packersdk.MockUi)
	currentTime := time.Now()
	err := prov.Provision(context.Background(), ui, new(packersdk.MockCommunicator), make(map[string]interface{}))

	if err != nil {
		t.Fatalf("prov failed: %v", err)
	}

	expectedMessages := []string{
		fmt.Sprintf("Pausing %s before the next provisioner...", waitTime),
		"20 seconds left until the next provisioner",
		"10 seconds left until the next provisioner",
	}

	if ui.SayMessages[0].Message != expectedMessages[0] {
		t.Fatalf("expected: %s, got: %s", expectedMessages[0], ui.SayMessages[0].Message)
	}

	if len(ui.SayMessages) != len(expectedMessages) {
		t.Fatalf(
			"length of expected messages different than messages sent. Expected: %v, got: %v, messages: %v",
			len(ui.SayMessages),
			len(expectedMessages),
			ui.SayMessages)
	}

	lastTime := currentTime
	for index, message := range expectedMessages {
		// Skiping first message as this has already been verified
		if index == 0 {
			continue
		}

		if ui.SayMessages[index].Message != message {
			t.Fatalf("expected: %s, got: %s", message, ui.SayMessages[index].Message)
		}

		waitTimeBetweenMessages := math.Round(ui.SayMessages[index].SayTime.Sub(lastTime).Seconds())
		if waitTimeBetweenMessages != 10 {
			t.Fatalf("Did not wait the appropriate amount of time message: %v", ui.SayMessages[index].Message)
		}

		// setting last time to current SayTime
		lastTime = ui.SayMessages[index].SayTime
	}
}

func TestPausedProvisionerCancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestDebuggedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(DebuggedProvisioner)
}

func TestDebuggedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	prov.Prepare(42)
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestDebuggedProvisionerProvision(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	writeReader(ui, "\n")
	prov.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestDebuggedProvisionerCancel(t *testing.T) {
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		topCtxCancel()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestRetriedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(RetriedProvisioner)
}

func TestRetriedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &RetriedProvisioner{
		Provisioner: mock,
	}

	err := prov.Prepare(42)
	if err != nil {
		t.Fatal("should not have errored")
	}
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestRetriedProvisionerProvision(t *testing.T) {
	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			return errors.New("failed")
		},
	}

	prov := &RetriedProvisioner{
		MaxRetries:  2,
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	err := prov.Provision(context.Background(), ui, comm, make(map[string]interface{}))
	if err != nil {
		t.Fatal("should not have errored")
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if !mock.ProvRetried {
		t.Fatal("prov should be retried")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestRetriedProvisionerCancelledProvision(t *testing.T) {
	// Don't retry if context is cancelled
	ctx, topCtxCancel := context.WithCancel(context.Background())

	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			topCtxCancel()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	prov := &RetriedProvisioner{
		MaxRetries:  2,
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	err := prov.Provision(ctx, ui, comm, make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have errored")
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvRetried {
		t.Fatal("prov should NOT be retried")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestRetriedProvisionerCancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &RetriedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]interface{}))
	if err == nil {
		t.Fatal("should have err")
	}
}
