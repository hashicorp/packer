package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func testMeta(t *testing.T) Meta {
	return Meta{
		Ui: new(cli.MockUi),
	}
}
