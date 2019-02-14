package common

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/packer/template/interpolate"
)

type DriverConfig struct {
	FusionAppPath           string `mapstructure:"fusion_app_path"`
	RemoteType              string `mapstructure:"remote_type"`
	RemoteDatastore         string `mapstructure:"remote_datastore"`
	RemoteCacheDatastore    string `mapstructure:"remote_cache_datastore"`
	RemoteCacheDirectory    string `mapstructure:"remote_cache_directory"`
	RemoteHost              string `mapstructure:"remote_host"`
	RemotePort              uint   `mapstructure:"remote_port"`
	RemoteUser              string `mapstructure:"remote_username"`
	RemotePassword          string `mapstructure:"remote_password"`
	RemotePrivateKey        string `mapstructure:"remote_private_key_file"`
	SkipValidateCredentials bool   `mapstructure:"skip_validate_credentials"`
}

func (c *DriverConfig) Prepare(ctx *interpolate.Context) []error {
	if c.FusionAppPath == "" {
		c.FusionAppPath = os.Getenv("FUSION_APP_PATH")
	}
	if c.FusionAppPath == "" {
		c.FusionAppPath = "/Applications/VMware Fusion.app"
	}
	if c.RemoteUser == "" {
		c.RemoteUser = "root"
	}
	if c.RemoteDatastore == "" {
		c.RemoteDatastore = "datastore1"
	}
	if c.RemoteCacheDatastore == "" {
		c.RemoteCacheDatastore = c.RemoteDatastore
	}
	if c.RemoteCacheDirectory == "" {
		c.RemoteCacheDirectory = "packer_cache"
	}
	if c.RemotePort == 0 {
		c.RemotePort = 22
	}

	return nil
}

func (c *DriverConfig) Validate(SkipExport bool) error {
	if c.RemoteType == "" || SkipExport == true {
		return nil
	}
	if c.RemotePassword == "" {
		return fmt.Errorf("exporting the vm (with ovftool) requires that " +
			"you set a value for remote_password")
	}
	if c.SkipValidateCredentials {
		return nil
	}

	// check that password is valid by sending a dummy ovftool command
	// now, so that we don't fail for a simple mistake after a long
	// build
	ovftool := GetOVFTool()
	ovfToolArgs := []string{"--noSSLVerify", "--verifyOnly", fmt.Sprintf("vi://%s:%s@%s",
		url.QueryEscape(c.RemoteUser),
		url.QueryEscape(c.RemotePassword),
		c.RemoteHost)}

	var out bytes.Buffer
	cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, ovftool, ovfToolArgs...)
	cmd.Stdout = &out

	// Need to manually close stdin or else the ofvtool call will hang
	// forever in a situation where the user has provided an invalid
	// password or username
	stdin, _ := cmd.StdinPipe()
	defer stdin.Close()

	if err := cmd.Run(); err != nil {
		outString := out.String()
		// The command *should* fail with this error, if it
		// authenticates properly.
		if !strings.Contains(outString, "Found wrong kind of object") {
			err := fmt.Errorf("ovftool validation error: %s; %s",
				err, outString)
			if strings.Contains(outString,
				"Enter login information for source") {
				err = fmt.Errorf("The username or password you " +
					"provided to ovftool is invalid.")
			}
			return err
		}
	}

	return nil
}
