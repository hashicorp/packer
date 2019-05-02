package command

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/packer/packer"
)

// NewParallelTestBuilder will return a New ParallelTestBuilder whose first run
// will lock until unlockOnce is closed and that will unlock after `runs`
// builds
func NewParallelTestBuilder(runs int) *ParallelTestBuilder {
	pb := &ParallelTestBuilder{
		unlockOnce: make(chan interface{}),
	}
	pb.wg.Add(runs)
	return pb
}

// The ParallelTestBuilder's first run will lock
type ParallelTestBuilder struct {
	once       sync.Once
	unlockOnce chan interface{}

	wg sync.WaitGroup
}

func (b *ParallelTestBuilder) Prepare(raws ...interface{}) ([]string, error) {
	return nil, nil
}

func (b *ParallelTestBuilder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	b.once.Do(func() {
		ui.Say("locking build")
		<-b.unlockOnce
		b.wg.Add(1) // avoid a panic
	})

	ui.Say("building")
	b.wg.Done()
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
	// testfile that running 6 builds, with first one locks 'forever', other
	// builds should go through.
	b := NewParallelTestBuilder(5)

	c := &BuildCommand{
		Meta: testMetaParallel(t, b),
	}

	args := []string{
		fmt.Sprintf("-parallel=2"),
		filepath.Join(testFixture("parallel"), "template.json"),
	}

	wg := errgroup.Group{}

	wg.Go(func() error {
		if code := c.Run(args); code != 0 {
			fatalCommand(t, c.Meta)
		}
		return nil
	})

	b.wg.Wait()         // ran 5 times
	close(b.unlockOnce) // unlock locking one
	wg.Wait()           // wait for termination
}
