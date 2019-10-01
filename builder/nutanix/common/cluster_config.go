package common

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hashicorp/packer/template/interpolate"
	//"github.com/hashicorp/packer/common"
)

// UsernameVar - key value in json request and env variable
const UsernameVar string = "NUTANIX_USERNAME"

// PasswordVar - key value in json request and env variable
const PasswordVar string = "NUTANIX_PASSWORD"

// EndpointVar - key value in json request and env variable
const EndpointVar string = "NUTANIX_ENDPOINT"

// PortVar - key value in json request and env variable
const PortVar string = "NUTANIX_PORT"

// InsecureVar - key value in json request and env variable
const InsecureVar string = "NUTANIX_INSECURE"

// NutanixCluster is the configuration defining the nutanix builder request
type NutanixCluster struct {
	ClusterUsername *string `mapstructure:"nutanix_username,omitempty"`
	ClusterPassword *string `mapstructure:"nutanix_password,omitempty"`
	ClusterEndpoint *string `mapstructure:"nutanix_endpoint,omitempty"`
	ClusterURL      string
	ClusterPort     *int  `mapstructure:"nutanix_port,omitempty"`
	ClusterInsecure *bool `mapstructure:"nutanix_insecure,omitempty"`
}

// Prepare will validate the integrity of the cluster request for the nutanix builder
func (n *NutanixCluster) Prepare(ctx *interpolate.Context) ([]string, []error) {
	var errs []error
	var warns []string

	// Validate Cluster Username
	if n.ClusterUsername == nil || *n.ClusterUsername == "" {
		log.Println("Nutanix Username missing from configuration, retrieving from env variable")
		u := os.Getenv(UsernameVar)
		if u == "" {
			errs = append(errs, fmt.Errorf("Missing %s", UsernameVar))
		} else {
			n.ClusterUsername = &u
		}
	}
	if n.ClusterUsername != nil {
		log.Printf("%s: %s", UsernameVar, *n.ClusterUsername)
	}

	// Validate Nutanix Password
	if n.ClusterPassword == nil || *n.ClusterPassword == "" {
		log.Println("Nutanix Password is missing from configuration, retrieving from env variable")
		p := os.Getenv(PasswordVar)
		if p == "" {
			errs = append(errs, fmt.Errorf("Missing %s", PasswordVar))
		}
		n.ClusterPassword = &p
	}

	// Validate Nutanix Cluster Port
	if n.ClusterPort == nil {
		log.Println("Nutanix Port is missing from configuration, retrieving from env variable")
		p, err := strconv.Atoi(os.Getenv(PortVar))
		if err != nil {
			errs = append(errs, fmt.Errorf("Missing or invalid %s", PortVar))
		} else if p < 1 {
			errs = append(errs, fmt.Errorf("Missing %s", PortVar))
		} else {
			log.Printf("%s: %d", PortVar, p)
			n.ClusterPort = &p
		}
	}

	// Validate ClusterEndpoint
	if n.ClusterEndpoint == nil {
		log.Println("Nutanix Endpoint is missing from configuration, retrieving from env variable")
		e := os.Getenv(EndpointVar)
		n.ClusterEndpoint = &e
	}
	if n.ClusterEndpoint == nil || *n.ClusterEndpoint == "" {
		errs = append(errs, fmt.Errorf("Missing %s", EndpointVar))
	} else {
		// Make entire url from protocol, endpoint, and port
		log.Printf("%s: %s", EndpointVar, *n.ClusterEndpoint)
		n.ClusterURL = "https://" + *n.ClusterEndpoint + ":" + strconv.Itoa(*n.ClusterPort)
	}

	// Validate Insecure setting
	if n.ClusterInsecure == nil {
		log.Println("Nutanix Insecure setting is missing from configuration, retrieving from env variable")
		b, err := strconv.ParseBool(os.Getenv(InsecureVar))
		if err != nil {
			log.Println("Nutanix Insecure setting is not set, defaulting to 'false'")
			// Default to not allow insecure setting
			b = false
		}
		n.ClusterInsecure = &b
	}

	return warns, errs
}
