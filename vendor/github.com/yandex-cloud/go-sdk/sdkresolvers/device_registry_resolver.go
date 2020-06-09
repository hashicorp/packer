package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type deviceRegistryResolver struct {
	BaseNameResolver
}

func DeviceRegistryResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &deviceRegistryResolver{
		BaseNameResolver: NewBaseNameResolver(name, "registry", opts...),
	}
}

func (r *deviceRegistryResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	nextPageToken := ""
	var regs []*iot.Registry
	for ok := true; ok; ok = len(nextPageToken) > 0 {
		resp, err := sdk.IoT().Devices().Registry().List(ctx, &iot.ListRegistriesRequest{
			FolderId:  r.FolderID(),
			PageSize:  DefaultResolverPageSize,
			PageToken: nextPageToken,
		}, opts...)
		if err != nil {
			return err
		}
		nextPageToken = resp.GetNextPageToken()
		regs = append(regs, resp.GetRegistries()...)
	}

	return r.findName(regs, err)
}
