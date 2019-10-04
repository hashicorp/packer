package chroot

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestBuilder_Prepare_DiskAsInput(t *testing.T) {
	b := Builder{}
	_, err := b.Prepare(map[string]interface{}{
		"source": "/subscriptions/28279221-ccbe-40f0-b70b-4d78ab822e09/resourceGroups/testrg/providers/Microsoft.Compute/disks/diskname",
	})

	if err != nil {
		// make sure there is no error about the source field
		errs, ok := err.(*packer.MultiError)
		if !ok {
			t.Error("Expected the returned error to be of type packer.MultiError")
		}
		for _, err := range errs.Errors {
			if matched, _ := regexp.MatchString(`(^|\W)source\W`, err.Error()); matched {
				t.Errorf("Did not expect an error about the 'source' field, but found %q", err)
			}
		}
	}
}

func TestBuilder_Prepare(t *testing.T) {
	type config map[string]interface{}
	
	tests := []struct {
		name     string
		config   config
		want     []string
		validate func(Config)
		wantErr  bool
	}{
		{
			name: "HappyPath",
			config: config{
				"client_id":         "123",
				"client_secret":     "456",
				"subscription_id":   "789",
				"resource_group":    "rgname",
				"image_resource_id": "/subscriptions/789/resourceGroups/otherrgname/providers/Microsoft.Compute/images/MyDebianOSImage-{{timestamp}}",
				"source":            "credativ:Debian:9:latest",
			},
			wantErr: false,
			validate: func(c Config){
				if(c.OSDiskSizeGB!=0){
					t.Fatalf("Expected OSDiskSizeGB to be 0, was %+v", c.OSDiskSizeGB)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{}

			got, err := b.Prepare(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Builder.Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.Prepare() = %v, want %v", got, tt.want)
			}
		})
	}
}
