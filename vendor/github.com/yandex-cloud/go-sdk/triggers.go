// Copyright (c) 2019 YANDEX LLC.

package ycsdk

import "github.com/yandex-cloud/go-sdk/gen/triggers"

const (
	TriggerServiceID Endpoint = "serverless-triggers"
)

func (sdk *SDK) Triggers() *triggers.Trigger {
	return triggers.NewTrigger(sdk.getConn(TriggerServiceID))
}
