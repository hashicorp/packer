package googlecompute

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

func ProcessAccountFile(text string, iap bool) (*jwt.Config, error) {
	driverScopes := getDriverScopes(iap)
	// Assume text is a JSON string
	conf, err := google.JWTConfigFromJSON([]byte(text), driverScopes...)
	if err != nil {
		// If text was not JSON, assume it is a file path instead
		if _, err := os.Stat(text); os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"account_file path does not exist: %s",
				text)
		}
		data, err := ioutil.ReadFile(text)
		if err != nil {
			return nil, fmt.Errorf(
				"Error reading account_file from path '%s': %s",
				text, err)
		}
		conf, err = google.JWTConfigFromJSON(data, driverScopes...)
		if err != nil {
			return nil, fmt.Errorf("Error parsing account_file: %s", err)
		}
	}
	return conf, nil
}
