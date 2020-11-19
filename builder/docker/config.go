//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package docker

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

var (
	errArtifactNotUsed     = fmt.Errorf("No instructions given for handling the artifact; expected commit, discard, or export_path")
	errArtifactUseConflict = fmt.Errorf("Cannot specify more than one of commit, discard, and export_path")
	errExportPathNotFile   = fmt.Errorf("export_path must be a file, not a directory")
	errImageNotSpecified   = fmt.Errorf("Image must be specified")
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	// Set the author (e-mail) of a commit.
	Author string `mapstructure:"author"`
	// Dockerfile instructions to add to the commit. Example of instructions
	// are CMD, ENTRYPOINT, ENV, and EXPOSE. Example: [ "USER ubuntu", "WORKDIR
	// /app", "EXPOSE 8080" ]
	Changes []string `mapstructure:"changes"`
	// If true, the container will be committed to an image rather than exported.
	Commit bool `mapstructure:"commit" required:"true"`

	// The directory inside container to mount temp directory from host server
	// for work [file provisioner](/docs/provisioners/file). This defaults
	// to c:/packer-files on windows and /packer-files on other systems.
	ContainerDir string `mapstructure:"container_dir" required:"false"`
	// An array of devices which will be accessible in container when it's run
	// without `--privileged` flag.
	Device []string `mapstructure:"device" required:"false"`
	// Throw away the container when the build is complete. This is useful for
	// the [artifice
	// post-processor](/docs/post-processors/artifice).
	Discard bool `mapstructure:"discard" required:"true"`
	// An array of additional [Linux
	// capabilities](https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities)
	// to grant to the container.
	CapAdd []string `mapstructure:"cap_add" required:"false"`
	// An array of [Linux
	// capabilities](https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities)
	// to drop from the container.
	CapDrop []string `mapstructure:"cap_drop" required:"false"`
	// Username (UID) to run remote commands with. You can also set the group
	// name/ID if you want: (UID or UID:GID). You may need this if you get
	// permission errors trying to run the shell or other provisioners.
	ExecUser string `mapstructure:"exec_user" required:"false"`
	// The path where the final container will be exported as a tar file.
	ExportPath string `mapstructure:"export_path" required:"true"`
	// The base image for the Docker container that will be started. This image
	// will be pulled from the Docker registry if it doesn't already exist.
	Image string `mapstructure:"image" required:"true"`
	// Set a message for the commit.
	Message string `mapstructure:"message" required:"true"`
	// If true, run the docker container with the `--privileged` flag. This
	// defaults to false if not set.
	Privileged bool `mapstructure:"privileged" required:"false"`
	Pty        bool
	// If true, the configured image will be pulled using `docker pull` prior
	// to use. Otherwise, it is assumed the image already exists and can be
	// used. This defaults to true if not set.
	Pull bool `mapstructure:"pull" required:"false"`
	// An array of arguments to pass to docker run in order to run the
	// container. By default this is set to `["-d", "-i", "-t",
	// "--entrypoint=/bin/sh", "--", "{{.Image}}"]` if you are using a linux
	// container, and `["-d", "-i", "-t", "--entrypoint=powershell", "--",
	// "{{.Image}}"]` if you are running a windows container. `{{.Image}}` is a
	// template variable that corresponds to the image template option. Passing
	// the entrypoint option this way will make it the default entrypoint of
	// the resulting image, so running docker run -it --rm  will start the
	// docker image from the /bin/sh shell interpreter; you could run a script
	// or another shell by running docker run -it --rm  -c /bin/bash. If your
	// docker image embeds a binary intended to be run often, you should
	// consider changing the default entrypoint to point to it.
	RunCommand []string `mapstructure:"run_command" required:"false"`
	// An array of additional tmpfs volumes to mount into this container.
	TmpFs []string `mapstructure:"tmpfs" required:"false"`
	// A mapping of additional volumes to mount into this container. The key of
	// the object is the host path, the value is the container path.
	Volumes map[string]string `mapstructure:"volumes" required:"false"`
	// If true, files uploaded to the container will be owned by the user the
	// container is running as. If false, the owner will depend on the version
	// of docker installed in the system. Defaults to true.
	FixUploadOwner bool `mapstructure:"fix_upload_owner" required:"false"`
	// If "true", tells Packer that you are building a Windows container
	// running on a windows host. This is necessary for building Windows
	// containers, because our normal docker bindings do not work for them.
	WindowsContainer bool `mapstructure:"windows_container" required:"false"`

	// This is used to login to dockerhub to pull a private base container. For
	// pushing to dockerhub, see the docker post-processors
	Login bool `mapstructure:"login" required:"false"`
	// The password to use to authenticate to login.
	LoginPassword string `mapstructure:"login_password" required:"false"`
	// The server address to login to.
	LoginServer string `mapstructure:"login_server" required:"false"`
	// The username to use to authenticate to login.
	LoginUsername string `mapstructure:"login_username" required:"false"`
	// Defaults to false. If true, the builder will login in order to pull the
	// image from Amazon EC2 Container Registry (ECR). The builder only logs in
	// for the duration of the pull. If true login_server is required and
	// login, login_username, and login_password will be ignored. For more
	// information see the section on ECR.
	EcrLogin        bool `mapstructure:"ecr_login" required:"false"`
	AwsAccessConfig `mapstructure:",squash"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {

	c.FixUploadOwner = true

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Defaults
	if len(c.RunCommand) == 0 {
		c.RunCommand = []string{"-d", "-i", "-t", "--entrypoint=/bin/sh", "--", "{{.Image}}"}
		if c.WindowsContainer {
			c.RunCommand = []string{"-d", "-i", "-t", "--entrypoint=powershell", "--", "{{.Image}}"}
		}
	}

	// Default Pull if it wasn't set
	hasPull := false
	for _, k := range md.Keys {
		if k == "pull" {
			hasPull = true
			break
		}
	}

	if !hasPull {
		c.Pull = true
	}

	// Default to the normal Docker type
	if c.Comm.Type == "" {
		c.Comm.Type = "docker"
		if c.WindowsContainer {
			c.Comm.Type = "dockerWindowsContainer"
		}
	}

	var errs *packersdk.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packersdk.MultiErrorAppend(errs, es...)
	}
	if c.Image == "" {
		errs = packersdk.MultiErrorAppend(errs, errImageNotSpecified)
	}

	if (c.ExportPath != "" && c.Commit) || (c.ExportPath != "" && c.Discard) || (c.Commit && c.Discard) {
		errs = packersdk.MultiErrorAppend(errs, errArtifactUseConflict)
	}

	if c.ExportPath == "" && !c.Commit && !c.Discard {
		errs = packersdk.MultiErrorAppend(errs, errArtifactNotUsed)
	}

	if c.ExportPath != "" {
		if fi, err := os.Stat(c.ExportPath); err == nil && fi.IsDir() {
			errs = packersdk.MultiErrorAppend(errs, errExportPathNotFile)
		}
	}

	if c.ContainerDir == "" {
		if c.WindowsContainer {
			c.ContainerDir = "c:/packer-files"
		} else {
			c.ContainerDir = "/packer-files"
		}
	}

	if c.EcrLogin && c.LoginServer == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("ECR login requires login server to be provided."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}
