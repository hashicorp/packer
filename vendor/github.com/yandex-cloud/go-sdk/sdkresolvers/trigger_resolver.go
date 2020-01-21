// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	triggers "github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type triggerResolver struct {
	BaseNameResolver
}

func TriggerResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &triggerResolver{
		BaseNameResolver: NewBaseNameResolver(name, "trigger", opts...),
	}
}

func (r *triggerResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.Triggers().Trigger().List(ctx, &triggers.ListTriggersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	}, opts...)
	return r.findName(resp.GetTriggers(), err)
}
