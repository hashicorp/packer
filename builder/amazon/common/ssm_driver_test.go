package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func NewSSMDriverWithMockSvc(svc *MockSSMSvc) *SSMDriver {
	config := SSMDriverConfig{
		SvcClient:   svc,
		Region:      "east",
		ProfileName: "default",
		SvcEndpoint: "example.com",
	}

	driver := SSMDriver{
		SSMDriverConfig: config,
		pluginCmdFunc:   func(ctx context.Context) error { return nil },
	}

	return &driver
}
func TestSSMDriver_StartSession(t *testing.T) {
	mockSvc := MockSSMSvc{}
	driver := NewSSMDriverWithMockSvc(&mockSvc)

	if driver.SvcClient == nil {
		t.Fatalf("SvcClient for driver should not be nil")
	}

	session, err := driver.StartSession(context.TODO(), MockStartSessionInput("fakeinstance"))
	if err != nil {
		t.Fatalf("calling StartSession should not error but got %v", err)
	}

	if !mockSvc.StartSessionCalled {
		t.Fatalf("expected test to call ssm mocks but didn't")
	}

	if session == nil {
		t.Errorf("expected session to be set after a successful call to StartSession")
	}

	if !reflect.DeepEqual(session, MockStartSessionOutput()) {
		t.Errorf("expected session to be %v but got %v", MockStartSessionOutput(), session)
	}
}

func TestSSMDriver_StartSessionWithError(t *testing.T) {
	mockSvc := MockSSMSvc{StartSessionError: fmt.Errorf("bogus error")}
	driver := NewSSMDriverWithMockSvc(&mockSvc)

	if driver.SvcClient == nil {
		t.Fatalf("SvcClient for driver should not be nil")
	}

	session, err := driver.StartSession(context.TODO(), MockStartSessionInput("fakeinstance"))
	if err == nil {
		t.Fatalf("StartSession should have thrown an error but didn't")
	}

	if !mockSvc.StartSessionCalled {
		t.Errorf("expected test to call StartSession mock but didn't")
	}

	if session != nil {
		t.Errorf("expected session to be nil after a bad StartSession call, but got %v", session)
	}
}

func TestSSMDriver_StopSession(t *testing.T) {
	mockSvc := MockSSMSvc{}
	driver := NewSSMDriverWithMockSvc(&mockSvc)

	if driver.SvcClient == nil {
		t.Fatalf("SvcClient for driver should not be nil")
	}

	// Calling StopSession before StartSession should fail
	err := driver.StopSession()
	if err == nil {
		t.Fatalf("calling StopSession() on a driver that has no started session should fail")
	}

	if driver.session != nil {
		t.Errorf("expected session to be default to nil")
	}

	if mockSvc.TerminateSessionCalled {
		t.Fatalf("a call to TerminateSession should not occur when there is no valid SSM session")
	}

	// Lets try calling start session, then stopping to see what happens.
	session, err := driver.StartSession(context.TODO(), MockStartSessionInput("fakeinstance"))
	if err != nil {
		t.Fatalf("calling StartSession should not error but got %v", err)
	}

	if !mockSvc.StartSessionCalled {
		t.Fatalf("expected test to call StartSession mock but didn't")
	}

	if session == nil || driver.session != session {
		t.Errorf("expected session to be set after a successful call to StartSession")
	}

	if !reflect.DeepEqual(session, MockStartSessionOutput()) {
		t.Errorf("expected session to be %v but got %v", MockStartSessionOutput(), session)
	}

	err = driver.StopSession()
	if err != nil {
		t.Errorf("calling StopSession() on a driver on a started session should not fail")
	}

	if !mockSvc.TerminateSessionCalled {
		t.Fatalf("expected test to call StopSession mock but didn't")
	}

}

func TestSSMDriver_Args(t *testing.T) {
	tt := []struct {
		Name             string
		ProfileName      string
		SkipStartSession bool
		ErrorExpected    bool
	}{
		{
			Name:             "NilSession",
			SkipStartSession: true,
			ErrorExpected:    true,
		},
		{
			Name:          "NonNilSession",
			ErrorExpected: false,
		},
		{
			Name:          "SessionWithProfileName",
			ProfileName:   "default",
			ErrorExpected: false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			mockSvc := MockSSMSvc{}
			driver := NewSSMDriverWithMockSvc(&mockSvc)
			driver.ProfileName = tc.ProfileName

			if driver.SvcClient == nil {
				t.Fatalf("svcclient for driver should not be nil")
			}

			if !tc.SkipStartSession {
				_, err := driver.StartSession(context.TODO(), MockStartSessionInput("fakeinstance"))
				if err != nil {
					t.Fatalf("got an error when calling StartSession %v", err)
				}
			}

			args, err := driver.Args()
			if tc.ErrorExpected && err == nil {
				t.Fatalf("Driver.Args with a %q should have failed but instead no error was returned", tc.Name)
			}

			if tc.ErrorExpected {
				return
			}

			if err != nil {
				t.Fatalf("got an error when it should've worked %v", err)
			}

			// validate launch script
			expectedArgString := fmt.Sprintf(`{"SessionId":"packerid","StreamUrl":"http://packer.io","TokenValue":"packer-token"} east StartSession %s {"DocumentName":"AWS-StartPortForwardingSession","Parameters":{"localPortNumber":["8001"],"portNumber":["22"]},"Target":"fakeinstance"} example.com`, tc.ProfileName)
			argString := strings.Join(args, " ")
			if argString != expectedArgString {
				t.Errorf("Expected launch script to be %q but got %q", expectedArgString, argString)
			}

		})
	}
}
