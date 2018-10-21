package cvm

import (
	"fmt"
	"github.com/hashicorp/packer/packer"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"log"
	"sort"
	"strings"
)

type Artifact struct {
	TencentCloudImages map[string]string
	BuilderIdValue     string
	Client             *cvm.Client
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
	return fmt.Sprintf("Tencentcloud images(%s) were created:\n\n", strings.Join(parts, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	for region, imageId := range a.TencentCloudImages {
		log.Printf("Delete tencentcloud image ID(%s) from region(%s)", imageId, region)

		describeReq := cvm.NewDescribeImagesRequest()
		describeReq.ImageIds = []*string{&imageId}

		describeResp, err := a.Client.DescribeImages(describeReq)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if *describeResp.Response.TotalCount == 0 {
			errors = append(errors, fmt.Errorf(
				"describe images failed, region(%s) ImageId(%s)", region, imageId))
		}

		describeShareReq := cvm.NewDescribeImageSharePermissionRequest()
		describeShareReq.ImageId = &imageId

		describeShareResp, err := a.Client.DescribeImageSharePermission(describeShareReq)
		var shareAccountIds []*string = nil
		if err != nil {
			errors = append(errors, err)
		} else {
			for _, sharePermission := range describeShareResp.Response.SharePermissionSet {
				shareAccountIds = append(shareAccountIds, sharePermission.AccountId)
			}
		}

		if shareAccountIds != nil && len(shareAccountIds) != 0 {
			cancelShareReq := cvm.NewModifyImageSharePermissionRequest()
			cancelShareReq.ImageId = &imageId
			cancelShareReq.AccountIds = shareAccountIds
			CANCEL := "CANCEL"
			cancelShareReq.Permission = &CANCEL
			_, err := a.Client.ModifyImageSharePermission(cancelShareReq)
			if err != nil {
				errors = append(errors, err)
			}
		}

		deleteReq := cvm.NewDeleteImagesRequest()
		deleteReq.ImageIds = []*string{&imageId}

		_, err = a.Client.DeleteImages(deleteReq)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 1 {
		return errors[0]
	} else if len(errors) > 1 {
		return &packer.MultiError{Errors: errors}
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
