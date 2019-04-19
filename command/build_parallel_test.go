package command

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/packer/packer"
)

type ParallelTestBuilder struct {
	Prepared int
	Built    int
	wg       *sync.WaitGroup
	m        *sync.Mutex
}

func (b *ParallelTestBuilder) Prepare(raws ...interface{}) ([]string, error) {
	b.Prepared++
	return nil, nil
}

func (b *ParallelTestBuilder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Say(fmt.Sprintf("count: %d", b.Built))
	b.Built++
	b.wg.Done()
	b.m.Lock()
	b.m.Unlock()
	return nil, nil
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaParallel(t *testing.T, builder *ParallelTestBuilder) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: &packer.CoreConfig{
			Components: packer.ComponentFinder{
				Builder: func(n string) (packer.Builder, error) {
					switch n {
					case "parallel-test":
						return builder, nil
					default:
						panic(n)
					}
				},
			},
		},
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func TestBuildParallel(t *testing.T) {
	defer cleanup()

	m := &sync.Mutex{}
	m.Lock()
	expected := 2
	wg := &sync.WaitGroup{}
	wg.Add(expected)
	b := &ParallelTestBuilder{
		wg: wg,
		m:  m,
	}

	c := &BuildCommand{
		Meta: testMetaParallel(t, b),
	}

	args := []string{
		fmt.Sprintf("-parallel=%d", expected),
		filepath.Join(testFixture("parallel"), "template.json"),
	}

	go func(t *testing.T, c *BuildCommand) {
		if code := c.Run(args); code != 0 {
			fatalCommand(t, c.Meta)
		}
	}(t, c)

	wg.Wait()
	if b.Prepared != 6 {
		t.Errorf("Expected all builds to be prepared, was %d", b.Prepared)
	}

	if b.Built != expected {
		t.Errorf("Expected only %d running/completed builds, was %d", expected, b.Built)
	}

	m.Unlock()
	wg.Add(math.MaxInt32)
}
