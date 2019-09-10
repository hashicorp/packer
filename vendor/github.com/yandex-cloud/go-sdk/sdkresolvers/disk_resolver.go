package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type diskResolver struct {
	BaseNameResolver
}

func DiskResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &diskResolver{
		BaseNameResolver: NewBaseNameResolver(name, "disk", opts...),
	}
}

func (r *diskResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Compute().Disk().List(ctx, &compute.ListDisksRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetDisks(), err)
}
