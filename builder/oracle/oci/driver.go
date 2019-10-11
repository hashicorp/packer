package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/core"
)

// Driver interfaces between the builder steps and the OCI SDK.
type Driver interface {
	CreateInstance(ctx context.Context, publicKey string) (string, error)
	CreateImage(ctx context.Context, id string) (core.Image, error)
	DeleteImage(ctx context.Context, id string) error
	GetInstanceIP(ctx context.Context, id string) (string, error)
	TerminateInstance(ctx context.Context, id string) error
	WaitForImageCreation(ctx context.Context, id string) error
	WaitForInstanceState(ctx context.Context, id string, waitStates []string, terminalState string) error
}
