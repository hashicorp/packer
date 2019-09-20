// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type userAccountByLoginResolver struct {
	BaseResolver
}

func UserAccountByLoginResolver(login string, opts ...ResolveOption) ycsdk.Resolver {
	return &userAccountByLoginResolver{
		BaseResolver: NewBaseResolver(login, opts...),
	}
}

func (r *userAccountByLoginResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	return r.Set(sdk.IAM().YandexPassportUserAccount().GetByLogin(ctx, &iam.GetUserAccountByLoginRequest{
		Login: r.Name,
	}, opts...))
}
