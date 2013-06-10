package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"testing"
)

type testCache struct {
	lockCalled    bool
	lockKey       string
	unlockCalled  bool
	unlockKey     string
	rlockCalled   bool
	rlockKey      string
	runlockCalled bool
	runlockKey    string
}

func (t *testCache) Lock(key string) string {
	t.lockCalled = true
	t.lockKey = key
	return "foo"
}

func (t *testCache) RLock(key string) (string, bool) {
	t.rlockCalled = true
	t.rlockKey = key
	return "foo", true
}

func (t *testCache) Unlock(key string) {
	t.unlockCalled = true
	t.unlockKey = key
}

func (t *testCache) RUnlock(key string) {
	t.runlockCalled = true
	t.runlockKey = key
}

func TestCache_Implements(t *testing.T) {
	var raw interface{}
	raw = Cache(nil)
	if _, ok := raw.(packer.Cache); !ok {
		t.Fatal("Cache must be a cache.")
	}
}

func TestCacheRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	c := new(testCache)

	// Start the server
	server := rpc.NewServer()
	RegisterCache(server, c)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	rpcClient, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")
	client := Cache(rpcClient)

	// Test Lock
	client.Lock("foo")
	assert.True(c.lockCalled, "should be called")
	assert.Equal(c.lockKey, "foo", "should have proper key")

	// Test Unlock
	client.Unlock("foo")
	assert.True(c.unlockCalled, "should be called")
	assert.Equal(c.unlockKey, "foo", "should have proper key")

	// Test RLock
	client.RLock("foo")
	assert.True(c.rlockCalled, "should be called")
	assert.Equal(c.rlockKey, "foo", "should have proper key")

	// Test RUnlock
	client.RUnlock("foo")
	assert.True(c.runlockCalled, "should be called")
	assert.Equal(c.runlockKey, "foo", "should have proper key")
}
