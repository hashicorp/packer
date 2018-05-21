package shell_local

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type ExecuteCommandTemplate struct {
	Vars    string
	Script  string
	Command string
}

func Run(ui packer.Ui, config *Config) (bool, error) {
	scripts := make([]string, len(config.Scripts))
	copy(scripts, config.Scripts)

	// If we have an inline script, then turn that into a temporary
	// shell script and use that.
	if config.Inline != nil {
		tempScriptFileName, err := createInlineScriptFile(config)
		if err != nil {
			return false, err
		}
		scripts = append(scripts, tempScriptFileName)

		// figure out what extension the file should have, and rename it.
		if config.TempfileExtension != "" {
			os.Rename(tempScriptFileName, fmt.Sprintf("%s.%s", tempScriptFileName, config.TempfileExtension))
			tempScriptFileName = fmt.Sprintf("%s.%s", tempScriptFileName, config.TempfileExtension)
		}
		defer os.Remove(tempScriptFileName)
	}

	// Create environment variables to set before executing the command
	flattenedEnvVars, err := createFlattenedEnvVars(config)
	if err != nil {
		return false, err
	}

	for _, script := range scripts {
		interpolatedCmds, err := createInterpolatedCommands(config, script, flattenedEnvVars)
		if err != nil {
			return false, err
		}
		ui.Say(fmt.Sprintf("Running local shell script: %s", script))

		comm := &Communicator{
			ExecuteCommand: interpolatedCmds,
		}

		// The remoteCmd generated here isn't actually run, but it allows us to
		// use the same interafce for the shell-local communicator as we use for
		// the other communicators; ultimately, this command is just used for
		// buffers and for reading the final exit status.
		flattenedCmd := strings.Join(interpolatedCmds, " ")
		cmd := &packer.RemoteCmd{Command: flattenedCmd}
		log.Printf("[INFO] (shell-local): starting local command: %s", flattenedCmd)

		if err := cmd.StartWithUi(comm, ui); err != nil {
			return false, fmt.Errorf(
				"Error executing script: %s\n\n"+
					"Please see output above for more information.",
				script)
		}
		if cmd.ExitStatus != 0 {
			return false, fmt.Errorf(
				"Erroneous exit code %d while executing script: %s\n\n"+
					"Please see output above for more information.",
				cmd.ExitStatus,
				script)
		}
	}

	return true, nil
}

func createInlineScriptFile(config *Config) (string, error) {
	tf, err := ioutil.TempFile("", "packer-shell")
	if err != nil {
		return "", fmt.Errorf("Error preparing shell script: %s", err)
	}
	defer tf.Close()
	// Write our contents to it
	writer := bufio.NewWriter(tf)
	if config.InlineShebang != "" {
		shebang := fmt.Sprintf("#!%s\n", config.InlineShebang)
		log.Printf("[INFO] (shell-local): Prepending inline script with %s", shebang)
		writer.WriteString(shebang)
	}
	for _, command := range config.Inline {
		if _, err := writer.WriteString(command + "\n"); err != nil {
			return "", fmt.Errorf("Error preparing shell script: %s", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing shell script: %s", err)
	}

	err = os.Chmod(tf.Name(), 0700)
	if err != nil {
		log.Printf("[ERROR] (shell-local): error modifying permissions of temp script file: %s", err.Error())
	}
	return tf.Name(), nil
}

// Generates the final command to send to the communicator, using either the
// user-provided ExecuteCommand or defaulting to something that makes sense for
// the host OS
func createInterpolatedCommands(config *Config, script string, flattenedEnvVars string) ([]string, error) {
	config.Ctx.Data = &ExecuteCommandTemplate{
		Vars:    flattenedEnvVars,
		Script:  script,
		Command: script,
	}

	interpolatedCmds := make([]string, len(config.ExecuteCommand))
	for i, cmd := range config.ExecuteCommand {
		interpolatedCmd, err := interpolate.Render(cmd, &config.Ctx)
		if err != nil {
			return nil, fmt.Errorf("Error processing command: %s", err)
		}
		interpolatedCmds[i] = interpolatedCmd
	}
	return interpolatedCmds, nil
}

func createFlattenedEnvVars(config *Config) (string, error) {
	flattened := ""
	envVars := make(map[string]string)

	// Always available Packer provided env vars
	envVars["PACKER_BUILD_NAME"] = fmt.Sprintf("%s", config.PackerBuildName)
	envVars["PACKER_BUILDER_TYPE"] = fmt.Sprintf("%s", config.PackerBuilderType)

	// Split vars into key/value components
	for _, envVar := range config.Vars {
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

	for _, key := range keys {
		flattened += fmt.Sprintf(config.EnvVarFormat, key, envVars[key])
	}
	return flattened, nil
}
