package shell_local

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type PostProcessor struct {
	config sl.Config
}

type ExecuteCommandTemplate struct {
	Vars   string
	Script string
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws)
	if err != nil {
		return err
	}

	return sl.Validate(&p.config)
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	scripts := make([]string, len(p.config.Scripts))
	copy(scripts, p.config.Scripts)

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if p.config.Inline != nil {
		tf, err := ioutil.TempFile("", "packer-shell")
		if err != nil {
			return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		scripts = append(scripts, tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		writer.WriteString(fmt.Sprintf("#!%s\n", p.config.InlineShebang))
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return nil, false, fmt.Errorf("Error preparing shell script: %s", err)
		}

		tf.Close()
	}

	// Create environment variables to set before executing the command
	flattenedEnvVars := p.createFlattenedEnvVars()

	for _, script := range scripts {

		p.config.Ctx.Data = &ExecuteCommandTemplate{
			Vars:   flattenedEnvVars,
			Script: script,
		}

		flattenedCmd := strings.Join(p.config.ExecuteCommand, " ")
		command, err := interpolate.Render(flattenedCmd, &p.config.Ctx)
		if err != nil {
			return nil, false, fmt.Errorf("Error processing command: %s", err)
		}

		ui.Say(fmt.Sprintf("Post processing with local shell script: %s", script))

		comm := &sl.Communicator{
			Ctx:            p.config.Ctx,
			ExecuteCommand: []string{flattenedCmd},
		}

		cmd := &packer.RemoteCmd{Command: command}

		log.Printf("starting local command: %s", command)
		if err := cmd.StartWithUi(comm, ui); err != nil {
			return nil, false, fmt.Errorf(
				"Error executing script: %s\n\n"+
					"Please see output above for more information.",
				script)
		}
		if cmd.ExitStatus != 0 {
			return nil, false, fmt.Errorf(
				"Erroneous exit code %d while executing script: %s\n\n"+
					"Please see output above for more information.",
				cmd.ExitStatus,
				script)
		}
	}

	return artifact, true, nil
}

func (p *PostProcessor) createFlattenedEnvVars() (flattened string) {
	flattened = ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = fmt.Sprintf("%s", p.config.PackerBuildName)
	envVars["PACKER_BUILDER_TYPE"] = fmt.Sprintf("%s", p.config.PackerBuilderType)

	// Split vars into key/value components
	for _, envVar := range p.config.Vars {
		keyValue := strings.SplitN(envVar, "=", 2)
		// Store pair, replacing any single quotes in value so they parse
		// correctly with required environment variable format
		envVars[keyValue[0]] = strings.Replace(keyValue[1], "'", `'"'"'`, -1)
	}

	// Create a list of env var keys in sorted order
	var keys []string
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Re-assemble vars surrounding value with single quotes and flatten
	for _, key := range keys {
		flattened += fmt.Sprintf("%s='%s' ", key, envVars[key])
	}
	return
}
