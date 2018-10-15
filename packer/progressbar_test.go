package packer

import (
	"sync"
	"testing"
)

func TestStackableProgressBar_race(t *testing.T) {
	bar := &StackableProgressBar{}

	start42Fn := func() { bar.Start(42) }
	finishFn := func() { bar.Finish() }
	add21 := func() { bar.Add(21) }
	// prefix := func() { bar.prefix() }

	type fields struct {
		pre   func()
		calls []func()
		post  func()
	}
	tests := []struct {
		name       string
		fields     fields
		iterations int
	}{
		{"all public", fields{nil, []func(){start42Fn, finishFn, add21, add21}, finishFn}, 300},
		{"all public", fields{start42Fn, []func(){add21, add21}, finishFn}, 300},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.iterations; i++ {
				if tt.fields.pre != nil {
					tt.fields.pre()
				}
				var startWg, endWg sync.WaitGroup
				startWg.Add(1)
				endWg.Add(len(tt.fields.calls))
				for _, call := range tt.fields.calls {
					call := call
					go func() {
						defer endWg.Done()
						startWg.Wait()
						call()
					}()
				}
				startWg.Done() // everyone is initialized, let's unlock everyone at the same time.
				// ....
				endWg.Wait() // wait for all calls to return.
				if tt.fields.post != nil {
					tt.fields.post()
				}
			}
		})
	}

}
