// Copyright (c) 2019 YANDEX LLC.

package ycsdk

import "github.com/yandex-cloud/go-sdk/gen/functions"

const (
	FunctionServiceID Endpoint = "serverless-functions"
)

func (sdk *SDK) Functions() *functions.Function {
	return functions.NewFunction(sdk.getConn(FunctionServiceID))
}
