package googlecompute

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// accountFile represents the structure of the account file JSON file.
type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}

func parseJSON(result interface{}, text string) error {
	r := strings.NewReader(text)
	dec := json.NewDecoder(r)
	return dec.Decode(result)
}

func processAccountFile(account_file *accountFile, text string) error {
	// Assume text is a JSON string
	if err := parseJSON(account_file, text); err != nil {
		// If text was not JSON, assume it is a file path instead
		if _, err := os.Stat(text); os.IsNotExist(err) {
			return fmt.Errorf(
				"account_file path does not exist: %s",
				text)
		}

		b, err := ioutil.ReadFile(text)
		if err != nil {
			return fmt.Errorf(
				"Error reading account_file from path '%s': %s",
				text, err)
		}

		contents := string(b)

		if err := parseJSON(account_file, contents); err != nil {
			return fmt.Errorf(
				"Error parsing account file '%s': %s",
				contents, err)
		}
	}

	return nil
}
