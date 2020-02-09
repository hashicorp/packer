// Copyright (c) 2018 Yandex LLC. All rights reserved.

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type registryResolver struct {
	BaseNameResolver
}

func RegistryResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &registryResolver{
		BaseNameResolver: NewBaseNameResolver(name, "registry", opts...),
	}
}

func (r *registryResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.ContainerRegistry().Registry().List(ctx, &containerregistry.ListRegistriesRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetRegistries(), err)
}
