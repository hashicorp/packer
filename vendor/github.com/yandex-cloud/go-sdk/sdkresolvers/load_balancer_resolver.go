// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Vasiliy Briginets <0x40@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	loadbalancer "github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type networkLoadBalancerResolver struct {
	BaseNameResolver
}

func NetworkLoadBalancerResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &networkLoadBalancerResolver{
		BaseNameResolver: NewBaseNameResolver(name, "network load balancer", opts...),
	}
}

func (r *networkLoadBalancerResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.LoadBalancer().NetworkLoadBalancer().List(ctx, &loadbalancer.ListNetworkLoadBalancersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetNetworkLoadBalancers(), err)
}

type targetGroupResolver struct {
	BaseNameResolver
}

func TargetGroupResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &targetGroupResolver{
		BaseNameResolver: NewBaseNameResolver(name, "target group", opts...),
	}
}

func (r *targetGroupResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.LoadBalancer().TargetGroup().List(ctx, &loadbalancer.ListTargetGroupsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetTargetGroups(), err)
}
