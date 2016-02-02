package common

import (
	"fmt"
    "os"
    "path"

	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/service/sts"
    "github.com/go-ini/ini"
)

type CLIConfig struct {
    SourceProfile   string

    Source          credentials.Value
    AssumeRoleInput sts.AssumeRoleInput
}

// Sets params in the struct based on the file section
func (c *CLIConfig) Prepare(name string) (error) {
    cfg, err := c.config()

    cfg_profile_name := fmt.Sprintf("profile %s", name)
    profile_cfg, err := cfg.GetSection(cfg_profile_name)
	if err != nil {
		return err
	}

	c.SourceProfile = profile_cfg.Key("source_profile").Value();
	if c.SourceProfile == "" {
		c.SourceProfile = name
	}

    c.AssumeRoleInput.RoleArn = aws.String(profile_cfg.Key("role_arn").Value())

	host, err := os.Hostname()
	if err != nil {
		return err
	}

	sessName := fmt.Sprintf("packer-%s", host)
    c.AssumeRoleInput.RoleSessionName = &sessName
	c.AssumeRoleInput.SerialNumber = aws.String(profile_cfg.Key("mfa_serial").Value())
	if extId := aws.String(profile_cfg.Key("external_id").Value()); extId != nil {
		c.AssumeRoleInput.ExternalId = extId
	}

    creds, err := c.credentials()
    cred_cfg, err := creds.GetSection(c.SourceProfile)
	if err != nil {
		return err
	}

    if len(c.SourceProfile) != 0 {
        c.Source.AccessKeyID = cred_cfg.Key("aws_access_key_id").Value()
        c.Source.SecretAccessKey = cred_cfg.Key("aws_secret_access_key").Value()
        c.Source.SessionToken = cred_cfg.Key("aws_session_token").Value()
    }
    return nil
}

func (c *CLIConfig) config() (*ini.File, error) {
	config_path := os.Getenv("AWS_CONFIG_FILE")
    if config_path == "" {
        config_path = path.Join(os.Getenv("HOME"), ".aws", "config")
    }
    ini, err := c.readFile(config_path)
	if err != nil {
		return nil, err
	}

    return ini, nil
}

func (c *CLIConfig) credentials() (*ini.File, error) {
	cred_path := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
    if cred_path == "" {
        cred_path = path.Join(os.Getenv("HOME"), ".aws", "credentials")
    }
    ini, err := c.readFile(cred_path)
	if err != nil {
		return nil, err
	}
    return ini, nil
}

func (c *CLIConfig) readFile(path string) (*ini.File, error) {
    cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
    return cfg, nil
}