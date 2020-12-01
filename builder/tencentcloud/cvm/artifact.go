package cvm

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type Artifact struct {
	TencentCloudImages map[string]string
	BuilderIdValue     string
	Client             *cvm.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.TencentCloudImages))
	for region, imageId := range a.TencentCloudImages {
		parts = append(parts, fmt.Sprintf("%s:%s", region, imageId))
	}
	sort.Strings(parts)

	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	parts := make([]string, 0, len(a.TencentCloudImages))
	for region, imageId := range a.TencentCloudImages {
		parts = append(parts, fmt.Sprintf("%s: %s", region, imageId))
	}
	sort.Strings(parts)

	return fmt.Sprintf("Tencentcloud images(%s) were created.\n\n", strings.Join(parts, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	ctx := context.TODO()
	errors := make([]error, 0)

	for region, imageId := range a.TencentCloudImages {
		log.Printf("Delete tencentcloud image ID(%s) from region(%s)", imageId, region)

		describeReq := cvm.NewDescribeImagesRequest()
		describeReq.ImageIds = []*string{&imageId}
		var describeResp *cvm.DescribeImagesResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			describeResp, e = a.Client.DescribeImages(describeReq)
			return e
		})
		if err != nil {
			errors = append(errors, err)
			continue
		}

		if *describeResp.Response.TotalCount == 0 {
			errors = append(errors, fmt.Errorf(
				"describe images failed, region(%s) ImageId(%s)", region, imageId))
		}

		var shareAccountIds []*string = nil
		describeShareReq := cvm.NewDescribeImageSharePermissionRequest()
		describeShareReq.ImageId = &imageId
		var describeShareResp *cvm.DescribeImageSharePermissionResponse
		err = Retry(ctx, func(ctx context.Context) error {
			var e error
			describeShareResp, e = a.Client.DescribeImageSharePermission(describeShareReq)
			return e
		})
		if err != nil {
			errors = append(errors, err)
		} else {
			for _, sharePermission := range describeShareResp.Response.SharePermissionSet {
				shareAccountIds = append(shareAccountIds, sharePermission.AccountId)
			}
		}

		if len(shareAccountIds) != 0 {
			cancelShareReq := cvm.NewModifyImageSharePermissionRequest()
			cancelShareReq.ImageId = &imageId
			cancelShareReq.AccountIds = shareAccountIds
			CANCEL := "CANCEL"
			cancelShareReq.Permission = &CANCEL
			err := Retry(ctx, func(ctx context.Context) error {
				_, e := a.Client.ModifyImageSharePermission(cancelShareReq)
				return e
			})
			if err != nil {
				errors = append(errors, err)
			}
		}

		deleteReq := cvm.NewDeleteImagesRequest()
		deleteReq.ImageIds = []*string{&imageId}
		err = Retry(ctx, func(ctx context.Context) error {
			_, e := a.Client.DeleteImages(deleteReq)
			return e
		})
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 1 {
		return errors[0]
	} else if len(errors) > 1 {
		return &packersdk.MultiError{Errors: errors}
	} else {
		return nil
	}
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.TencentCloudImages {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
