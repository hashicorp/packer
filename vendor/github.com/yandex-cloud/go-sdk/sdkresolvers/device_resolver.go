package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type deviceResolver struct {
	BaseNameResolver
}

func DeviceResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &deviceResolver{
		BaseNameResolver: NewBaseNameResolver(name, "device", opts...),
	}
}

func (r *deviceResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	nextPageToken := ""
	var devices []*iot.Device
	for ok := true; ok; ok = len(nextPageToken) > 0 {
		resp, err := sdk.IoT().Devices().Device().List(ctx, &iot.ListDevicesRequest{
			Id:        &iot.ListDevicesRequest_FolderId{FolderId: r.FolderID()},
			PageSize:  DefaultResolverPageSize,
			PageToken: nextPageToken,
		}, opts...)
		if err != nil {
			return err
		}
		nextPageToken = resp.GetNextPageToken()
		devices = append(devices, resp.GetDevices()...)
	}

	return r.findName(devices, err)
}
