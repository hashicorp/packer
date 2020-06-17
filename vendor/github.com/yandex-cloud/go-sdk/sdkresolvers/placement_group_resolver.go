// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Dmitry Konishchev <konishchev@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	compute "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type placementGroupResolver struct {
	BaseNameResolver
}

func PlacementGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &placementGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "placement group", opts...),
	}
}

func (r *placementGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	if err := r.ensureFolderID(); err != nil {
		return err
	}

	resp, err := sdk.Compute().PlacementGroup().List(ctx, &compute.ListPlacementGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)

	return r.findName(resp.GetPlacementGroups(), err)
}
