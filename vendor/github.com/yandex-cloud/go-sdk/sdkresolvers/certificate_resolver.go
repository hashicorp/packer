// Copyright (c) 2020 Yandex LLC. All rights reserved.
// Author: Petr Zhalybin <pjalybin@yandex-team.ru>

package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/certificatemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type certificateResolver struct {
	BaseNameResolver
}

func CertificateResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &certificateResolver{
		BaseNameResolver: NewBaseNameResolver(name, "certificate", opts...),
	}
}

func (r *certificateResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}
	res := []*certificatemanager.Certificate{}
	nextPageToken := ""
	for ok := true; ok; ok = len(nextPageToken) > 0 {
		resp, err := sdk.Certificates().Certificate().List(ctx, &certificatemanager.ListCertificatesRequest{
			FolderId:  r.FolderID(),
			PageSize:  DefaultResolverPageSize,
			PageToken: nextPageToken,
		}, opts...)
		if err != nil {
			return err
		}
		nextPageToken = resp.GetNextPageToken()
		res = append(res, resp.GetCertificates()...)
	}
	return r.findName(res, err)
}
