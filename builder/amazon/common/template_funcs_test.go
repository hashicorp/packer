package common

import (
	"testing"
)

func TestAMITemplatePrepare_clean(t *testing.T) {
	origName := "AMZamz09()./-_:&^ $%[]#'@"
	expected := "AMZamz09()./-_--- --[]-'@"

	name := templateCleanAMIName(origName)

	if name != expected {
		t.Fatalf("template names do not match: expected %s got %s\n", expected, name)
	}
}
