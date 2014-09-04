package rpc

import (
	"github.com/mitchellh/packer/packer"
	"log"
	"net/rpc"
)

// An implementation of packer.Cache where the cache is actually executed
// over an RPC connection.
type cache struct {
	client *rpc.Client
}

// CacheServer wraps a packer.Cache implementation and makes it exportable
// as part of a Golang RPC server.
type CacheServer struct {
	cache packer.Cache
}

type CacheRLockResponse struct {
	Path   string
	Exists bool
}

func (c *cache) Lock(key string) (result string) {
	if err := c.client.Call("Cache.Lock", key, &result); err != nil {
		log.Printf("[ERR] Cache.Lock error: %s", err)
		return
	}

	return
}

func (c *cache) RLock(key string) (string, bool) {
	var result CacheRLockResponse
	if err := c.client.Call("Cache.RLock", key, &result); err != nil {
		log.Printf("[ERR] Cache.RLock error: %s", err)
		return "", false
	}

	return result.Path, result.Exists
}

func (c *cache) Unlock(key string) {
	if err := c.client.Call("Cache.Unlock", key, new(interface{})); err != nil {
		log.Printf("[ERR] Cache.Unlock error: %s", err)
		return
	}
}

func (c *cache) RUnlock(key string) {
	if err := c.client.Call("Cache.RUnlock", key, new(interface{})); err != nil {
		log.Printf("[ERR] Cache.RUnlock error: %s", err)
		return
	}
}

func (c *CacheServer) Lock(key string, result *string) error {
	*result = c.cache.Lock(key)
	return nil
}

func (c *CacheServer) Unlock(key string, result *interface{}) error {
	c.cache.Unlock(key)
	return nil
}

func (c *CacheServer) RLock(key string, result *CacheRLockResponse) error {
	path, exists := c.cache.RLock(key)
	*result = CacheRLockResponse{path, exists}
	return nil
}

func (c *CacheServer) RUnlock(key string, result *interface{}) error {
	c.cache.RUnlock(key)
	return nil
}
