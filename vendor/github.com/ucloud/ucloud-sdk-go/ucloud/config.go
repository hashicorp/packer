package ucloud

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

// Config is the config of ucloud sdk, use for setting up client
type Config struct {
	// Region is the region of backend service
	// See also <https://docs.ucloud.cn/api/summary/regionlist> ...
	Region string `default:""`

	// Zone is the zone of backend service
	// See also <https://docs.ucloud.cn/api/summary/regionlist> ...
	Zone string `default:""`

	// ProjectId is the unique identify of project, used for organize resources,
	// Most of resources should belong to a project.
	// Sub-Account must have an project id.
	// See also <https://docs.ucloud.cn/api/summary/get_project_list> ...
	ProjectId string `default:""`

	// BaseUrl is the url of backend api
	// See also <doc link> ...
	BaseUrl string `default:"https://api.ucloud.cn"`

	// UserAgent is an attribute for sdk client, used for distinguish who is using sdk.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	// It will be appended to the end of sdk user-agent.
	// eg. "Terraform/0.10.1" -> "GO/1.9.1 GO-SDK/0.1.0 Terraform/0.10.1"
	// NOTE: it will conflict with the User-Agent of HTTPHeaders
	UserAgent string `default:""`

	// Timeout is timeout for every request.
	Timeout time.Duration `default:"30s"`

	// MaxRetries is the number of max retry times.
	// Set MaxRetries more than 0 to enable auto-retry for network and service availability problem
	// if auto-retry is enabled, it will enable default retry policy using exponential backoff.
	MaxRetries int `default:"0"`

	// LogLevel is equal to logrus level,
	// if logLevel not be set, use INFO level as default.
	LogLevel log.Level `default:"log.InfoLevel"`

	// Logging level by action, used to filter logging messages by action
	// use SetActionLevel() and GetActionLevel() to modify
	actionLoggingLevels map[string]log.Level
}

// NewConfig will return a new client config with default options.
func NewConfig() Config {
	cfg := Config{
		BaseUrl:             "https://api.ucloud.cn",
		Timeout:             30 * time.Second,
		MaxRetries:          0,
		LogLevel:            log.WarnLevel,
		actionLoggingLevels: make(map[string]log.Level),
	}
	return cfg
}

// GetActionLevel will return the log level of action
func (c *Config) GetActionLevel(action string) log.Level {
	if level, ok := c.actionLoggingLevels[action]; ok {
		return level
	}
	return c.LogLevel
}

// SetActionLevel will return the log level of action
func (c *Config) SetActionLevel(action string, level log.Level) {
	c.actionLoggingLevels[action] = level
}
