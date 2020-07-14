package api_client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

// LXDClient connects to the LXD server through its go library.
type LXDClient struct {
	server lxd.InstanceServer
}

func NewLXDClient(path string) (*LXDClient, error) {
	c, err := lxd.ConnectLXDUnix("", nil)
	if err != nil {
		return nil, err
	}
	return &LXDClient{server: c}, nil
}

// DeleteImage delets an an LXD image and waits until it is deleted or returns an error if unsuccessful.
func (s *LXDClient) DeleteImage(name string) error {
	op, err := s.server.DeleteImage(name)
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	return nil
}

// LaunchContainer creates and starts an LXD container and waits until it is created and started or returns an error if unsuccessful.
func (s *LXDClient) LaunchContainer(name string, image string, profile string, launchConfig map[string]string) error {
	if len(launchConfig) > 0 {
		return fmt.Errorf("launchConfig is not supported for the API client")
	}
	err := s.CreateContainer(name, image, profile)
	if err != nil {
		return err
	}
	return s.StartContainer(name)
}

// CreateContainer creates an LXD container and waits until it is created or returns an error if unsuccessful.
func (s *LXDClient) CreateContainer(name string, image string, profile string) error {
	req := api.ContainersPost{
		Name: name,
		ContainerPut: api.ContainerPut{
			Profiles:  []string{profile},
			Ephemeral: false,
		},
		Source: api.ContainerSource{
			Type:  "image",
			Alias: image,
			Server: "https://cloud-images.ubuntu.com/daily",
			Protocol: "simplestreams",
		},
	}
	op, err := s.server.CreateContainer(req)
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	return nil
}

// StartContainer starts an LXD container and waits until it is started or returns an error if unsuccessful.
func (s *LXDClient) StartContainer(name string) error {
	req := api.InstanceStatePut{
		Action:  "start",
		Timeout: -1,
	}
	op, err := s.server.UpdateInstanceState(name, req, "")
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	return nil
}

// PublishContainer creates an LXD image from a stopped LXD container and waits until the image is created or
// returns an error if unsuccessful. On successful creating it returns the fingerprint of the image.
func (s *LXDClient) PublishContainer(name string, outImage string, publishProperties map[string]string) (string, error) {
	if len(publishProperties) > 0 {
		return "", fmt.Errorf("publishProperties is not supported for the API client")
	}
	req := api.ImagesPost{
		Source: &api.ImagesPostSource{
			Type: "container",
			Name: name,
		},
		Aliases: []api.ImageAlias{
			{Name: outImage},
		},
	}
	op, err := s.server.CreateImage(req, nil)
	if err != nil {
		return "", err
	}
	if err = op.Wait(); err != nil {
		return "", err
	}
	fingerprint := op.Get().Metadata["fingerprint"].(string)
	return fingerprint, nil
}

// StopContainer stops an LXD container and waits until it is stopped or returns an error if unsuccessful.
func (s *LXDClient) StopContainer(name string) error {
	reqState := api.ContainerStatePut{
		Action: "stop",
		Force:  true,
	}
	op, err := s.server.UpdateContainerState(name, reqState, "")
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	return nil
}

// DeleteContainer deletes an LXD container and waits until it is deleted or returns an error if unsuccessful.
func (s *LXDClient) DeleteContainer(name string) error {
	if err := s.StopContainer(name); err != nil {
		return err
	}
	op, err := s.server.DeleteContainer(name)
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	return nil
}

// ExecuteContainer executes cmd inside the container with name.
func (s *LXDClient) ExecuteContainer(name string, wrapper func(string) (string, error), cmd *packer.RemoteCmd) error {
	stdin := ioutil.NopCloser(bytes.NewReader(nil))
	stdout := os.Stdout

	// Prepare the command
	req := api.InstanceExecPost{
		Command:     strings.Split(cmd.Command, " "),
		WaitForWS:   true,
		Interactive: false,
	}

	execArgs := lxd.InstanceExecArgs{
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   os.Stderr,
		DataDone: make(chan bool),
	}
	op, err := s.server.ExecInstance(name, req, &execArgs)
	if err != nil {
		return err
	}
	if err = op.Wait(); err != nil {
		return err
	}
	opAPI := op.Get()

	if int(opAPI.Metadata["return"].(float64)) == 1 {
		log.Println("some error")
	}
	return nil
}
