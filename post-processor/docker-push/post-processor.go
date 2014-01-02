package dockerpush

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os/exec"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Registry string `mapstructure:"registry"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)

	templates := map[string]*string{
		"username": &p.config.Username,
		"password": &p.config.Password,
	}

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil

}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	id := artifact.Id()
	ui.Say("Pushing image: " + id)

	if p.config.Registry == "" {

		if p.config.Email == "" {
			cmd := exec.Command("docker",
				"login",
				"-u="+p.config.Username,
				"-p="+p.config.Password)

			if err := cmd.Run(); err != nil {
				ui.Say("Login to the registry " + p.config.Registry + " failed")
				return nil, false, err
			}

		} else {
			cmd := exec.Command("docker",
				"login",
				"-u="+p.config.Username,
				"-p="+p.config.Password,
				"-e="+p.config.Email)

			if err := cmd.Run(); err != nil {
				ui.Say("Login to the registry " + p.config.Registry + " failed")
				return nil, false, err
			}

		}

	} else {
		if p.config.Email == "" {
			cmd := exec.Command("docker",
				"login",
				"-u="+p.config.Username,
				"-p="+p.config.Password,
				p.config.Registry)

			if err := cmd.Run(); err != nil {
				ui.Say("Login to the registry " + p.config.Registry + " failed")
				return nil, false, err
			}

		} else {
			cmd := exec.Command("docker",
				"login",
				"-u="+p.config.Username,
				"-p="+p.config.Password,
				"-e="+p.config.Email,
				p.config.Registry)

			if err := cmd.Run(); err != nil {
				ui.Say("Login to the registry " + p.config.Registry + " failed")
				return nil, false, err
			}

		}
	}

	cmd := exec.Command("docker", "push", id)
	if err := cmd.Run(); err != nil {
		ui.Say("Failed to push image: " + id)
		return nil, false, err
	}

	return nil, true, nil
}
