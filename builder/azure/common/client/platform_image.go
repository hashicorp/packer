package client

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute/computeapi"
	"github.com/Azure/go-autorest/autorest/to"
)

var platformImageRegex = regexp.MustCompile(`^[-_.a-zA-Z0-9]+:[-_.a-zA-Z0-9]+:[-_.a-zA-Z0-9]+:[-_.a-zA-Z0-9]+$`)

type VirtualMachineImagesClientAPI interface {
	computeapi.VirtualMachineImagesClientAPI
	// extensions
	GetLatest(ctx context.Context, publisher, offer, sku, location string) (*compute.VirtualMachineImageResource, error)
}

var _ VirtualMachineImagesClientAPI = virtualMachineImagesClientAPI{}

type virtualMachineImagesClientAPI struct {
	computeapi.VirtualMachineImagesClientAPI
}

func ParsePlatformImageURN(urn string) (image *PlatformImage, err error) {
	if !platformImageRegex.Match([]byte(urn)) {
		return nil, fmt.Errorf("%q is not a valid platform image specifier", urn)
	}
	parts := strings.Split(urn, ":")
	return &PlatformImage{parts[0], parts[1], parts[2], parts[3]}, nil
}

func (c virtualMachineImagesClientAPI) GetLatest(ctx context.Context, publisher, offer, sku, location string) (*compute.VirtualMachineImageResource, error) {
	result, err := c.List(ctx, location, publisher, offer, sku, "", to.Int32Ptr(1), "name desc")
	if err != nil {
		return nil, err
	}
	if result.Value == nil || len(*result.Value) == 0 {
		return nil, fmt.Errorf("%s:%s:%s:latest could not be found in location %s", publisher, offer, sku, location)
	}

	return &(*result.Value)[0], nil
}

type PlatformImage struct {
	Publisher, Offer, Sku, Version string
}

func (pi PlatformImage) URN() string {
	return fmt.Sprintf("%s:%s:%s:%s",
		pi.Publisher,
		pi.Offer,
		pi.Sku,
		pi.Version)
}
