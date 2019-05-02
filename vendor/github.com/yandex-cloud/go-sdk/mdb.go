// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package ycsdk

import (
	"github.com/yandex-cloud/go-sdk/gen/mdb/clickhouse"
	"github.com/yandex-cloud/go-sdk/gen/mdb/mongodb"
	"github.com/yandex-cloud/go-sdk/gen/mdb/postgresql"
	"github.com/yandex-cloud/go-sdk/gen/mdb/redis"
)

const (
	MDBMongoDBServiceID    Endpoint = "managed-mongodb"
	MDBClickhouseServiceID Endpoint = "managed-clickhouse"
	MDBPostgreSQLServiceID Endpoint = "managed-postgresql"
	MDBRedisServiceID      Endpoint = "managed-redis"
)

type MDB struct {
	sdk *SDK
}

func (m *MDB) PostgreSQL() *postgresql.PostgreSQL {
	return postgresql.NewPostgreSQL(m.sdk.getConn(MDBPostgreSQLServiceID))
}

func (m *MDB) MongoDB() *mongodb.MongoDB {
	return mongodb.NewMongoDB(m.sdk.getConn(MDBMongoDBServiceID))
}

func (m *MDB) Clickhouse() *clickhouse.Clickhouse {
	return clickhouse.NewClickhouse(m.sdk.getConn(MDBClickhouseServiceID))
}

func (m *MDB) Redis() *redis.Redis {
	return redis.NewRedis(m.sdk.getConn(MDBRedisServiceID))
}
