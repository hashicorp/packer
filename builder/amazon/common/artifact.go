package common

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Artifact is an artifact implementation that contains built AMIs.
type Artifact struct {
	// A map of regions to AMI IDs.
	Amis map[string]string

	// BuilderId is the unique ID for the builder that created this AMI
	BuilderIdValue string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}

	// EC2 connection for performing API stuff.
	Session *session.Session
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
	return fmt.Sprintf("AMIs were created:\n%s\n", strings.Join(amiStrings, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

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

		regionConn := ec2.New(a.Session, &aws.Config{
			Region: aws.String(region),
		})

		// Get image metadata
		imageResp, err := regionConn.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: []*string{&imageId},
		})
		if err != nil {
			errors = append(errors, err)
		}
		if len(imageResp.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for AMI (%s), no images found", imageId)
			errors = append(errors, err)
		}

		err = DestroyAMIs([]*string{&imageId}, regionConn)
		if err != nil {
			errors = append(errors, err)
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

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.Amis {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
