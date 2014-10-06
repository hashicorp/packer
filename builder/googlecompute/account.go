package googlecompute

import (
	"encoding/json"
	"os"
)

// accountFile represents the structure of the account file JSON file.
type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}

// clientSecretsFile represents the structure of the client secrets JSON file.
type clientSecretsFile struct {
	Web struct {
		AuthURI     string `json:"auth_uri"`
		ClientEmail string `json:"client_email"`
		ClientId    string `json:"client_id"`
		TokenURI    string `json:"token_uri"`
	}
}

func loadJSON(result interface{}, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	return dec.Decode(result)
}
