package packer_registry

import (
	"crypto/sha1"
	"fmt"
	"os"
	"sync"
	"time"
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
	UseGitBackend bool
}

// NewIteration returns a pointer to an Iteration that can be used for storing Packer build details needed by PAR.
func NewIteration(opts IterationOptions) *Iteration {
	i := Iteration{
		builds:         sync.Map{},
		expectedBuilds: make([]string, 0),
	}

	// By default we try to load a Fingerprint from the environment variable.
	// If no variable is defined we should try to load a fingerprint from Git, or other VCS.
	i.Fingerprint = os.Getenv("HCP_PACKER_BUILD_FINGERPRINT")

	// Simulating a Git SHA
	if i.Fingerprint == "" {
		s := []byte(time.Now().String())
		i.Fingerprint = fmt.Sprintf("%x", sha1.Sum(s))
	}

	return &i
}
