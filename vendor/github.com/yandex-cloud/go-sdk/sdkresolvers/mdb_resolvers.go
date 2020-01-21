package sdkresolvers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	redis "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func PostgreSQLClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &postgreSQLClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

type postgreSQLClusterResolver struct {
	BaseNameResolver
}

func (r *postgreSQLClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.MDB().PostgreSQL().Cluster().List(ctx, &postgresql.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName(resp.GetClusters(), err)
}

func MongoDBClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &mongodbClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

type mongodbClusterResolver struct {
	BaseNameResolver
}

func (r *mongodbClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.MDB().MongoDB().Cluster().List(ctx, &mongodb.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName(resp.GetClusters(), err)
}

func ClickhouseClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &clickhouseClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

type clickhouseClusterResolver struct {
	BaseNameResolver
}

func (r *clickhouseClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.MDB().Clickhouse().Cluster().List(ctx, &clickhouse.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName(resp.GetClusters(), err)
}

func RedisClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &redisClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

type redisClusterResolver struct {
	BaseNameResolver
}

func (r *redisClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.MDB().Redis().Cluster().List(ctx, &redis.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName(resp.GetClusters(), err)
}

func MySQLClusterResolver(name string, opts ...ResolveOption) ycsdk.Resolver {
	return &mySQLClusterResolver{
		BaseNameResolver: NewBaseNameResolver(name, "cluster", opts...),
	}
}

type mySQLClusterResolver struct {
	BaseNameResolver
}

func (r *mySQLClusterResolver) Run(ctx context.Context, sdk *ycsdk.SDK, opts ...grpc.CallOption) error {
	err := r.ensureFolderID()
	if err != nil {
		return err
	}

	resp, err := sdk.MDB().MySQL().Cluster().List(ctx, &mysql.ListClustersRequest{
		FolderId: r.FolderID(),
		Filter:   CreateResolverFilter("name", r.Name),
		PageSize: DefaultResolverPageSize,
	})
	return r.findName(resp.GetClusters(), err)
}
