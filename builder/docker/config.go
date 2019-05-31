//go:generate struct-markdown

package docker

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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

	Author           string
	Changes          []string
	Commit           bool
	// The directory inside container to mount temp
    // directory from host server for work file
    // provisioner. This defaults to
    // c:/packer-files on windows and /packer-files on other systems.
	ContainerDir     string `mapstructure:"container_dir" required:"false"`
	Discard          bool
	// Username (UID) to run remote commands with. You can
    // also set the group name/ID if you want: (UID or UID:GID).
    // You may need this if you get permission errors trying to run the shell or
    // other provisioners.
	ExecUser         string `mapstructure:"exec_user" required:"false"`
	ExportPath       string `mapstructure:"export_path"`
	Image            string
	Message          string
	// If true, run the docker container with the
    // --privileged flag. This defaults to false if not set.
	Privileged       bool `mapstructure:"privileged" required:"false"`
	Pty              bool
	Pull             bool
	// An array of arguments to pass to
    // docker run in order to run the container. By default this is set to
    // ["-d", "-i", "-t", "--entrypoint=/bin/sh", "--", "{{.Image}}"] if you are
    // using a linux container, and
    // ["-d", "-i", "-t", "--entrypoint=powershell", "--", "{{.Image}}"] if you
    // are running a windows container. {{.Image}} is a template variable that
    // corresponds to the image template option. Passing the entrypoint option
    // this way will make it the default entrypoint of the resulting image, so
    // running docker run -it --rm  will start the docker image from the
    // /bin/sh shell interpreter; you could run a script or another shell by
    // running docker run -it --rm  -c /bin/bash. If your docker image
    // embeds a binary intended to be run often, you should consider changing the
    // default entrypoint to point to it.
	RunCommand       []string `mapstructure:"run_command" required:"false"`
	Volumes          map[string]string
	// If true, files uploaded to the container
    // will be owned by the user the container is running as. If false, the owner
    // will depend on the version of docker installed in the system. Defaults to
    // true.
	FixUploadOwner   bool `mapstructure:"fix_upload_owner" required:"false"`
	// If "true", tells Packer that you are building a
    // Windows container running on a windows host. This is necessary for building
    // Windows containers, because our normal docker bindings do not work for them.
	WindowsContainer bool `mapstructure:"windows_container" required:"false"`

	// This is used to login to dockerhub to pull a private base container. For
	// pushing to dockerhub, see the docker post-processors
	Login           bool
	// The password to use to authenticate to login.
	LoginPassword   string `mapstructure:"login_password" required:"false"`
	// The server address to login to.
	LoginServer     string `mapstructure:"login_server" required:"false"`
	// The username to use to authenticate to login.
	LoginUsername   string `mapstructure:"login_username" required:"false"`
	// Defaults to false. If true, the builder will login
    // in order to pull the image from Amazon EC2 Container Registry
    // (ECR). The builder only logs in for the
    // duration of the pull. If true login_server is required and login,
    // login_username, and login_password will be ignored. For more
    // information see the section on ECR.
	EcrLogin        bool   `mapstructure:"ecr_login" required:"false"`
	AwsAccessConfig `mapstructure:",squash"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	c.FixUploadOwner = true

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
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
		if k == "Pull" {
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

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs, errImageNotSpecified)
	}

	if (c.ExportPath != "" && c.Commit) || (c.ExportPath != "" && c.Discard) || (c.Commit && c.Discard) {
		errs = packer.MultiErrorAppend(errs, errArtifactUseConflict)
	}

	if c.ExportPath == "" && !c.Commit && !c.Discard {
		errs = packer.MultiErrorAppend(errs, errArtifactNotUsed)
	}

	if c.ExportPath != "" {
		if fi, err := os.Stat(c.ExportPath); err == nil && fi.IsDir() {
			errs = packer.MultiErrorAppend(errs, errExportPathNotFile)
		}
	}

	if c.ContainerDir == "" {
		if runtime.GOOS == "windows" {
			c.ContainerDir = "c:/packer-files"
		} else {
			c.ContainerDir = "/packer-files"
		}
	}

	if c.EcrLogin && c.LoginServer == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("ECR login requires login server to be provided."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
