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
	ID            string
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
		//i.Fingerprint = "00ee249320213a1e20578a551c11f47bbdd94ea4"
	}

	return &i
}
