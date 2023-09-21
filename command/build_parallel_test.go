// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"

	"golang.org/x/sync/errgroup"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner/sleep"
)

// NewParallelTestBuilder will return a New ParallelTestBuilder that will
// unlock after `runs` builds
func NewParallelTestBuilder(runs int) *ParallelTestBuilder {
	pb := &ParallelTestBuilder{}
	pb.wg.Add(runs)
	return pb
}

// The ParallelTestBuilder's first run will lock
type ParallelTestBuilder struct {
	wg sync.WaitGroup
}

func (b *ParallelTestBuilder) ConfigSpec() hcldec.ObjectSpec { return nil }

func (b *ParallelTestBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return nil, nil, nil
}

func (b *ParallelTestBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	ui.Say("building")
	b.wg.Done()
	return nil, nil
}

// LockedBuilder wont run until unlock is called
type LockedBuilder struct{ unlock chan interface{} }

func (b *LockedBuilder) ConfigSpec() hcldec.ObjectSpec { return nil }

func (b *LockedBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return nil, nil, nil
}

func (b *LockedBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	ui.Say("locking build")
	select {
	case <-b.unlock:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return nil, nil
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaParallel(t *testing.T, builder *ParallelTestBuilder, locked *LockedBuilder) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: &packer.CoreConfig{
			Components: packer.ComponentFinder{
				PluginConfig: &packer.PluginConfig{
					Builders: packer.MapOfBuilder{
						"parallel-test": func() (packersdk.Builder, error) { return builder, nil },
						"file":          func() (packersdk.Builder, error) { return &file.Builder{}, nil },
						"lock":          func() (packersdk.Builder, error) { return locked, nil },
					},
					Provisioners: packer.MapOfProvisioner{
						"sleep": func() (packersdk.Provisioner, error) { return &sleep.Provisioner{}, nil },
					},
				},
			},
		},
		Ui: &packersdk.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func TestBuildParallel_1(t *testing.T) {
	// testfile has 6 builds, with first one locks 'forever', other builds
	// should go through.
	b := NewParallelTestBuilder(5)
	locked := &LockedBuilder{unlock: make(chan interface{})}

	c := &BuildCommand{
		Meta: testMetaParallel(t, b, locked),
	}

	args := []string{
		"-parallel-builds=10",
		filepath.Join(testFixture("parallel"), "1lock-5wg.json"),
	}

	wg := errgroup.Group{}

	wg.Go(func() error {
		if code := c.Run(args); code != 0 {
			fatalCommand(t, c.Meta)
		}
		return nil
	})

	b.wg.Wait()          // ran 5 times
	close(locked.unlock) // unlock locking one
	wg.Wait()            // wait for termination
}

func TestBuildParallel_2(t *testing.T) {
	// testfile has 6 builds, 2 of them lock 'forever', other builds
	// should go through.
	b := NewParallelTestBuilder(4)
	locked := &LockedBuilder{unlock: make(chan interface{})}

	c := &BuildCommand{
		Meta: testMetaParallel(t, b, locked),
	}

	args := []string{
		"-parallel-builds=3",
		filepath.Join(testFixture("parallel"), "2lock-4wg.json"),
	}

	wg := errgroup.Group{}

	wg.Go(func() error {
		if code := c.Run(args); code != 0 {
			fatalCommand(t, c.Meta)
		}
		return nil
	})

	b.wg.Wait()          // ran 4 times
	close(locked.unlock) // unlock locking one
	wg.Wait()            // wait for termination
}

func TestBuildParallel_Timeout(t *testing.T) {
	// testfile has 6 builds, 1 of them locks 'forever', one locks and times
	// out other builds should go through.
	b := NewParallelTestBuilder(4)
	locked := &LockedBuilder{unlock: make(chan interface{})}

	c := &BuildCommand{
		Meta: testMetaParallel(t, b, locked),
	}

	args := []string{
		"-parallel-builds=3",
		filepath.Join(testFixture("parallel"), "2lock-timeout.json"),
	}

	wg := errgroup.Group{}

	wg.Go(func() error {
		if code := c.Run(args); code == 0 {
			fatalCommand(t, c.Meta)
		}
		return nil
	})

	b.wg.Wait()          // ran 4 times
	close(locked.unlock) // unlock locking one
	wg.Wait()            // wait for termination
}
