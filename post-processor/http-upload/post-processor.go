package httpupload

import (
	"bufio"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	DestinationURL      string `mapstructure:"destination_url"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)

	templates := map[string]*string{
		"destination_url": &p.config.DestinationURL,
	}

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}

		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	for _, path := range artifact.Files() {
		if file, err := os.Open(path); err != nil {
			return nil, false, fmt.Errorf("Failed: %s", err)
		} else {
			var destination_url = p.config.DestinationURL
			if strings.HasSuffix(destination_url, "/") {
				destination_url = strings.TrimRight(destination_url, "/")
			}
			
			if req, err := http.NewRequest("PUT", destination_url + "/" + filepath.Base(path), bufio.NewReader(file)); err != nil {
				return nil, false, fmt.Errorf("Failed: %s", err)
			} else {
				client := &http.Client{}
				ui.Message(fmt.Sprintf("Uploading file %s to %s", path, destination_url))
				if resp, err := client.Do(req); err != nil {
					return nil, false, fmt.Errorf("Failed: %s", err)
				} else {
					fmt.Printf("%s", resp)
				}

				if err := file.Close(); err != nil {
					return nil, false, fmt.Errorf("Failed: %s", err)
				}
			}
		}
	}
	return artifact, false, nil
}
