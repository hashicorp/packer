package proxmoxiso

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type uploaderMock struct {
	fail      bool
	wasCalled bool
}

func (m *uploaderMock) Upload(node string, storage string, contentType string, filename string, file io.Reader) error {
	m.wasCalled = true
	if m.fail {
		return fmt.Errorf("Testing induced failure")
	}
	return nil
}

var _ uploader = &uploaderMock{}

func TestUploadISO(t *testing.T) {
	cs := []struct {
		name          string
		builderConfig *Config
		downloadPath  string
		failUpload    bool

		expectError        bool
		expectUploadCalled bool
		expectedISOPath    string
		expectedAction     multistep.StepAction
	}{
		{
			name:          "should not call upload unless configured to do so",
			builderConfig: &Config{shouldUploadISO: false, ISOFile: "local:iso/some-file"},

			expectUploadCalled: false,
			expectedISOPath:    "local:iso/some-file",
			expectedAction:     multistep.ActionContinue,
		},
		{
			name: "success should continue",
			builderConfig: &Config{
				shouldUploadISO: true,
				ISOStoragePool:  "local",
				ISOConfig:       commonsteps.ISOConfig{ISOUrls: []string{"http://server.example/some-file.iso"}},
			},
			downloadPath: "testdata/test.iso",

			expectedISOPath:    "local:iso/some-file.iso",
			expectUploadCalled: true,
			expectedAction:     multistep.ActionContinue,
		},
		{
			name: "failing upload should halt",
			builderConfig: &Config{
				shouldUploadISO: true,
				ISOStoragePool:  "local",
				ISOConfig:       commonsteps.ISOConfig{ISOUrls: []string{"http://server.example/some-file.iso"}},
			},
			downloadPath: "testdata/test.iso",
			failUpload:   true,

			expectError:        true,
			expectUploadCalled: true,
			expectedAction:     multistep.ActionHalt,
		},
		{
			name: "downloader: state misconfiguration should halt",
			builderConfig: &Config{
				shouldUploadISO: true,
				ISOStoragePool:  "local",
				ISOConfig:       commonsteps.ISOConfig{ISOUrls: []string{"http://server.example/some-file.iso"}},
			},

			expectError:        true,
			expectUploadCalled: false,
			expectedAction:     multistep.ActionHalt,
		},
		{
			name: "downloader: file unreadable should halt",
			builderConfig: &Config{
				shouldUploadISO: true,
				ISOStoragePool:  "local",
				ISOConfig:       commonsteps.ISOConfig{ISOUrls: []string{"http://server.example/some-file.iso"}},
			},
			downloadPath: "testdata/non-existent.iso",

			expectError:        true,
			expectUploadCalled: false,
			expectedAction:     multistep.ActionHalt,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			m := &uploaderMock{fail: c.failUpload}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packersdk.TestUi(t))
			state.Put("iso-config", c.builderConfig)
			state.Put(downloadPathKey, c.downloadPath)
			state.Put("proxmoxClient", m)

			step := stepUploadISO{}
			action := step.Run(context.TODO(), state)
			step.Cleanup(state)

			if action != c.expectedAction {
				t.Errorf("Expected action to be %v, got %v", c.expectedAction, action)
			}
			if m.wasCalled != c.expectUploadCalled {
				t.Errorf("Expected mock to be called: %v, got: %v", c.expectUploadCalled, m.wasCalled)
			}
			err, gotError := state.GetOk("error")
			if gotError != c.expectError {
				t.Errorf("Expected error state to be: %v, got: %v", c.expectError, gotError)
			}
			if err == nil {
				if isoPath := state.Get("iso_file"); isoPath != c.expectedISOPath {
					if _, ok := isoPath.(string); !ok {
						isoPath = ""
					}
					t.Errorf("Expected state iso_path to be %q, got %q", c.expectedISOPath, isoPath)
				}
			}
		})
	}
}
