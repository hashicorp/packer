package external

import (
	"fmt"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// CredentialProvider is the provider to store and provide credential instance
type CredentialProvider interface {
	Credential() *auth.Credential
}

// ClientConfigProvider is the provider to store and provide config instance
type ClientConfigProvider interface {
	Config() *ucloud.Config
}

// ConfigProvider is the provider to store and provide config/credential instance
type ConfigProvider interface {
	CredentialProvider
	ClientConfigProvider
}

// config will read configuration
type config struct {
	// Named profile
	Profile              string
	SharedConfigFile     string
	SharedCredentialFile string

	// Credential configuration
	PublicKey     string
	PrivateKey    string
	SecurityToken string
	CanExpire     bool
	Expires       time.Time

	// Client configuration
	ProjectId string
	Zone      string
	Region    string
	BaseUrl   string
	Timeout   time.Duration
}

func newConfig() *config {
	return &config{}
}

func (c *config) loadEnv() error {
	cfg, err := loadEnvConfig()
	if err != nil {
		return err
	}

	err = c.merge(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *config) loadFileIfExist() error {
	cfg, err := loadSharedConfigFile(
		c.SharedConfigFile,
		c.SharedCredentialFile,
		c.Profile,
	)
	if err != nil {
		return err
	}

	err = c.merge(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *config) merge(other *config) error {
	if other == nil {
		return nil
	}

	setStringify(&c.Profile, other.Profile)
	setStringify(&c.SharedConfigFile, other.SharedConfigFile)
	setStringify(&c.SharedCredentialFile, other.SharedCredentialFile)
	setStringify(&c.PublicKey, other.PublicKey)
	setStringify(&c.PrivateKey, other.PrivateKey)
	setStringify(&c.ProjectId, other.ProjectId)
	setStringify(&c.Zone, other.Zone)
	setStringify(&c.Region, other.Region)
	setStringify(&c.BaseUrl, other.BaseUrl)
	if other.Timeout != time.Duration(0) {
		c.Timeout = other.Timeout
	}
	return nil
}

// Config is the configuration of ucloud client
func (c *config) Config() *ucloud.Config {
	cfg := ucloud.NewConfig()
	setStringify(&cfg.ProjectId, c.ProjectId)
	setStringify(&cfg.Zone, c.Zone)
	setStringify(&cfg.Region, c.Region)
	setStringify(&cfg.BaseUrl, c.BaseUrl)

	if c.Timeout != time.Duration(0) {
		cfg.Timeout = c.Timeout
	}
	return &cfg
}

// Credential is the configuration of ucloud authorization information
func (c *config) Credential() *auth.Credential {
	cred := auth.NewCredential()
	setStringify(&cred.PublicKey, c.PublicKey)
	setStringify(&cred.PrivateKey, c.PrivateKey)
	setStringify(&cred.SecurityToken, c.SecurityToken)
	cred.CanExpire = c.CanExpire
	cred.Expires = c.Expires
	return &cred
}

// LoadDefaultUCloudConfig is the default loader to load config
func LoadDefaultUCloudConfig() (ConfigProvider, error) {
	cfg := newConfig()

	if err := cfg.loadEnv(); err != nil {
		return nil, fmt.Errorf("error on loading env, %s", err)
	}

	if err := cfg.loadFileIfExist(); err != nil {
		return nil, fmt.Errorf("error on loading shared config file, %s", err)
	}

	return cfg, nil
}

func setStringify(p *string, s string) {
	if p != nil && len(s) != 0 {
		*p = s
	}
}
