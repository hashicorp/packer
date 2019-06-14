package uhost

import (
	"fmt"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/hashicorp/packer/version"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"os"
)

type AccessConfig struct {
	PublicKey  string `mapstructure:"public_key"`
	PrivateKey string `mapstructure:"private_key"`
	Region     string `mapstructure:"region"`
	ProjectId  string `mapstructure:"project_id"`

	client *UCloudClient
}

func (c *AccessConfig) Client() (*UCloudClient, error) {
	if c.client != nil {
		return c.client, nil
	}

	cfg := ucloud.NewConfig()
	cfg.Region = c.Region
	cfg.ProjectId = c.ProjectId
	cfg.UserAgent = fmt.Sprintf("Packer-UCloud/%s", version.FormattedVersion())

	cred := auth.NewCredential()
	cred.PublicKey = c.PublicKey
	cred.PrivateKey = c.PrivateKey

	c.client = &UCloudClient{}
	c.client.uhostconn = uhost.NewClient(&cfg, &cred)
	c.client.unetconn = unet.NewClient(&cfg, &cred)
	c.client.vpcconn = vpc.NewClient(&cfg, &cred)
	c.client.uaccountconn = uaccount.NewClient(&cfg, &cred)

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

	if c.ProjectId == "" {
		c.ProjectId = os.Getenv("UCLOUD_PROJECT_ID")
	}

	if c.PublicKey == "" || c.PrivateKey == "" || c.ProjectId == "" {
		return fmt.Errorf("%q, %q, and %q must be set", "public_key", "private_key", "project_id")
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
	conn := client.uaccountconn
	if err != nil {
		return nil, err
	}

	req := conn.NewGetProjectListRequest()
	resp, err := conn.GetProjectList(req)
	if err != nil {
		return nil, err
	}

	validProjectIds := make([]string, len(resp.ProjectSet))
	for _, val := range resp.ProjectSet {
		if !isStringIn(val.ProjectId, validProjectIds) {
			validProjectIds = append(validProjectIds, val.ProjectId)
		}
	}

	return validProjectIds, nil
}

func (c *AccessConfig) getSupportedRegions() ([]string, error) {
	client, err := c.Client()
	conn := client.uaccountconn
	if err != nil {
		return nil, err
	}

	req := conn.NewGetRegionRequest()
	resp, err := conn.GetRegion(req)
	if err != nil {
		return nil, err
	}

	validRegions := make([]string, len(resp.Regions))
	for _, val := range resp.Regions {
		if !isStringIn(val.Region, validRegions) {
			validRegions = append(validRegions, val.Region)
		}
	}

	return validRegions, nil
}

func (c *AccessConfig) getSupportedZones(region string) ([]string, error) {
	client, err := c.Client()
	conn := client.uaccountconn
	if err != nil {
		return nil, err
	}

	req := conn.NewGetRegionRequest()
	resp, err := conn.GetRegion(req)
	if err != nil {
		return nil, err
	}

	validZones := make([]string, len(resp.Regions))
	for _, val := range resp.Regions {
		if val.Region == region && !isStringIn(val.Zone, validZones) {
			validZones = append(validZones, val.Zone)
		}

	}

	return validZones, nil
}
