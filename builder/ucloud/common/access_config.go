//go:generate struct-markdown
package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/builder/ucloud/version"
	"github.com/ucloud/ucloud-sdk-go/external"
	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/ufile"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

type AccessConfig struct {
	// This is the UCloud public key. It must be provided unless `profile` is set,
	// but it can also be sourced from the `UCLOUD_PUBLIC_KEY` environment variable.
	PublicKey string `mapstructure:"public_key" required:"true"`
	// This is the UCloud private key. It must be provided unless `profile` is set,
	// but it can also be sourced from the `UCLOUD_PRIVATE_KEY` environment variable.
	PrivateKey string `mapstructure:"private_key" required:"true"`
	// This is the UCloud region. It must be provided, but it can also be sourced from
	// the `UCLOUD_REGION` environment variables.
	Region string `mapstructure:"region" required:"true"`
	// This is the UCloud project id. It must be provided, but it can also be sourced
	// from the `UCLOUD_PROJECT_ID` environment variables.
	ProjectId string `mapstructure:"project_id" required:"true"`
	// This is the base url. (Default: `https://api.ucloud.cn`).
	BaseUrl string `mapstructure:"base_url" required:"false"`
	// This is the UCloud profile name as set in the shared credentials file, it can
	// also be sourced from the `UCLOUD_PROFILE` environment variables.
	Profile string `mapstructure:"profile" required:"false"`
	// This is the path to the shared credentials file, it can also be sourced from
	// the `UCLOUD_SHARED_CREDENTIAL_FILE` environment variables. If this is not set
	// and a profile is specified, `~/.ucloud/credential.json` will be used.
	SharedCredentialsFile string `mapstructure:"shared_credentials_file" required:"false"`

	client *UCloudClient
}

type cloudShellCredential struct {
	Cookie    string `json:"cookie"`
	Profile   string `json:"profile"`
	CSRFToken string `json:"csrf_token"`
}

func (c *AccessConfig) Client() (*UCloudClient, error) {
	if c.client != nil {
		return c.client, nil
	}

	cfg := ucloud.NewConfig()
	cfg.Region = c.Region
	cfg.ProjectId = c.ProjectId
	if c.BaseUrl != "" {
		cfg.BaseUrl = c.BaseUrl
	}
	cfg.LogLevel = log.PanicLevel
	cfg.UserAgent = fmt.Sprintf("Packer-UCloud/%s", version.UcloudPluginVersion.FormattedVersion())
	// set default max retry count
	cfg.MaxRetries = 3

	cred := auth.NewCredential()
	var cloudShellCredHandler ucloud.HttpRequestHandler
	if len(c.Profile) > 0 {
		// load public/private key from shared credential file
		credV, err := external.LoadUCloudCredentialFile(c.SharedCredentialsFile, c.Profile)
		if err != nil {
			return nil, fmt.Errorf("cannot load shared %q credential file, %s", c.Profile, err)
		}
		cred = *credV
	} else if len(c.PublicKey) > 0 && len(c.PrivateKey) > 0 {
		cred.PublicKey = c.PublicKey
		cred.PrivateKey = c.PrivateKey
	} else if v := os.Getenv("CLOUD_SHELL"); v == "true" {
		csCred := make([]cloudShellCredential, 0)
		// load credential from default cloud shell credential path
		if err := loadJSONFile(defaultCloudShellCredPath(), &csCred); err != nil {
			return nil, fmt.Errorf("must set credential about public_key and private_key, %s", err)
		}
		// get default cloud shell credential
		defaultCsCred := &cloudShellCredential{}
		for i := 0; i < len(csCred); i++ {
			if csCred[i].Profile == "default" {
				defaultCsCred = &csCred[i]
				break
			}
		}
		if defaultCsCred == nil || len(defaultCsCred.Cookie) == 0 || len(defaultCsCred.CSRFToken) == 0 {
			return nil, fmt.Errorf("must set credential about public_key and private_key, default credential is null")
		}

		// set cloud shell client handler
		cloudShellCredHandler = func(c *ucloud.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
			if err := req.SetHeader("Cookie", defaultCsCred.Cookie); err != nil {
				return nil, err
			}
			if err := req.SetHeader("Csrf-Token", defaultCsCred.CSRFToken); err != nil {
				return nil, err
			}
			return req, nil
		}
	} else {
		return nil, fmt.Errorf("must set credential about public_key and private_key")
	}

	c.client = &UCloudClient{}
	c.client.UHostConn = uhost.NewClient(&cfg, &cred)
	c.client.UNetConn = unet.NewClient(&cfg, &cred)
	c.client.VPCConn = vpc.NewClient(&cfg, &cred)
	c.client.UAccountConn = uaccount.NewClient(&cfg, &cred)
	c.client.UFileConn = ufile.NewClient(&cfg, &cred)

	if cloudShellCredHandler != nil {
		if err := c.client.UHostConn.AddHttpRequestHandler(cloudShellCredHandler); err != nil {
			return nil, err
		}
		if err := c.client.UNetConn.AddHttpRequestHandler(cloudShellCredHandler); err != nil {
			return nil, err
		}
		if err := c.client.VPCConn.AddHttpRequestHandler(cloudShellCredHandler); err != nil {
			return nil, err
		}
		if err := c.client.UAccountConn.AddHttpRequestHandler(cloudShellCredHandler); err != nil {
			return nil, err
		}
		if err := c.client.UFileConn.AddHttpRequestHandler(cloudShellCredHandler); err != nil {
			return nil, err
		}
	}

	return c.client, nil
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if c.Region == "" {
		c.Region = os.Getenv("UCLOUD_REGION")
	}

	if c.Region == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "region"))
	}

	if c.ProjectId == "" {
		c.ProjectId = os.Getenv("UCLOUD_PROJECT_ID")
	}

	if c.ProjectId == "" {
		errs = append(errs, fmt.Errorf("%q must be set", "projectId"))
	}

	if c.BaseUrl != "" {
		if _, err := url.Parse(c.BaseUrl); err != nil {
			errs = append(errs, fmt.Errorf("%q is invalid, should be an valid ucloud base_url, got %q, parse error: %s", "base_url", c.BaseUrl, err))
		}
	}

	if c.Profile == "" {
		c.Profile = os.Getenv("UCLOUD_PROFILE")
	}

	if c.SharedCredentialsFile == "" {
		c.SharedCredentialsFile = os.Getenv("UCLOUD_SHARED_CREDENTIAL_FILE")
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AccessConfig) Config() error {
	if c.PublicKey == "" {
		c.PublicKey = os.Getenv("UCLOUD_PUBLIC_KEY")
	}
	if c.PrivateKey == "" {
		c.PrivateKey = os.Getenv("UCLOUD_PRIVATE_KEY")
	}

	if c.Profile == "" {
		c.Profile = os.Getenv("UCLOUD_PROFILE")
	}

	if c.SharedCredentialsFile == "" {
		c.SharedCredentialsFile = os.Getenv("UCLOUD_SHARED_CREDENTIAL_FILE")
	}

	if (c.PublicKey == "" || c.PrivateKey == "") && c.Profile == "" && os.Getenv("CLOUD_SHELL") != "true" {
		return fmt.Errorf("%q, %q must be set in template file or environment variables", "public_key", "private_key")
	}
	return nil

}

