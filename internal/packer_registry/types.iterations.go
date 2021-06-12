package packer_registry

import (
	"crypto/sha1"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
)

type Builds struct {
	sync.RWMutex
	m map[string]*Build
}

type Build struct {
	ComponentType string
	RunUUID       string
	Metadata      map[string]string
	PARtifacts    BuildPARtifacts
	Status        models.HashicorpCloudPackerBuildStatus
}

func NewBuilds() Builds {
	return Builds{
		m: make(map[string]*Build),
	}
}

type PARtifact struct {
	ID                           string
	ProviderName, ProviderRegion string
}

type BuildPARtifacts struct {
	sync.RWMutex
	m map[string][]PARtifact
}

func NewBuildPARtifacts() BuildPARtifacts {
	return BuildPARtifacts{
		m: make(map[string][]PARtifact),
	}
}

type Iteration struct {
	ID           string
	Author       string
	AncestorSlug string
	Fingerprint  string
	RunUUID      string
	Labels       map[string]string
	Builds       Builds
	sync.RWMutex
}

type IterationOptions struct {
	UseGitBackend bool
}

func NewIteration(opts IterationOptions) *Iteration {
	i := Iteration{
		Builds: NewBuilds(),
	}

	if !opts.UseGitBackend {
		i.Author = os.Getenv("USER")
		s := []byte(time.Now().String())
		i.Fingerprint = fmt.Sprintf("%x", sha1.Sum(s))
	}

	return &i
}

/*
func (i *Iteration) UpdateBuild(ctx context.Context, name string) error {
	buildInput := &models.HashicorpCloudPackerCreateBuildRequest{
		BucketSlug:  i.Bucket.Slug,
		Fingerprint: i.Fingerprint,
		Build: &models.HashicorpCloudPackerBuild{
			ComponentType: name,
			IterationID:   i.ID,
			Status:        models.NewHashicorpCloudPackerBuildStatus(models.HashicorpCloudPackerBuildStatusRUNNING),
		},
	}

	err := UpsertBuild(ctx, i.client, buildInput)

	return err
}
*/
