package registry

import (
	"errors"
	"fmt"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer/internal/hcp/env"
)

type Iteration struct {
	ID             string
	AncestorSlug   string
	Fingerprint    string
	RunUUID        string
	Labels         map[string]string
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
func (i *Iteration) Initialize(opts IterationOptions) error {
	if i == nil {
		return errors.New("Unexpected call to initialize for a nil Iteration")
	}

	// By default we try to load a Fingerprint from the environment variable.
	// If no variable is defined we should try to load a fingerprint from Git, or other VCS.
	i.Fingerprint = os.Getenv(env.HCPPackerBuildFingerprint)

	if i.Fingerprint != "" {
		return nil
	}

	fp, err := GetGitFingerprint(opts)
	if err != nil {
		return err
	}
	i.Fingerprint = fp

	return nil
}

// GetGitFingerprint returns the HEAD commit for some template dir defined in opt.TemplateBaseDir.
// If the base directory is not under version control an error is returned.
func GetGitFingerprint(opts IterationOptions) (string, error) {
	r, err := git.PlainOpenWithOptions(opts.TemplateBaseDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return "", fmt.Errorf("Packer could not read the fingerprint from git." +
			"\n\nIf your Packer template is not within a git managed directory, " +
			"you can set the HCP_PACKER_BUILD_FINGERPRINT environment variable. " +
			"The fingerprint must be less than 32 characters and can contain letters and numbers.")
	}

	// The config can be used to retrieve user identity. for example,
	// c.User.Email. Leaving in but commented because I'm not sure we care
	// about this identity right now. - Megan
	//
	// c, err := r.ConfigScoped(config.GlobalScope)
	// if err != nil {
	//      return "", fmt.Errorf("Error setting git scope", err)
	// }
	ref, err := r.Head()
	if err != nil {
		// If we get there, we're in a Git dir, but HEAD cannot be read.
		//
		// This may happen when there's no commit in the git dir.
		return "", fmt.Errorf("Packer could not read a git SHA in directory %q.\n"+
			"This may happen if your template is in a git repository without any "+
			"commits. You can either add a commit in this directory, or set the "+
			"HCP_PACKER_BUILD_FINGERPRINT environment variable. The fingerprint "+
			"must be less than 32 characters and can contain letters and numbers.\n"+
			"Error: %s", opts.TemplateBaseDir, err)
	}

	// log.Printf("Author: %v, Commit: %v\n", c.User.Email, ref.Hash())

	return ref.Hash().String(), nil
}

// StoreBuild stores a build for buildName to an active iteration.
func (i *Iteration) StoreBuild(buildName string, build *Build) {
	i.builds.Store(buildName, build)
}

// Build gets the store build associated with buildName in the active iteration.
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

// HasBuild checks if iteration has a stored build associated with buildName.
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
