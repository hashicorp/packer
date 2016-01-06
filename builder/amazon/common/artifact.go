package common

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/packer/packer"
)

// Artifact is an artifact implementation that contains built AMIs.
type Artifact struct {
	// A map of regions to AMI IDs.
	Amis map[string]string

	// BuilderId is the unique ID for the builder that created this AMI
	BuilderIdValue string

	// EC2 connection for performing API stuff.
	Conn *ec2.EC2
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.Amis))
	for region, amiId := range a.Amis {
		parts = append(parts, fmt.Sprintf("%s:%s", region, amiId))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	amiStrings := make([]string, 0, len(a.Amis))
	for region, id := range a.Amis {
		single := fmt.Sprintf("%s: %s", region, id)
		amiStrings = append(amiStrings, single)
	}

	sort.Strings(amiStrings)
	return fmt.Sprintf("AMIs were created:\n\n%s", strings.Join(amiStrings, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	for region, imageId := range a.Amis {
		log.Printf("Deregistering image ID (%s) from region (%s)", imageId, region)

		regionConfig := &aws.Config{
			Credentials: a.Conn.Config.Credentials,
			Region:      aws.String(region),
		}
		sess := session.New(regionConfig)
		regionConn := ec2.New(sess)

		input := &ec2.DeregisterImageInput{
			ImageId: &imageId,
		}
		if _, err := regionConn.DeregisterImage(input); err != nil {
			errors = append(errors, err)
		}

		// TODO(mitchellh): Delete the snapshots associated with an AMI too
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packer.MultiError{Errors: errors}
		}
	}

	return nil
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.Amis {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
