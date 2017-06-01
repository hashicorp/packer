package docker

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"golang.org/x/net/context"
)

type DockerApiDriver struct {
	Ui  packer.Ui
	Ctx *interpolate.Context

	l         sync.Mutex
	client    *client.Client
	repoToken string
}

func dockerApiDriverInit(ctx *interpolate.Context, ui packer.Ui) DockerApiDriver {

	// TODO Allow specefying DOCKER_
	client, _ := client.NewEnvClient()

	return DockerApiDriver{
		Ui:     ui,
		Ctx:    ctx,
		client: client,
	}
}

func (d DockerApiDriver) DeleteImage(id string) error {

	log.Printf("Deleting image: %s", id)
	_, err := d.client.ImageRemove(context.Background(), id, types.ImageRemoveOptions{})

	return err
}

func (d DockerApiDriver) Commit(id string) (string, error) {

	resp, err := d.client.ContainerCommit(context.Background(), id, types.ContainerCommitOptions{})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (d DockerApiDriver) Export(id string, dst io.Writer) error {

	log.Printf("Exporting container: %s", id)
	reader, err := d.client.ContainerExport(context.Background(), id)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, reader)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApiDriver) Import(path string, repo string) (string, error) {
	// There should be only one artifact of the Docker builder
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	source := types.ImageImportSource{
		Source: file,
	}
	resp, err := d.client.ImageImport(context.Background(), source, repo, types.ImageImportOptions{})
	if err != nil {
		return "", err
	}
	log.Printf("Imported: %v\n", resp) // TODO parse the respones
	return "", nil                     // TODO
}

func (d DockerApiDriver) IPAddress(id string) (string, error) {

	resp, err := d.client.ContainerInspect(context.Background(), id)
	return resp.NetworkSettings.IPAddress, err
}

func (d DockerApiDriver) Login(repo, email, user, pass string) error {
	auth := types.AuthConfig{
		ServerAddress: repo,
		Email:         email,
		Username:      user,
		Password:      pass,
	}
	resp, err := d.client.RegistryLogin(context.Background(), auth)
	d.repoToken = resp.IdentityToken

	return err
}

func (d DockerApiDriver) Logout(repo string) error {
	d.repoToken = ""
	return nil
}

func (d DockerApiDriver) Pull(image string) error {
	// TODO handle login

	//tmp := func() (string, error) {
	//	return "", nil
	//}

	opts := types.ImagePullOptions{
		All: false,
		//RegistryAuth:  d.repoToken,
		//PrivilegeFunc: nil,
	}
	reader, err := d.client.ImagePull(context.Background(), image, opts)
	defer reader.Close()
	if err != nil {
		return err
	}
	return readAndStream(reader, d.Ui)
}

func (d DockerApiDriver) Push(name string) error {

	reader, err := d.client.ImagePush(context.Background(), name, types.ImagePushOptions{})
	if err != nil {
		return err
	}
	return readAndStream(reader, d.Ui)
}

func (d DockerApiDriver) SaveImage(id string, dst io.Writer) error {

	log.Printf("Exporting image: %s", id)
	_, err := d.client.ImageSave(context.Background(), []string{id})
	return err
}

func (d DockerApiDriver) StartContainer(config *ContainerConfig) (string, error) {

	// for host, guest := range config.Volumes {
	// 	args = append(args, "-v", fmt.Sprintf("%s:%s", host, guest))
	// }

	conf := container.Config{
		AttachStdout: false,
		Tty:          true,
		Env:          []string{},
		Cmd:          []string{"/bin/ash"}, // TODO
		Image:        config.Image,
		//Volumes:      config.Volumes, // TODO
	}
	hostConf := container.HostConfig{
		Privileged: config.Privileged,
	}
	networkConf := network.NetworkingConfig{}

	d.Ui.Message("Creating container")
	//  -d -i -t {{.Image}} /bin/bash
	resp, err := d.client.ContainerCreate(context.Background(), &conf, &hostConf,
		&networkConf, "")
	if err != nil {
		return "", err
	}
	for warning := range resp.Warnings {
		log.Printf("Warning: %s\n", warning)
	}

	d.Ui.Message("Starting container")
	opts := types.ContainerStartOptions{}
	err = d.client.ContainerStart(context.Background(), resp.ID, opts)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d DockerApiDriver) StopContainer(id string) error {

	err := d.client.ContainerKill(context.Background(), id, "KILL")
	if err != nil {
		return err
	}

	return d.client.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
}

func (d DockerApiDriver) TagImage(id string, repo string, force bool) error {
	// TODO force
	return d.client.ImageTag(context.Background(), id, repo)
}

func (d DockerApiDriver) Verify() error {
	d.Ui.Say("Warning! You are running the EXPERMINATAL Docker API driver!")
	var err error
	d.client, err = client.NewEnvClient()
	return err
}

func (d DockerApiDriver) Version() (*version.Version, error) {
	// output, err := exec.Command("docker", "-v").Output()
	// if err != nil {
	// 	return nil, err
	// }

	// match := regexp.MustCompile(version.VersionRegexpRaw).FindSubmatch(output)
	// if match == nil {
	// 	return nil, fmt.Errorf("unknown version: %s", output)
	// }

	// return version.NewVersion(string(match[0]))
	return version.NewVersion("1.0")
}
