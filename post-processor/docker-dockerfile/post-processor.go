package dockerfile

import (
	"fmt"
	"log"
	"bytes"
	"bufio"
	"text/template"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "packer.post-processor.docker-dockerfile"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	From       string
	Maintainer string            `mapstructure:"maintainer"`
	Cmd        interface{}       `mapstructure:"cmd"`
	Label      map[string]string `mapstructure:"label"`
	Expose     []string          `mapstructure:"expose"`
	Env        map[string]string `mapstructure:"env"`
	Entrypoint interface{}       `mapstructure:"entrypoint"`
	Volume     []string          `mapstructure:"volume"`
	User       string            `mapstructure:"user"`
	WorkDir    string            `mapstructure:"workdir"`

	ctx interpolate.Context
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil

}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != dockerimport.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only build with Dockerfile from Docker builder artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	p.config.From = artifact.Id()

	template_str := `FROM {{ .From }}
{{ if .Maintainer }}MAINTAINER {{ .Maintainer }}
{{ end }}{{ if .Cmd }}CMD {{ process .Cmd }}
{{ end }}{{ if .Label }}{{ range $k, $v := .Label }}LABEL "{{ $k }}"="{{ $v }}"
{{ end }}{{ end }}{{ if .Expose }}EXPOSE {{ join .Expose " " }}
{{ end }}{{ if .Env }}{{ range $k, $v := .Env }}ENV {{ $k }} {{ $v }}
{{ end }}{{ end }}{{ if .Entrypoint }}ENTRYPOINT {{ process .Entrypoint }}
{{ end }}{{ if .Volume }}VOLUME {{ process .Volume }}
{{ end }}{{ if .User }}USER {{ .User }}
{{ end }}{{ if .WorkDir }}WORKDIR {{ .WorkDir }}{{ end }}`

	dockerfile := new(bytes.Buffer)
	template_writer := bufio.NewWriter(dockerfile)

	tmpl, err := template.New("Dockerfile").Funcs(template.FuncMap{
			"process": p.processVar,
			"join":    strings.Join,
		}).Parse(template_str)
	if err != nil {
		return nil, false, err
	}
	err = tmpl.Execute(template_writer, p.config)
	if err != nil {
		return nil, false, err
	}
	template_writer.Flush()
	log.Printf("Dockerfile:\n%s", dockerfile.String())

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Ctx: &p.config.ctx, Ui: ui}
	}
	
	ui.Message("Base image ID: " + artifact.Id())
	id, err := driver.BuildImage(dockerfile)
	if err != nil {
		return nil, false, err
	}

	ui.Message("Image ID: " + id)

	// Build the artifact
	artifact = &docker.ImportArtifact{
		BuilderIdValue: BuilderId,
		Driver:         driver,
		IdValue:        id,
	}

	return artifact, true, nil
}

// Process a variable of unknown type.
func (p *PostProcessor) processVar(variable interface{}) (string, error) {
	switch t := variable.(type) {
	case []string:
		array := make([]string, 0, len(t))
		for _, item := range t {
			array = append(array, item)
		}
		res, _ := json.Marshal(array)
		return string(res), nil
	case []interface{}:
		array := make([]string, 0, len(t))
		for _, item := range t {
			array = append(array, item.(string))
		}
		res, _ := json.Marshal(array)
		return string(res), nil
	case string:
		return t, nil
	case nil:
		return "", nil
	}

	err := fmt.Errorf("Unsupported variable type: %s", reflect.TypeOf(variable))
	return "", err
}
