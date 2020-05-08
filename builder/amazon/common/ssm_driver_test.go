package common

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
)

func TestSSMDriver_StartSession(t *testing.T) {
	tt := []struct {
		Name          string
		PluginName    string
		ErrorExpected bool
	}{
		{"NonExistingPlugin", "boguspluginname", true},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			driver := SSMDriver{
				Region:          "region",
				Session:         new(ssm.StartSessionOutput),
				SessionParams:   ssm.StartSessionInput{},
				SessionEndpoint: "endpoint",
				PluginName:      tc.PluginName}

			ctx := context.TODO()
			err := driver.StartSession(ctx)

			if tc.ErrorExpected && err == nil {
				t.Fatalf("Executing %q should have failed but instead no error was returned", tc.PluginName)
			}

		})
	}
}

func TestSSMDriver_Args(t *testing.T) {
	tt := []struct {
		Name          string
		Session       *ssm.StartSessionOutput
		ProfileName   string
		ErrorExpected bool
	}{
		{
			Name:          "NilSession",
			ErrorExpected: true,
		},
		{
			Name:          "NonNilSession",
			Session:       new(ssm.StartSessionOutput),
			ErrorExpected: false,
		},
		{
			Name:          "SessionWithProfileName",
			Session:       new(ssm.StartSessionOutput),
			ProfileName:   "default",
			ErrorExpected: false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			driver := SSMDriver{
				Region:          "region",
				ProfileName:     tc.ProfileName,
				Session:         tc.Session,
				SessionParams:   ssm.StartSessionInput{},
				SessionEndpoint: "amazon.com/sessions",
			}

			args, err := driver.Args()
			if tc.ErrorExpected && err == nil {
				t.Fatalf("SSMDriver.Args with a %q should have failed but instead no error was returned", tc.Name)
			}

			if tc.ErrorExpected {
				return
			}

			// validate launch script
			expectedArgString := fmt.Sprintf(`{"SessionId":null,"StreamUrl":null,"TokenValue":null} %s StartSession %s {"DocumentName":null,"Parameters":null,"Target":null} %s`, driver.Region, driver.ProfileName, driver.SessionEndpoint)
			argString := strings.Join(args, " ")
			if argString != expectedArgString {
				t.Errorf("Expected launch script to be %q but got %q", expectedArgString, argString)
			}

		})
	}
}
