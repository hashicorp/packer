/*
Package rpc contains the implementation of the remote procedure call code that
the Packer core uses to communicate with packer plugins. As a plugin maintainer,
you are unlikely to need to directly import or use this package, but it
underpins the packer server that all plugins must implement.
*/
package rpc

import "encoding/gob"

func init() {
	gob.Register(new(map[string]string))
	gob.Register(make([]interface{}, 0))
	gob.Register(new(BasicError))
}
