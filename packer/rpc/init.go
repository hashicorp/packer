package rpc

import "encoding/gob"

func init() {
	gob.Register(new(map[string]interface{}))
	gob.Register(new(BasicError))
}
