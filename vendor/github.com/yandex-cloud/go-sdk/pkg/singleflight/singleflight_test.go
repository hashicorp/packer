package singleflight

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	var g Group
	v := g.Do("key", func() interface{} {
		return "bar"
	})
	assert.Equal(t, "bar", v.(string))
}

func TestDoAsync(t *testing.T) {
	var g Group
	var calledDone sync.WaitGroup
	calledDone.Add(1)
	g.DoAsync("key", func() interface{} {
		calledDone.Done()
		return "bar"
	})
	calledDone.Wait()
}

func TestDoDupSuppress(t *testing.T) {
	var g Group
	c := make(chan string)
	var calls int32
	fn := func() interface{} {
		atomic.AddInt32(&calls, 1)
		return <-c
	}

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v := g.Do("key", fn)
			assert.Equal(t, "bar", v)
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond) // Let goroutines above block.
	c <- "bar"
	wg.Wait()
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))
}

func TestDoAsyncDupSuppress(t *testing.T) {
	var g Group
	c := make(chan string)
	var calls int32
	fn := func() interface{} {
		atomic.AddInt32(&calls, 1)
		return <-c
	}

	const n = 10
	for i := 0; i < n; i++ {
		g.DoAsync("key", fn)
	}
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			v := g.Do("key", fn)
			assert.Equal(t, "bar", v)
			wg.Done()
		}()
	}
	time.Sleep(100 * time.Millisecond)
	c <- "bar"
	wg.Wait()
	assert.EqualValues(t, 1, atomic.LoadInt32(&calls))
}

func TestCall(t *testing.T) {
	infl := Call{}
	wg := sync.WaitGroup{}
	const qty = 1000
	wg.Add(qty)
	for i := 0; i < qty; i++ {
		go func(q int) {
			infl.Do(func() interface{} {
				return fmt.Sprintf("PAYLOAD %d", q)
			})
			wg.Done()
		}(i)
	}

	wg.Wait()
}
