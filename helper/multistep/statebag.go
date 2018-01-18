package multistep

import (
	"context"
	"sync"
)

// StateBag implements StateBag by using a normal map underneath
// protected by a RWMutex.
type StateBag struct {
	data map[string]interface{}
	l    sync.RWMutex
	once sync.Once
	ctx  context.Context
}

func (b *StateBag) Context() context.Context {
	if b.ctx != nil {
		return b.ctx
	}
	return context.Background()
}

func (b *StateBag) Get(k string) interface{} {
	result, _ := b.GetOk(k)
	return result
}

func (b *StateBag) GetOk(k string) (interface{}, bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	result, ok := b.data[k]
	return result, ok
}

func (b *StateBag) Put(k string, v interface{}) {
	b.l.Lock()
	defer b.l.Unlock()

	// Make sure the map is initialized one time, on write
	b.once.Do(func() {
		b.data = make(map[string]interface{})
	})

	// Write the data
	b.data[k] = v
}
