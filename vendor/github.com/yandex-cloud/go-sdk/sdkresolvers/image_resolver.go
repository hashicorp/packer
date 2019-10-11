// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

type imageResolver struct {
	BaseNameResolver
}

func ImageResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &imageResolver{
		BaseNameResolver: NewBaseNameResolver(name, "image", opts...),
	}
}

func (r *imageResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Compute().Image().List(ctx, &compute.ListImagesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetImages(), err)
}

type imageByFamilyResolver struct {
	BaseNameResolver
}

func ImageByFamilyResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &imageByFamilyResolver{
		BaseNameResolver: NewBaseNameResolver(name, "image from family", opts...),
	}
}

func (r *imageByFamilyResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	img, err := sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: r.FolderID(),
		Family:   r.Name,
	})
	if err != nil {
		err = sdkerrors.WithMessagef(err, "failed to find image with family \"%v\"", r.Name)
	}
	return r.Set(img, err)
}
