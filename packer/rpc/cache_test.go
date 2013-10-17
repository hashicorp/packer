package rpc

import (
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
	// Create the interface to test
	c := new(testCache)

	// Start the server
	server := rpc.NewServer()
	RegisterCache(server, c)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	rpcClient, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
	client := Cache(rpcClient)

	// Test Lock
	client.Lock("foo")
	if !c.lockCalled {
		t.Fatal("should be called")
	}
	if c.lockKey != "foo" {
		t.Fatalf("bad: %s", c.lockKey)
	}

	// Test Unlock
	client.Unlock("foo")
	if !c.unlockCalled {
		t.Fatal("should be called")
	}
	if c.unlockKey != "foo" {
		t.Fatalf("bad: %s", c.unlockKey)
	}

	// Test RLock
	client.RLock("foo")
	if !c.rlockCalled {
		t.Fatal("should be called")
	}
	if c.rlockKey != "foo" {
		t.Fatalf("bad: %s", c.rlockKey)
	}

	// Test RUnlock
	client.RUnlock("foo")
	if !c.runlockCalled {
		t.Fatal("should be called")
	}
	if c.runlockKey != "foo" {
		t.Fatalf("bad: %s", c.runlockKey)
	}
}
