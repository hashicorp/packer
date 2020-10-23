package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type addressResolver struct {
	BaseNameResolver
}

func AddressResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &addressResolver{
		BaseNameResolver: NewBaseNameResolver(name, "address", opts...),
	}
}

func (r *addressResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	if err := r.ensureFolderID(); err != nil {
		return err
	}

	resp, err := sdk.VPC().Address().List(ctx, &vpc.ListAddressesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetAddresses(), err)
}
