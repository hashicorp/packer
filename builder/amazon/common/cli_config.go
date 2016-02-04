package common

import (
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

type CLIConfig struct {
	ProfileName   string
	SourceProfile string

	AssumeRoleInput   *sts.AssumeRoleInput
	SourceCredentials *credentials.Credentials

	profileCfg  *ini.Section
	profileCred *ini.Section
}

// Return a new CLIConfig with stored profile settings
func NewFromProfile(name string) (*CLIConfig, error) {
	c := &CLIConfig{}
	c.AssumeRoleInput = new(sts.AssumeRoleInput)
	err := c.Prepare(name)
	if err != nil {
		return nil, err
	}
	sessName, err := c.getSessionName(c.profileCfg.Key("role_session_name").Value())
	if err != nil {
		return nil, err
	}
	c.AssumeRoleInput.RoleSessionName = aws.String(sessName)
	arn := c.profileCfg.Key("role_arn").Value()
	if arn != "" {
		c.AssumeRoleInput.RoleArn = aws.String(arn)
	}
	id := c.profileCfg.Key("external_id").Value()
	if id != "" {
		c.AssumeRoleInput.ExternalId = aws.String(id)
	}
	c.SourceCredentials = credentials.NewStaticCredentials(
		c.profileCred.Key("aws_access_key_id").Value(),
		c.profileCred.Key("aws_secret_access_key").Value(),
		c.profileCred.Key("aws_session_token").Value(),
	)
	return c, nil
}

// Return AWS Credentials using current profile. Must supply source config.
func (c *CLIConfig) CredentialsFromProfile(conf *aws.Config) (*credentials.Credentials, error) {
	// If the profile name is equal to the source profile, there is no role to assume so return
	// the source credentials as they were captured.
	if c.ProfileName == c.SourceProfile {
		return c.SourceCredentials, nil
	}
	srcCfg := aws.NewConfig().Copy(conf).WithCredentials(c.SourceCredentials)
	svc := sts.New(session.New(), srcCfg)
	res, err := svc.AssumeRole(c.AssumeRoleInput)
	if err != nil {
		return nil, err
	}
	return credentials.NewStaticCredentials(
		*res.Credentials.AccessKeyId,
		*res.Credentials.SecretAccessKey,
		*res.Credentials.SessionToken,
	), nil
}

// Sets params in the struct based on the file section
func (c *CLIConfig) Prepare(name string) error {
	var err error
	c.ProfileName = name
	c.profileCfg, err = configFromName(c.ProfileName)
	if err != nil {
		return err
	}
	c.SourceProfile = c.profileCfg.Key("source_profile").Value()
	if c.SourceProfile == "" {
		c.SourceProfile = c.ProfileName
	}
	c.profileCred, err = credsFromName(c.SourceProfile)
	if err != nil {
		return err
	}
	return nil
}

func (c *CLIConfig) getSessionName(rawName string) (string, error) {
	if rawName == "" {
		name := "packer-"
		host, err := os.Hostname()
		if err != nil {
			return name, err
		}
		return fmt.Sprintf("%s%s", name, host), nil
	} else {
		return rawName, nil
	}
}

func configFromName(name string) (*ini.Section, error) {
	filePath := os.Getenv("AWS_CONFIG_FILE")
	if filePath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		filePath = path.Join(home, ".aws", "config")
	}
	file, err := readFile(filePath)
	if err != nil {
		return nil, err
	}
	profileName := fmt.Sprintf("profile %s", name)
	cfg, err := file.GetSection(profileName)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func credsFromName(name string) (*ini.Section, error) {
	filePath := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if filePath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		filePath = path.Join(home, ".aws", "credentials")
	}
	file, err := readFile(filePath)
	if err != nil {
		return nil, err
	}
	cfg, err := file.GetSection(name)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func readFile(path string) (*ini.File, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
