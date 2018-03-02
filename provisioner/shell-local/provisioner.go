package shell

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	config sl.Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws...)
	if err != nil {
		return err
	}
	convertPath := false
	if len(p.config.ExecuteCommand) == 0 && runtime.GOOS == "windows" {
		convertPath = true
		p.config.ExecuteCommand = []string{
			"bash",
			"-c",
			"{{.Vars}} {{.Script}}",
		}
	}

	err = sl.Validate(&p.config)
	if err != nil {
		return err
	}

	if convertPath {
		for index, script := range p.config.Scripts {
			p.config.Scripts[index], err = convertToWindowsBashPath(script)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertToWindowsBashPath(winPath string) (string, error) {
	// get absolute path of script, and morph it into the bash path
	winAbsPath, err := filepath.Abs(winPath)
	if err != nil {
		return "", fmt.Errorf("Error converting %s to absolute path: %s", winPath, err.Error())
	}
	winAbsPath = strings.Replace(winAbsPath, "\\", "/", -1)
	winBashPath := strings.Replace(winAbsPath, "C:/", "/mnt/c/", 1)
	return winBashPath, nil
}

func (p *Provisioner) Provision(ui packer.Ui, _ packer.Communicator) error {
	_, retErr := sl.Run(ui, &p.config)
	if retErr != nil {
		return retErr
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just do nothing. When the process ends, so will our provisioner
}
