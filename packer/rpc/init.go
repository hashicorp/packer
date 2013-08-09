package rpc

import "encoding/gob"

func init() {
	gob.Register(new(map[string]interface{}))
	gob.Register(new(map[string]string))
	gob.Register(make([]interface{}, 0))
	gob.Register(new(BasicError))
}
