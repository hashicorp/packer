package docker

import (
	"fmt"
	"log"
	"os/exec"
)

type Artifact struct {
	Repository string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return make([]string, 0)
}

func (a *Artifact) Id() string {
	return a.Repository
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Image ID: %s", a.Repository)
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d", a.Id())

	cmd := exec.Command("docker", "rmi", a.Id())
	return cmd.Run()
}
