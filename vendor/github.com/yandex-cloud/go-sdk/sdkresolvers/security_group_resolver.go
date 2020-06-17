package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type securityGroupResolver struct {
	BaseNameResolver
}

func SecurityGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &securityGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "security_group", opts...),
	}
}

func (r *securityGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.VPC().SecurityGroup().List(ctx, &vpc.ListSecurityGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetSecurityGroups(), err)
}
