// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	functions "github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type functionResolver struct {
	BaseNameResolver
}

func FunctionResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &functionResolver{
		BaseNameResolver: NewBaseNameResolver(name, "function", opts...),
	}
}

func (r *functionResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Functions().Function().List(ctx, &functions.ListFunctionsRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetFunctions(), err)
}
