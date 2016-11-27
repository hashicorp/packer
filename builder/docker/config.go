package docker

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var (
	errArtifactNotUsed        = fmt.Errorf("No instructions given for handling the artifact; expected commit, discard, or export_path")
	errArtifactUseConflict    = fmt.Errorf("Cannot specify more than one of commit, discard, and export_path")
	errExportPathNotFile      = fmt.Errorf("export_path must be a file, not a directory")
	errDockerCertPathNotFound = fmt.Errorf("docker_cert_path could not be found")
	errImageNotSpecified      = fmt.Errorf("Image must be specified")
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	DockerHostConfig    DockerHostConfig    `mapstructure:",squash"`

	Commit     bool
	Discard    bool
	ExportPath string `mapstructure:"export_path"`
	Image      string
	Pty        bool
	Pull       *bool
	RunCommand []string `mapstructure:"run_command"`
	Volumes    map[string]string
	Privileged bool `mapstructure:"privileged"`
	Author     string
	Changes    []string
	Message    string

	// This is used to login to dockerhub to pull a private base container. For
	// pushing to dockerhub, see the docker post-processors
	Login           bool
	LoginEmail      string `mapstructure:"login_email"`
	LoginPassword   string `mapstructure:"login_password"`
	LoginServer     string `mapstructure:"login_server"`
	LoginUsername   string `mapstructure:"login_username"`
	EcrLogin        bool   `mapstructure:"ecr_login"`
	AwsAccessConfig `mapstructure:",squash"`

	ctx interpolate.Context
}

type DockerHostConfig struct {
	Host      string `mapstructure:"docker_host"`
	TlsVerify *bool  `mapstructure:"docker_tls_verify"`
	CertPath  string `mapstructure:"docker_cert_path"`
}

func (c *DockerHostConfig) Prepare() []error {
	var errs []error

	if c.CertPath != "" {
		if fi, err := os.Stat(c.CertPath); err != nil && !fi.IsDir() {
			errs = append(errs, errDockerCertPathNotFound)
		}
		files := []string{"ca.pem", "cert.pem", "key.pem"}
		for i := range files {
			if fi, err := os.Stat(filepath.Join(c.CertPath, files[i])); err != nil && fi.IsDir() {
				errs = append(errs, fmt.Errorf("Could not read file: %s", files[i]))
			}
		}
	}
	return errs
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
	log.Printf("DockerHostConfig: %v", c.DockerHostConfig)

	// Defaults
	if len(c.RunCommand) == 0 {
		c.RunCommand = []string{"-d", "-i", "-t", "{{.Image}}", "/bin/bash"}
	}

	// Default Pull if it wasn't set
	if c.Pull == nil {
		t := true
		c.Pull = &t
	}

	// Default Docker TLS Verify if it wasn't set
	if c.DockerHostConfig.TlsVerify == nil {
		t := true
		c.DockerHostConfig.TlsVerify = &t
	}

	// Default to the normal Docker type
	if c.Comm.Type == "" {
		c.Comm.Type = "docker"
	}

	var errs *packer.MultiError

	if es := c.DockerHostConfig.Prepare(); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

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

	if c.EcrLogin && c.LoginServer == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("ECR login requires login server to be provided."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	log.Printf("Config: %v", c)
	log.Printf("Comm: %v", &c.Comm)
	log.Printf("DockerHostConfig: %v", &c.DockerHostConfig)
	return c, nil, nil
}