func (c *AccessConfig) ValidateProjectId(projectId string) error {
	supportedProjectIds, err := c.getSupportedProjectIds()
	if err != nil {
		return err
	}

	for _, supportedProjectId := range supportedProjectIds {
		if projectId == supportedProjectId {
			return nil
		}
	}

	return fmt.Errorf("%q is invalid, should be an valid ucloud project_id, got %q", "project_id", projectId)
}

func (c *AccessConfig) ValidateRegion(region string) error {
	supportedRegions, err := c.getSupportedRegions()
	if err != nil {
		return err
	}

	for _, supportedRegion := range supportedRegions {
		if region == supportedRegion {
			return nil
		}
	}

	return fmt.Errorf("%q is invalid, should be an valid ucloud region, got %q", "region", region)
}

func (c *AccessConfig) ValidateZone(region, zone string) error {
	supportedZones, err := c.getSupportedZones(region)
	if err != nil {
		return err
	}

	for _, supportedZone := range supportedZones {
		if zone == supportedZone {
			return nil
		}
	}

	return fmt.Errorf("%q is invalid, should be an valid ucloud zone, got %q", "availability_zone", zone)
}

func (c *AccessConfig) getSupportedProjectIds() ([]string, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}
	conn := client.UAccountConn
	req := conn.NewGetProjectListRequest()
	resp, err := conn.GetProjectList(req)
	if err != nil {
		return nil, err
	}

	validProjectIds := make([]string, len(resp.ProjectSet))
	for _, val := range resp.ProjectSet {
		if !IsStringIn(val.ProjectId, validProjectIds) {
			validProjectIds = append(validProjectIds, val.ProjectId)
		}
	}

	return validProjectIds, nil
}

func (c *AccessConfig) getSupportedRegions() ([]string, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	conn := client.UAccountConn
	req := conn.NewGetRegionRequest()
	resp, err := conn.GetRegion(req)
	if err != nil {
		return nil, err
	}

	validRegions := make([]string, len(resp.Regions))
	for _, val := range resp.Regions {
		if !IsStringIn(val.Region, validRegions) {
			validRegions = append(validRegions, val.Region)
		}
	}

	return validRegions, nil
}

func (c *AccessConfig) getSupportedZones(region string) ([]string, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	conn := client.UAccountConn
	req := conn.NewGetRegionRequest()
	resp, err := conn.GetRegion(req)
	if err != nil {
		return nil, err
	}

	validZones := make([]string, len(resp.Regions))
	for _, val := range resp.Regions {
		if val.Region == region && !IsStringIn(val.Zone, validZones) {
			validZones = append(validZones, val.Zone)
		}

	}

	return validZones, nil
}

func defaultCloudShellCredPath() string {
	return filepath.Join(userHomeDir(), ".ucloud", "credential.json")
}

func loadJSONFile(path string, p interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, p)
	if err != nil {
		return err
	}
	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
