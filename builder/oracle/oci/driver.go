package oci

import "github.com/oracle/oci-go-sdk/core"

// Driver interfaces between the builder steps and the OCI SDK.
type Driver interface {
	CreateInstance(publicKey string) (string, error)
	CreateImage(id string) (core.Image, error)
	DeleteImage(id string) error
	GetInstanceIP(id string) (string, error)
	TerminateInstance(id string) error
	WaitForImageCreation(id string) error
	WaitForInstanceState(id string, waitStates []string, terminalState string) error
}
