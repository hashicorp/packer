package request

import (
	"fmt"
	"reflect"
)

type GenericRequest interface {
	Common

	SetPayload(m map[string]interface{}) error
	GetPayload() map[string]interface{}
}

type BaseGenericRequest struct {
	CommonBase

	payload map[string]interface{}
}

func (r *BaseGenericRequest) SetPayload(m map[string]interface{}) error {
	if m["Region"] != nil && reflect.ValueOf(m["Region"]).Type().Kind() != reflect.String {
		return fmt.Errorf("request SetPayload error, the Region must set a String value")
	}
	if m["Zone"] != nil && reflect.ValueOf(m["Zone"]).Type().Kind() != reflect.String {
		return fmt.Errorf("request SetPayload error, the Zone must set a String value")
	}
	if m["Action"] != nil && reflect.ValueOf(m["Action"]).Type().Kind() != reflect.String {
		return fmt.Errorf("request SetPayload error, the Action must set a String value")
	}
	if m["ProjectId"] != nil && reflect.ValueOf(m["ProjectId"]).Type().Kind() != reflect.String {
		return fmt.Errorf("request SetPayload error, the ProjectId must set a String value")
	}
	r.payload = m
	return nil
}

func (r BaseGenericRequest) GetPayload() map[string]interface{} {
	m := make(map[string]interface{})
	if len(r.CommonBase.GetRegion()) != 0 {
		m["Region"] = r.CommonBase.GetRegion()
	}

	if len(r.CommonBase.GetZone()) != 0 {
		m["Zone"] = r.CommonBase.GetZone()
	}

	if len(r.CommonBase.GetAction()) != 0 {
		m["Action"] = r.CommonBase.GetAction()
	}

	if len(r.CommonBase.GetProjectId()) != 0 {
		m["ProjectId"] = r.CommonBase.GetProjectId()
	}

	for k, v := range r.payload {
		m[k] = v
	}

	return m
}

func (r *BaseGenericRequest) GetAction() string {
	if r.payload["Action"] != nil {
		return r.payload["Action"].(string)
	}

	return r.CommonBase.GetAction()
}

func (r *BaseGenericRequest) GetRegion() string {
	if r.payload["Region"] != nil {
		return r.payload["Region"].(string)
	}

	return r.CommonBase.GetRegion()
}

func (r *BaseGenericRequest) GetZone() string {
	if r.payload["Zone"] != nil {
		return r.payload["Zone"].(string)
	}

	return r.CommonBase.GetZone()
}

func (r *BaseGenericRequest) GetProjectId() string {
	if r.payload["ProjectId"] != nil {
		return r.payload["ProjectId"].(string)
	}

	return r.CommonBase.GetProjectId()
}
