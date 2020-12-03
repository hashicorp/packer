package ebsvolume

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// map of region to list of volume IDs
type EbsVolumes map[string][]string

// Artifact is an artifact implementation that contains built AMIs.
type Artifact struct {
	// A map of regions to EBS Volume IDs.
	Volumes EbsVolumes

	// BuilderId is the unique ID for the builder that created this AMI
	BuilderIdValue string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}

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

// returns a sorted list of region:ID pairs
func (a *Artifact) idList() []string {
	parts := make([]string, 0, len(a.Volumes))
	for region, volumeIDs := range a.Volumes {
		for _, volumeID := range volumeIDs {
			parts = append(parts, fmt.Sprintf("%s:%s", region, volumeID))
		}
	}

	sort.Strings(parts)
	return parts
}

func (a *Artifact) Id() string {
	return strings.Join(a.idList(), ",")
}

func (a *Artifact) String() string {
	return fmt.Sprintf("EBS Volumes were created:\n\n%s", strings.Join(a.idList(), "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}
	return nil
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	for region, volumeIDs := range a.Volumes {
		for _, volumeID := range volumeIDs {
			log.Printf("Deregistering Volume ID (%s) from region (%s)", volumeID, region)

			input := &ec2.DeleteVolumeInput{
				VolumeId: &volumeID,
			}
			if _, err := a.Conn.DeleteVolume(input); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packersdk.MultiError{Errors: errors}
		}
	}

	return nil
}
