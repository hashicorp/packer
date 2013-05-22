package amazonebs

import (
	"fmt"
	"strings"
)

type artifact struct {
	// A map of regions to AMI IDs.
	amis map[string]string
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (*artifact) Files() []string {
	// We have no files
	return nil
}

func (*artifact) Id() string {
	// TODO(mitchellh): Id
	return "TODO"
}

func (a *artifact) String() string {
	amiStrings := make([]string, 0, len(a.amis))
	for region, id := range a.amis {
		single := fmt.Sprintf("%s: %s", region, id)
		amiStrings = append(amiStrings, single)
	}

	return fmt.Sprintf("AMIs were created:\n\n%s", strings.Join(amiStrings, "\n"))
}
