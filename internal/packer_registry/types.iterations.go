package packer_registry

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	git "github.com/go-git/go-git/v5"
	// "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
)

type Iteration struct {
	ID           string
	AncestorSlug string
	Fingerprint  string
	RunUUID      string
	Labels       map[string]string
	builds       Builds
}

type IterationOptions struct {
	UseGitBackend bool
}

func GetGitFingerprint() (string, error) {
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("Error loading git sha", err)
	}
	// The config can be used to retrieve user identity. for example,
	// c.User.Email. Leaving in but commented because I'm not sure we care
	// about this identity right now. - Megan
	//
	// c, err := r.ConfigScoped(config.GlobalScope)
	// if err != nil {
	// 	return "", fmt.Errorf("Error setting git scope", err)
	// }
	ref, _ := r.Head()
	// log.Printf("Author: %v, Commit: %v\n", c.User.Email, ref.Hash())
	return ref.Hash().String(), nil
}

// NewIteration returns a pointer to an Iteration that can be used for storing Packer build details needed by PAR.
func NewIteration(opts IterationOptions) (*Iteration, error) {
	i := Iteration{
		builds: NewBuilds(),
	}

	// By default we try to load a Fingerprint from the environment variable.
	// If no variable is defined we should try to load a fingerprint from Git, or other VCS.
	i.Fingerprint = os.Getenv("HCP_PACKER_BUILD_FINGERPRINT")

	// Simulating a Git SHA
	if i.Fingerprint == "" {
		s := []byte(time.Now().String())
		// TODO allow user to set fingerprint through Packer block or
		// environment variable?
		i.Fingerprint = fmt.Sprintf("%x", sha1.Sum(s))
		//i.Fingerprint = "00ee249320213a1e20578a551c11f47bbdd94ea4"
	} else {
		fp, err := GetGitFingerprint()
		if err != nil {
			return nil, err
		}
		i.Fingerprint = fp
	}

	return &i, nil
}
