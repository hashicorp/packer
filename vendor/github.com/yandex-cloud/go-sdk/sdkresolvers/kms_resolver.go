// Copyright (c) 2019 Yandex LLC. All rights reserved.
// Author: Shavkat Husanov <khusanov@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type symmetricKeyResolver struct {
	BaseNameResolver
}

func SymmetricKeyResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &symmetricKeyResolver{
		BaseNameResolver: NewBaseNameResolver(name, "symmetric-key", opts...),
	}
}

func (r *symmetricKeyResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}
	res := []*kms.SymmetricKey{}
	nextPageToken := ""
	for ok := true; ok; ok = len(nextPageToken) > 0 {
		resp, err := sdk.KMS().SymmetricKey().List(ctx, &kms.ListSymmetricKeysRequest{
			FolderId:  r.FolderID(),
			PageSize:  DefaultResolverPageSize,
			PageToken: nextPageToken,
		}, opts...)
		if err != nil {
			return err
		}
		nextPageToken = resp.GetNextPageToken()
		res = append(res, resp.GetKeys()...)
	}
	return r.findName(res, err)
}
