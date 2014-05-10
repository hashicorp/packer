package packer

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"os"
	"strconv"
	"text/template"
	"time"
)

// InitTime is the UTC time when this package was initialized. It is
// used as the timestamp for all configuration templates so that they
// match for a single build.
var InitTime time.Time

func init() {
	InitTime = time.Now().UTC()
}

// ConfigTemplate processes string data as a text/template with some common
// elements and functions available. Plugin creators should process as
// many fields as possible through this.
type ConfigTemplate struct {
	UserVars map[string]string

	root *template.Template
	i    int
}

// NewConfigTemplate creates a new configuration template processor.
func NewConfigTemplate() (*ConfigTemplate, error) {
	result := &ConfigTemplate{
		UserVars: make(map[string]string),
	}

	result.root = template.New("configTemplateRoot")
	result.root.Funcs(template.FuncMap{
		"env":       templateDisableEnv,
		"pwd":       templatePwd,
		"isotime":   templateISOTime,
		"timestamp": templateTimestamp,
		"user":      result.templateUser,
		"uuid":      templateUuid,
	})

	return result, nil
}

// Process processes a single string, compiling and executing the template.
func (t *ConfigTemplate) Process(s string, data interface{}) (string, error) {
	tpl, err := t.root.New(t.nextTemplateName()).Parse(s)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Validate the template.
func (t *ConfigTemplate) Validate(s string) error {
	root, err := t.root.Clone()
	if err != nil {
		return err
	}

	_, err = root.New("template").Parse(s)
	return err
}

// Add additional functions to the template
func (t *ConfigTemplate) Funcs(funcs template.FuncMap) {
	t.root.Funcs(funcs)
}

func (t *ConfigTemplate) nextTemplateName() string {
	name := fmt.Sprintf("tpl%d", t.i)
	t.i++
	return name
}

// User is the function exposed as "user" within the templates and
// looks up user variables.
func (t *ConfigTemplate) templateUser(n string) (string, error) {
	result, ok := t.UserVars[n]
	if !ok {
		return "", fmt.Errorf("uknown user var: %s", n)
	}

	return result, nil
}

func templateDisableEnv(n string) (string, error) {
	return "", fmt.Errorf(
		"Environmental variables can only be used as default values for user variables.")
}

func templateDisableUser(n string) (string, error) {
	return "", fmt.Errorf(
		"User variable can't be used within a default value for a user variable: %s", n)
}

func templateEnv(n string) string {
	return os.Getenv(n)
}

func templateISOTime() string {
	return InitTime.Format(time.RFC3339)
}

func templatePwd() (string, error) {
	return os.Getwd()
}

func templateTimestamp() string {
	return strconv.FormatInt(InitTime.Unix(), 10)
}

func templateUuid() string {
	return uuid.TimeOrderedUUID()
}
