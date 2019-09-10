// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type networkResolver struct {
	BaseNameResolver
}

func NetworkResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &networkResolver{
		BaseNameResolver: NewBaseNameResolver(name, "network", opts...),
	}
}

func (r *networkResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.VPC().Network().List(ctx, &vpc.ListNetworksRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetNetworks(), err)
}

type subnetResolver struct {
	BaseNameResolver
}

func SubnetResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &subnetResolver{
		BaseNameResolver: NewBaseNameResolver(name, "subnet", opts...),
	}
}

func (r *subnetResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetSubnets(), err)
}
