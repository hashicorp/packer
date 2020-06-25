package external

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
)

func userHomeDir() string {
	//get user home directory
	usr, err := user.Current()
	if err != nil {
		return "~"
	}
	return usr.HomeDir
}

func loadJSONFile(path string, p interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, p)
	if err != nil {
		return err
	}

	return nil
}
