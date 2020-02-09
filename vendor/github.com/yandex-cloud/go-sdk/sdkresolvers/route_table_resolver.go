// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Alexey Baranov <baranovich@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type routeTableResolver struct {
	BaseNameResolver
}

func RouteTableResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &routeTableResolver{
		BaseNameResolver: NewBaseNameResolver(name, "route_table", opts...),
	}
}

func (r *routeTableResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.VPC().RouteTable().List(ctx, &vpc.ListRouteTablesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetRouteTables(), err)
}
