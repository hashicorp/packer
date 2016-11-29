package docker

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	godocker "github.com/fsouza/go-dockerclient"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type DockerApiDriver struct {
	Ui     packer.Ui
	Ctx    *interpolate.Context
	Config DockerHostConfig

	client        *godocker.Client
	auth          godocker.AuthConfiguration
	identityToken string
}

func (d *DockerApiDriver) DeleteImage(id string) error {
	log.Printf("Deleting image: %s", id)
	return d.client.RemoveImage(id)
}

func (d *DockerApiDriver) Commit(id, author string, changes []string, message string) (string, error) {

	// TODO changes
	config := godocker.Config{}

	image, err := d.client.CommitContainer(godocker.CommitContainerOptions{
		Container: id,
		Message:   message,
		Author:    author,
		Run:       &config,
	})
	if err != nil {
		return "", err
	}
	return image.ID, nil
}

func (d *DockerApiDriver) Export(id string, dst io.Writer) error {
	log.Printf("Exporting container: %s", id)

	return d.client.ExportContainer(godocker.ExportContainerOptions{
		ID:           id,
		OutputStream: dst,
	})
}

func (d *DockerApiDriver) Import(path string, repo string) (string, error) {
	// There should be only one artifact of the Docker builder
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	err = d.client.ImportImage(godocker.ImportImageOptions{
		Repository:  repo,
		Source:      path,
		InputStream: file,
		// TODO Fix !?
	})

	if err != nil {
		return "", err
	}
	//log.Printf("Imported: %v\n", resp) // TODO parse the respones
	return "", nil // TODO
}

func (d *DockerApiDriver) IPAddress(id string) (string, error) {

	resp, err := d.client.InspectContainer(id)
	return resp.NetworkSettings.IPAddress, err
}

func (d *DockerApiDriver) Login(repo, email, user, pass string) error {
	auth := godocker.AuthConfiguration{
		ServerAddress: repo,
		Email:         email,
		Username:      user,
		Password:      pass,
	}
	status, err := d.client.AuthCheck(&auth)

	if err != nil {
		return err
	}

	d.auth = auth
	d.identityToken = status.IdentityToken
	log.Printf("auth: %v\ntoken: %s\nstatus: %v", d.auth, d.identityToken, status) // TODO DEBUG
	return nil
}

func (d *DockerApiDriver) Logout(repo string) error {
	d.identityToken = ""
	return nil
}

// TODO split imageTag -> image, tag
func (d *DockerApiDriver) Pull(imageTag string) error {

	tmp := strings.Split(imageTag, ":")
	image := tmp[0]
	tag := "latest"
	if len(tmp) > 1 {
		tag = tmp[1]
	}

	var output bytes.Buffer
	opts := godocker.PullImageOptions{
		Repository:   image,
		Tag:          tag,
		OutputStream: &output,
	}
	err := d.client.PullImage(opts, d.auth)
	if err != nil {
		return err
	}
	return readAndStream(&output, d.Ui)
}

func (d *DockerApiDriver) Push(name string) error {

	var output bytes.Buffer
	opts := godocker.PushImageOptions{
		Name: name,
		// Tag:  "latest",
		// Registry: "",
		OutputStream: &output,
	}
	err := d.client.PushImage(opts, d.auth)
	if err != nil {
		return err
	}
	return readAndStream(&output, d.Ui)
}

func (d *DockerApiDriver) SaveImage(id string, dst io.Writer) error {

	opts := godocker.ExportImageOptions{
		Name:         id,
		OutputStream: dst,
	}
	log.Printf("Exporting image: %s", id)
	err := d.client.ExportImage(opts)
	return err
}

func (d *DockerApiDriver) StartContainer(config *ContainerConfig) (string, error) {

	// for host, guest := range config.Volumes {
	// 	args = append(args, "-v", fmt.Sprintf("%s:%s", host, guest))
	// }

	conf := godocker.Config{
		AttachStdout: false,
		Tty:          true,
		Env:          []string{}, // TODO
		Cmd:          config.RunCommand,
		Image:        config.Image,
		//Volumes:      config.Volumes, // TODO
	}
	hostCfg := godocker.HostConfig{
		Privileged: config.Privileged,
	}
	network := godocker.NetworkingConfig{}

	opts := godocker.CreateContainerOptions{
		Config:           &conf,
		HostConfig:       &hostCfg,
		NetworkingConfig: &network,
	}

	d.Ui.Message("Creating container")
	//  -d -i -t {{.Image}} /bin/bash
	container, err := d.client.CreateContainer(opts)
	if err != nil {
		return "", err
	}
	log.Printf("Created container: %s", container.ID)
	// for warning := range resp.Warnings {
	// 	log.Printf("Warning: %s\n", warning)
	// }

	d.Ui.Message("Starting container")
	err = d.client.StartContainer(container.ID, nil)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func (d *DockerApiDriver) StopContainer(id string) error {

	err := d.client.KillContainer(godocker.KillContainerOptions{
		ID:     id,
		Signal: godocker.SIGKILL,
	})
	if err != nil {
		return err
	}

	return d.client.RemoveContainer(godocker.RemoveContainerOptions{ID: id})
}

func (d *DockerApiDriver) TagImage(id string, repo string, force bool) error {
	return d.client.TagImage(id, godocker.TagImageOptions{
		Repo:  repo,
		Force: force,
		// Tag: "",
	})
}

func (d *DockerApiDriver) Verify() error {
	d.Ui.Say("Warning! You are running the EXPERMINATAL Docker API driver!")

	var err error
	var client *godocker.Client

	if d.client == nil {
		if d.Config.Host == "" {
			log.Println("Using Docker Host settings from environment variables.")
			client, err = godocker.NewClientFromEnv()
		} else {
			if *d.Config.TlsVerify {
				log.Printf("Using Docker Host: %s with verified TLS.", d.Config.Host)
				client, err = godocker.NewTLSClient(d.Config.Host,
					filepath.Join(d.Config.CertPath, "cert.pem"),
					filepath.Join(d.Config.CertPath, "key.pem"),
					filepath.Join(d.Config.CertPath, "ca.pem"))
			} else {
				log.Printf("Using Docker Host: %s", d.Config.Host)
				client, err = godocker.NewClient(d.Config.Host)
			}
		}
		d.client = client
	}

	log.Printf("Docker: %+v", d.client)
	return err
}

func (d *DockerApiDriver) Version() (*version.Version, error) {
	if d.client == nil {
		return nil, fmt.Errorf("No client %+v", d)
	}
	env, err := d.client.Version()
	if err != nil {
		return nil, err
	}
	return version.NewVersion(env.Get("Version"))
}
