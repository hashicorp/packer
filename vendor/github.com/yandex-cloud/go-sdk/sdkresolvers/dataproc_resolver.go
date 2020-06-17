// Copyright (c) 2019 YANDEX LLC.

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	dataproc "github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type dataprocClusterResolver struct {
	BaseNameResolver
}

func DataprocClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &dataprocClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

func (r *dataprocClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Dataproc().Cluster().List(ctx, &dataproc.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetClusters(), err)
}

type dataprocSubclusterResolver struct {
	BaseNameResolver
}

func DataprocSubclusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &dataprocSubclusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "subcluster", opts...),
	}
}

func (r *dataprocSubclusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Dataproc().Subcluster().List(ctx, &dataproc.ListSubclustersRequest{
		ClusterId: r.opts.clusterID,
		Filter:    CreateResolverFilter("name", r.Name),
		PageSize:  DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetSubclusters(), err)
}

type dataprocJobResolver struct {
	BaseNameResolver
}

func DataprocJobResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &dataprocJobResolver{
		BaseNameResolver: NewBaseNameResolver(name, "job", opts...),
	}
}

func (r *dataprocJobResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	resp, err := sdk.Dataproc().Job().List(ctx, &dataproc.ListJobsRequest{
		ClusterId: r.opts.clusterID,
		Filter:    CreateResolverFilter("name", r.Name),
		PageSize:  DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetJobs(), err)
}
