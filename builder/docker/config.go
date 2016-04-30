package docker

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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

	Commit     bool
	Discard    bool
	ExportPath string `mapstructure:"export_path"`
	Image      string
	Pty        bool
	Pull       bool
	RunCommand []string `mapstructure:"run_command"`
	Volumes    map[string]string
	Privileged bool `mapstructure:"privileged"`

	// This is used to login to dockerhub to pull a private base container. For
	// pushing to dockerhub, see the docker post-processors
	Login         bool
	LoginEmail    string `mapstructure:"login_email"`
	LoginPassword string `mapstructure:"login_password"`
	LoginServer   string `mapstructure:"login_server"`
	LoginUsername string `mapstructure:"login_username"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

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
		c.RunCommand = []string{"-d", "-i", "-t", "{{.Image}}", "/bin/bash"}
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

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}
