package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type repositoryResolver struct {
	BaseResolver
}

func RepositoryResolver(login string, opts ...ResolveOption) ycsdk.Resolver {
	return &repositoryResolver{
		BaseResolver: NewBaseResolver(login, opts...),
	}
}

func (r *repositoryResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	return r.Set(sdk.ContainerRegistry().Repository().GetByName(ctx, &containerregistry.GetRepositoryByNameRequest{
		RepositoryName: r.Name,
	}, opts...))
}
