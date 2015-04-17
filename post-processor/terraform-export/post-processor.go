package terraformexport

import (
	"fmt"
	"encoding/json"
	"os"
	"strings"
	"io/ioutil"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/amazon/chroot"
	"github.com/mitchellh/packer/builder/amazon/ebs"
	"github.com/mitchellh/packer/builder/amazon/instance"
	"github.com/mitchellh/mapstructure"
)

const ArtifactStateMetadata = "awsgeneric.artifact.metadata"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Output string `mapstructure:"output"`
	Variable string `mapstructure:"variable"`
	EnablePartialMatch bool `mapstructure:"enable_partial_match"`
	AmiName string `mapstructure:"ami_name"`

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
		"output" : &p.config.Output,
		"ami_name" : &p.config.AmiName,
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

func (p *PostProcessor) metadata(artifact packer.Artifact) map[string]string {
	var metadata map[string]string
	metadataRaw := artifact.State(ArtifactStateMetadata)
	if metadataRaw != nil {
		if err := mapstructure.Decode(metadataRaw, &metadata); err != nil {
			panic(err)
		}
	}

	return metadata
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != chroot.BuilderId && artifact.BuilderId() != ebs.BuilderId && artifact.BuilderId() != instance.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from AWS builder artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	// Get the named AMIs from the AWS builder
	awsMetadata := p.metadata(artifact)

	if(awsMetadata == nil) {
		return nil, false, fmt.Errorf("No AMIs were available to process")
	}

	config := map[string]string {}

	if _, err := os.Stat(p.config.Output); err == nil {
		file, e := ioutil.ReadFile(p.config.Output)
	    if e == nil {
	        json.Unmarshal(file, &config)
	    }else {
	    	ui.Message("Unable to load existing tfvars file '" + p.config.Output + "': " + e.Error())
	    }
	}

	for sourceAmiName, amiId := range awsMetadata {
		var amiName = p.config.AmiName

		if(p.config.EnablePartialMatch) {
			if(!strings.HasPrefix(sourceAmiName, amiName)) {
				continue;
			}
		}else {
			if(amiName != sourceAmiName) {
				continue;
			}
		}

		config[p.config.Variable] = amiId
	}

	jsonString, err := json.Marshal(config)

	if(err != nil) {
		return nil, false, fmt.Errorf("Error serializing to '%s'", p.config.Output)
	}

	f, err := os.Create(p.config.Output)
    if err != nil {
        panic(err)
    }

    defer f.Close()

    _, err = f.Write(jsonString)
	if err != nil {
        panic(err)
    }
    f.Sync()

	return NewArtifact(p.config.Output, p.config.Variable), true, nil
}
