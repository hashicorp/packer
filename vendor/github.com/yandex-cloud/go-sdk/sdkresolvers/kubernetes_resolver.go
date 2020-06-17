package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	k8s "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type kubernetesClusterResolver struct {
	BaseNameResolver
}

func KubernetesClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &kubernetesClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "kubernetes_cluster", opts...),
	}
}

func (r *kubernetesClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Kubernetes().Cluster().List(ctx, &k8s.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetClusters(), err)
}

type kubernetesNodeGroupResolver struct {
	BaseNameResolver
}

func KubernetesNodeGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &kubernetesNodeGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "kubernetes_node_group", opts...),
	}
}

func (r *kubernetesNodeGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Kubernetes().NodeGroup().List(ctx, &k8s.ListNodeGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetNodeGroups(), err)
}
