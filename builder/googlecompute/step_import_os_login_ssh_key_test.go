package googlecompute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"google.golang.org/api/oauth2/v2"
)

func TestStepImportOSLoginSSHKey_impl(t *testing.T) {
	var _ multistep.Step = new(StepImportOSLoginSSHKey)
}

func TestStepImportOSLoginSSHKey(t *testing.T) {
	tt := []struct {
		Name           string
		UseOSLogin     bool
		ExpectedEmail  string
		ExpectedAction multistep.StepAction
		PubKeyExpected bool
	}{
		{
			Name:           "UseOSLoginDisabled",
			ExpectedAction: multistep.ActionContinue,
		},
		{
			Name:           "UseOSLoginWithAccountFile",
			UseOSLogin:     true,
			ExpectedAction: multistep.ActionContinue,
			ExpectedEmail:  "raffi-compute@developer.gserviceaccount.com",
			PubKeyExpected: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		state := testState(t)
		step := new(StepImportOSLoginSSHKey)
		defer step.Cleanup(state)

		config := state.Get("config").(*Config)
		config.UseOSLogin = tc.UseOSLogin

		if tc.PubKeyExpected {
			config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}
		}

		if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v", action)
		}

		if step.accountEmail != tc.ExpectedEmail {
			t.Fatalf("expected accountEmail to be %q but got %q", tc.ExpectedEmail, step.accountEmail)
		}

		if _, ok := state.GetOk("ssh_key_public_sha256"); !ok && tc.PubKeyExpected {
			t.Fatal("expected to see a public key")
		}
	}
}

func TestStepImportOSLoginSSHKey_withAccountFile(t *testing.T) {
	// default teststate contains an account file
	state := testState(t)
	step := new(StepImportOSLoginSSHKey)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = true
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	fakeAccountEmail := "raffi-compute@developer.gserviceaccount.com"
	if step.accountEmail != fakeAccountEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeAccountEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withNoAccountFile(t *testing.T) {
	state := testState(t)
	fakeAccountEmail := "testing@packer.io"
	step := &StepImportOSLoginSSHKey{
		TokeninfoFunc: func(ctx context.Context) (*oauth2.Tokeninfo, error) {
			return &oauth2.Tokeninfo{Email: fakeAccountEmail}, nil
		},
	}
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.account = nil
	config.UseOSLogin = true
	config.Comm.SSHPublicKey = []byte{'k', 'e', 'y'}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != fakeAccountEmail {
		t.Fatalf("expected accountEmail to be %q but got %q", fakeAccountEmail, step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if !ok {
		t.Fatal("expected to see a public key")
	}

	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	if pubKey != hex.EncodeToString(sha256sum[:]) {
		t.Errorf("expected to see a matching public key, but got %q", pubKey)
	}
}

func TestStepImportOSLoginSSHKey_withPrivateSSHKey(t *testing.T) {
	// default teststate contains an account file
	state := testState(t)
	step := new(StepImportOSLoginSSHKey)
	defer step.Cleanup(state)

	config := state.Get("config").(*Config)
	config.UseOSLogin = true
	config.Comm.SSHPrivateKey = []byte{'k', 'e', 'y'}
	config.Comm.SSHPublicKey = nil

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if step.accountEmail != "" {
		t.Fatalf("expected accountEmail to be unset but got %q", step.accountEmail)
	}

	pubKey, ok := state.GetOk("ssh_key_public_sha256")
	if ok {
		t.Errorf("expected to not see a public key when using a dedicated private key, but got %q", pubKey)
	}
}
