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
	// Path to "VMware Fusion.app". By default this is
    // /Applications/VMware Fusion.app but this setting allows you to
    // customize this.
	FusionAppPath           string `mapstructure:"fusion_app_path" required:"false"`
	// The type of remote machine that will be used to
    // build this VM rather than a local desktop product. The only value accepted
    // for this currently is esx5. If this is not set, a desktop product will
    // be used. By default, this is not set.
	RemoteType              string `mapstructure:"remote_type" required:"false"`
	// The path to the datastore where the VM will be stored
    // on the ESXi machine.
	RemoteDatastore         string `mapstructure:"remote_datastore" required:"false"`
	// The path to the datastore where supporting files
    // will be stored during the build on the remote machine.
	RemoteCacheDatastore    string `mapstructure:"remote_cache_datastore" required:"false"`
	// The path where the ISO and/or floppy files will
    // be stored during the build on the remote machine. The path is relative to
    // the remote_cache_datastore on the remote machine.
	RemoteCacheDirectory    string `mapstructure:"remote_cache_directory" required:"false"`
	// The host of the remote machine used for access.
    // This is only required if remote_type is enabled.
	RemoteHost              string `mapstructure:"remote_host" required:"false"`
	// The SSH port of the remote machine
	RemotePort              int    `mapstructure:"remote_port" required:"false"`
	// The SSH username used to access the remote machine.
	RemoteUser              string `mapstructure:"remote_username" required:"false"`
	// The SSH password for access to the remote machine.
	RemotePassword          string `mapstructure:"remote_password" required:"false"`
	// The SSH key for access to the remote machine.
	RemotePrivateKey        string `mapstructure:"remote_private_key_file" required:"false"`
	// When Packer is preparing to run a
    // remote esxi build, and export is not disable, by default it runs a no-op
    // ovftool command to make sure that the remote_username and remote_password
    // given are valid. If you set this flag to true, Packer will skip this
    // validation. Default: false.
	SkipValidateCredentials bool   `mapstructure:"skip_validate_credentials" required:"false"`
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
