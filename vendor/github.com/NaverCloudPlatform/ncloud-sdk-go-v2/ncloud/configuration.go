package ncloud

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type APIKey struct {
	AccessKey string
	SecretKey string
}

type Configuration struct {
	BasePath      string            `json:"basePath,omitempty"`
	Host          string            `json:"host,omitempty"`
	Scheme        string            `json:"scheme,omitempty"`
	DefaultHeader map[string]string `json:"defaultHeader,omitempty"`
	UserAgent     string            `json:"userAgent,omitempty"`
	HTTPClient    *http.Client
	APIKey        *APIKey
}

func Keys() *APIKey {
	apiKey := &APIKey{
		AccessKey: "",
		SecretKey: "",
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return nil
	}

	if usr.HomeDir == "" {
		log.Fatal("use.HomeDir is nil")
		return nil
	}

	configureFile := filepath.Join(usr.HomeDir, ".ncloud", "configure")
	file, err := os.Open(configureFile)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, "=")
		switch strings.TrimSpace(s[0]) {
		case "ncloud_access_key_id":
			apiKey.AccessKey = strings.TrimSpace(s[1])
		case "ncloud_secret_access_key":
			apiKey.SecretKey = strings.TrimSpace(s[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil
	}

	return apiKey
}

func (c *Configuration) AddDefaultHeader(key string, value string) {
	c.DefaultHeader[key] = value
}
