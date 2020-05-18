package command

import (
    "os"
    "github.com/hashicorp/packer/packer"
    "bytes"
)

func FuzzBuild(data []byte) int {
	var out, errui bytes.Buffer
	f, err := os.Create("data.json")
	if err != nil {
		return -1
	}
	defer os.Remove("data.json")
	_, err = f.Write(data)
	if err != nil {
		return -1
	}
	c := &BuildCommand{
		Meta: Meta{
			Ui: &packer.BasicUi{
				Writer:      &out,
				ErrorWriter: &errui,
			},
		},
	}
	args := []string{"data.json"}
	if code := c.Run(args); code != 0 {
		return 0
	}
	return 1
}
