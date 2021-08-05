package packer_registry

import (
	"errors"
	"fmt"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
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
func NewIteration(opts IterationOptions) (*Iteration, error) {
	i := Iteration{
		expectedBuilds: make([]string, 0),
	}

	// By default we try to load a Fingerprint from the environment variable.
	// If no variable is defined we should try to load a fingerprint from Git, or other VCS.
	i.Fingerprint = os.Getenv("HCP_PACKER_BUILD_FINGERPRINT")

	// get a Git SHA
	if i.Fingerprint != "" {
		return &i, nil
	}

	fp, err := GetGitFingerprint(opts)
	if err != nil {
		return nil, err
	}
	i.Fingerprint = fp
	return &i, nil
}

// GetGitFingerprint returns the HEAD commit for some template dir defined in opt.TemplateBaseDir.
// If the base directory is not under version control an error is returned.
func GetGitFingerprint(opts IterationOptions) (string, error) {
	r, err := git.PlainOpenWithOptions(opts.TemplateBaseDir, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return "", fmt.Errorf("Packer was unable to load a git sha. "+
			"If your Packer template is not in a git repo, please add a unique "+
			"template fingerprint using the env var HCP_PACKER_BUILD_FINGERPRINT. "+
			"Error: %s", err)
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
		return "", fmt.Errorf("Packer encountered an issue reading the git info for the path %q.\n"+
			"If your Packer template is not in a git repo, please add a unique "+
			"template fingerprint using the env var HCP_PACKER_BUILD_FINGERPRINT. "+
			"Error: %s", opts.TemplateBaseDir, err)
	}

	// log.Printf("Author: %v, Commit: %v\n", c.User.Email, ref.Hash())

	return ref.Hash().String(), nil
}

// AddImageToBuild appends one or more images artifacts to the build referred to by buildName.
func (i *Iteration) AddImageToBuild(buildName string, images ...Image) error {
	existingBuild, ok := i.builds.Load(buildName)
	if !ok {
		return errors.New("no build found for the name " + buildName)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", buildName)
	}

	if build.Images == nil {
		build.Images = make(map[string]Image)
	}

	for _, image := range images {
		if build.CloudProvider == "" {
			build.CloudProvider = image.ProviderName
		}

		k := fmt.Sprintf("%s.region:%s", buildName, image.ProviderRegion)
		build.Images[k] = image
	}

	i.builds.Store(buildName, build)
	return nil
}

// AddLabelsToBuild merges the contents of data to the labels associated with the build referred to by buildName.
func (i *Iteration) AddLabelsToBuild(buildName string, data map[string]string) error {
	existingBuild, ok := i.builds.Load(buildName)
	if !ok {
		return errors.New("no associated build found for the name " + buildName)
	}

	build, ok := existingBuild.(*Build)
	if !ok {
		return fmt.Errorf("the build for the component %q does not appear to be a valid registry Build", buildName)
	}

	for k, v := range data {
		if _, ok := build.Labels[k]; ok {
			continue
		}
		build.Labels[k] = v
	}

	i.builds.Store(buildName, build)

	return nil
}
