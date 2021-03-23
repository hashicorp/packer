package internal

import (
	"context"
	"net/http"
)

type submatchesKeyType struct{}

var submatchesKey submatchesKeyType

func SetSubmatches(req *http.Request, submatches []string) *http.Request {
	if len(submatches) > 0 {
		return req.WithContext(context.WithValue(req.Context(), submatchesKey, submatches))
	}
	return req
}

func GetSubmatches(req *http.Request) []string {
	sm, _ := req.Context().Value(submatchesKey).([]string)
	return sm
}
