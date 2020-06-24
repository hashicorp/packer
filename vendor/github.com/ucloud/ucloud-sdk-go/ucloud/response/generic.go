package response

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type GenericResponse interface {
	Common

	SetPayload(m map[string]interface{}) error
	GetPayload() map[string]interface{}
	Unmarshal(interface{}) error
}
type BaseGenericResponse struct {
	CommonBase

	payload map[string]interface{}
}

func (r *BaseGenericResponse) SetPayload(m map[string]interface{}) error {
	if m["Message"] != nil && reflect.ValueOf(m["Message"]).Type().Kind() != reflect.String {
		return fmt.Errorf("response SetPayload error, the Message must set a String value")
	}
	if m["Action"] != nil && reflect.ValueOf(m["Action"]).Type().Kind() != reflect.String {
		return fmt.Errorf("response SetPayload error, the Action must set a String value")
	}
	if m["RetCode"] != nil && reflect.ValueOf(m["RetCode"]).Type().Kind() != reflect.Float64 {
		return fmt.Errorf("response SetPayload error, the RetCode must set a Float64 value")
	}
	r.payload = m
	return nil
}

func (r BaseGenericResponse) GetPayload() map[string]interface{} {
	m := make(map[string]interface{})

	if len(r.CommonBase.GetMessage()) != 0 {
		m["Message"] = r.CommonBase.GetMessage()
	}

	if len(r.CommonBase.GetAction()) != 0 {
		m["Action"] = r.CommonBase.GetAction()
	}

	m["RetCode"] = r.CommonBase.GetRetCode()

	for k, v := range r.payload {
		m[k] = v
	}

	return m
}

func (r *BaseGenericResponse) GetAction() string {
	if r.payload["Action"] != nil {
		return r.payload["Action"].(string)
	}

	return r.CommonBase.GetAction()
}

func (r *BaseGenericResponse) GetMessage() string {
	if r.payload["Message"] != nil {
		return r.payload["Message"].(string)
	}

	return r.CommonBase.GetMessage()
}

func (r *BaseGenericResponse) GetRetCode() int {
	if r.payload["RetCode"] != nil {
		return int(r.payload["RetCode"].(float64))
	}

	return r.CommonBase.GetRetCode()
}

func (r BaseGenericResponse) Unmarshal(resp interface{}) error {
	body, err := json.Marshal(r.GetPayload())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}
	return nil
}
