package ebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
)

type artifact struct {
	// A map of regions to AMI IDs.
	amis map[string]string

	// EC2 connection for performing API stuff.
	conn *ec2.EC2
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (*artifact) Files() []string {
	// We have no files
	return nil
}

func (a *artifact) Id() string {
	parts := make([]string, 0, len(a.amis))
	for region, amiId := range a.amis {
		parts = append(parts, fmt.Sprintf("%s:%s", region, amiId))
	}

	return strings.Join(parts, ",")
}

func (a *artifact) String() string {
	amiStrings := make([]string, 0, len(a.amis))
	for region, id := range a.amis {
		single := fmt.Sprintf("%s: %s", region, id)
		amiStrings = append(amiStrings, single)
	}

	return fmt.Sprintf("AMIs were created:\n\n%s", strings.Join(amiStrings, "\n"))
}

func (a *artifact) Destroy() error {
	errors := make([]error, 0)

	for _, imageId := range a.amis {
		log.Printf("Deregistering image ID: %s", imageId)
		if _, err := a.conn.DeregisterImage(imageId); err != nil {
			errors = append(errors, err)
		}

		// TODO(mitchellh): Delete the snapshots associated with an AMI too
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packer.MultiError{errors}
		}
	}

	return nil
}
