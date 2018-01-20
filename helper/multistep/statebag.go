package multistep

import (
	"context"
	"sync"
)

// Add context to state bag to prevent changing step signature

// StateBag holds the state that is used by the Runner and Steps. The
// StateBag implementation must be safe for concurrent access.
type StateBag interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
	Put(string, interface{})
	Context() context.Context
	WithContext(context.Context) StateBag
}

// BasicStateBag implements StateBag by using a normal map underneath
// protected by a RWMutex.
type BasicStateBag struct {
	data map[string]interface{}
	l    sync.RWMutex
	ctx  context.Context
}

func NewBasicStateBag() *BasicStateBag {
	b := new(BasicStateBag)
	b.data = make(map[string]interface{})
	return b
}

func (b *BasicStateBag) Get(k string) interface{} {
	result, _ := b.GetOk(k)
	return result
}

func (b *BasicStateBag) GetOk(k string) (interface{}, bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	result, ok := b.data[k]
	return result, ok
}

func (b *BasicStateBag) Put(k string, v interface{}) {
	b.l.Lock()
	defer b.l.Unlock()

	// Write the data
	b.data[k] = v
}

func (b *BasicStateBag) Context() context.Context {
	if b.ctx != nil {
		return b.ctx
	}
	return context.Background()
}

// WithContext returns a copy of BasicStateBag with the provided context
// We copy the state bag
func (b *BasicStateBag) WithContext(ctx context.Context) *BasicStateBag {
	if ctx == nil {
		panic("nil context")
	}
	// read lock because copying is a read operation
	b.l.RLock()
	defer b.l.RUnlock()

	b2 := NewBasicStateBag()

	for k, v := range b.data {
		b2.data[k] = v
	}
	b2.ctx = ctx
	return b2
}
