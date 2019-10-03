package chroot

import (
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
