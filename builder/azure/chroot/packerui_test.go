package chroot

import (
	"io/ioutil"
	"strings"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// testUI returns a test ui plus a function to retrieve the errors written to the ui
func testUI() (packersdk.Ui, func() string) {
	errorBuffer := &strings.Builder{}
	ui := &packer.BasicUi{
		Reader:      strings.NewReader(""),
		Writer:      ioutil.Discard,
		ErrorWriter: errorBuffer,
	}
	return ui, errorBuffer.String
}
