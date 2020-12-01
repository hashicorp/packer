package common

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common/client"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Artifact is an artifact implementation that contains built Managed Images or Disks.
type Artifact struct {
	// Array of the Azure resource IDs that were created.
	Resources []string

	// BuilderId is the unique ID for the builder that created this AMI
	BuilderIdValue string

	// Azure client for performing API stuff.
	AzureClientSet client.AzureClientSet

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.Resources))
	for _, resource := range a.Resources {
		parts = append(parts, strings.ToLower(resource))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	parts := make([]string, 0, len(a.Resources))
	for _, resource := range a.Resources {
		parts = append(parts, strings.ToLower(resource))
	}

	sort.Strings(parts)
	return fmt.Sprintf("Azure resources created:\n%s\n", strings.Join(parts, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	errs := make([]error, 0)

	for _, resource := range a.Resources {
		log.Printf("Deleting resource %s", resource)

		id, err := azure.ParseResourceID(resource)
		if err != nil {
			return fmt.Errorf("Unable to parse resource id (%s): %v", resource, err)
		}

		ctx := context.TODO()
		restype := strings.ToLower(fmt.Sprintf("%s/%s", id.Provider, id.ResourceType))

		switch restype {
		case "microsoft.compute/images":
			res, err := a.AzureClientSet.ImagesClient().Delete(ctx, id.ResourceGroup, id.ResourceName)
			if err != nil {
				errs = append(errs, fmt.Errorf("Unable to initiate deletion of resource (%s): %v", resource, err))
			} else {
				err := res.WaitForCompletionRef(ctx, a.AzureClientSet.PollClient())
				if err != nil {
					errs = append(errs, fmt.Errorf("Unable to complete deletion of resource (%s): %v", resource, err))
				}
			}
		default:
			errs = append(errs, fmt.Errorf("Don't know how to delete resources of type %s (%s)", resource, restype))
		}

	}

	if len(errs) > 0 {
		if len(errs) == 1 {
			return errs[0]
		} else {
			return &packersdk.MultiError{Errors: errs}
		}
	}

	return nil
}
