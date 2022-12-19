package registry

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/oklog/ulid"
)

type Iteration struct {
	ID             string
	AncestorSlug   string
	Fingerprint    string
	RunUUID        string
	builds         sync.Map
	expectedBuilds []string
}

type IterationOptions struct {
	TemplateBaseDir string
}

// NewIteration returns a pointer to an Iteration that can be used for storing Packer build details needed by PAR.
func NewIteration() *Iteration {
	i := Iteration{
		expectedBuilds: make([]string, 0),
	}

	return &i
}

// Initialize prepares the iteration to be used with an active HCP Packer registry bucket.
func (i *Iteration) Initialize() error {
	if i == nil {
		return errors.New("Unexpected call to initialize for a nil Iteration")
	}

	// By default we try to load a Fingerprint from the environment variable.
	// If no variable is defined we generate a new fingerprint.
	i.Fingerprint = os.Getenv(env.HCPPackerBuildFingerprint)

	if i.Fingerprint != "" {
		return nil
	}

	fp, err := ulid.New(ulid.Now(), ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0))
	if err != nil {
		return fmt.Errorf("Failed to generate a fingerprint: %s", err)
	}
	i.Fingerprint = fp.String()

	return nil
}

//StoreBuild stores a build for buildName to an active iteration.
func (i *Iteration) StoreBuild(buildName string, build *Build) {
	i.builds.Store(buildName, build)
}

//Build gets the store build associated with buildName in the active iteration.
func (i *Iteration) Build(buildName string) (*Build, error) {
	build, ok := i.builds.Load(buildName)
	if !ok {
		return nil, errors.New("no associated build found for the name " + buildName)
	}

	b, ok := build.(*Build)
	if !ok {
		return nil, fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", buildName)
	}

	return b, nil
}

//HasBuild checks if iteration has a stored build associated with buildName.
func (i *Iteration) HasBuild(buildName string) bool {
	_, ok := i.builds.Load(buildName)

	return ok
}

// AddImageToBuild appends one or more images artifacts to the build referred to by buildName.
func (i *Iteration) AddImageToBuild(buildName string, images ...registryimage.Image) error {
	build, err := i.Build(buildName)
	if err != nil {
		return fmt.Errorf("AddImageToBuild: %w", err)
	}

	err = build.AddImages(images...)
	if err != nil {
		return fmt.Errorf("AddImageToBuild: %w", err)
	}

	i.StoreBuild(buildName, build)
	return nil
}

// AddLabelsToBuild merges the contents of data to the labels associated with the build referred to by buildName.
func (i *Iteration) AddLabelsToBuild(buildName string, data map[string]string) error {
	build, err := i.Build(buildName)
	if err != nil {
		return fmt.Errorf("AddLabelsToBuild: %w", err)
	}

	build.MergeLabels(data)

	i.StoreBuild(buildName, build)
	return nil
}

// AddSHAToBuildLabels adds the Git SHA for the current iteration (if set) as a label for all the builds of the iteration
func (i *Iteration) AddSHAToBuildLabels(sha string) {
	i.builds.Range(func(_, v any) bool {
		b, ok := v.(*Build)
		if !ok {
			return true
		}

		b.MergeLabels(map[string]string{
			"git_sha": sha,
		})

		return true
	})
}

// RemainingBuilds returns the list of builds that are not in a DONE status
func (i *Iteration) RemainingBuilds() []*Build {
	var todo []*Build

	i.builds.Range(func(k, v any) bool {
		build, ok := v.(*Build)
		if !ok {
			// Unlikely since the builds map contains only Build instances
			return true
		}

		if build.Status != models.HashicorpCloudPackerBuildStatusDONE {
			todo = append(todo, build)
		}
		return true
	})

	return todo
}

func (i *Iteration) iterationStatusSummary(ui sdkpacker.Ui) {
	rem := i.RemainingBuilds()
	if rem == nil {
		return
	}

	buf := &strings.Builder{}

	buf.WriteString(fmt.Sprintf(
		"\nIteration %q is not complete, the following builds are not done:\n\n",
		i.Fingerprint))
	for _, b := range rem {
		buf.WriteString(fmt.Sprintf("* %q: %s\n", b.ComponentType, b.Status))
	}
	buf.WriteString("\nYou may resume work on this iteration in further Packer builds by defining the following variable in your environment:\n")
	buf.WriteString(fmt.Sprintf("HCP_PACKER_BUILD_FINGERPRINT=%q", i.Fingerprint))

	ui.Say(buf.String())
}
