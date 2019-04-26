package packer

import (
	"bytes"
	"io/ioutil"
	"testing"

	"golang.org/x/sync/errgroup"
)

// The following tests rarelly just happen. So we run them 100 times.

func TestProgressTracking_open_close(t *testing.T) {
	var bar *uiProgressBar

	tracker := bar.TrackProgress("1,", 1, 42, ioutil.NopCloser(nil))
	tracker.Close()

	tracker = bar.TrackProgress("2,", 1, 42, ioutil.NopCloser(nil))
	tracker.Close()
}

func TestProgressTracking_multi_open_close(t *testing.T) {
	var bar *uiProgressBar
	g := errgroup.Group{}

	for i := 0; i < 100; i++ {
		g.Go(func() error {
			tracker := bar.TrackProgress("file,", 1, 42, ioutil.NopCloser(nil))
			return tracker.Close()
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestProgressTracking_races(t *testing.T) {
	var bar *uiProgressBar
	g := errgroup.Group{}

	for i := 0; i < 100; i++ {
		g.Go(func() error {
			txt := []byte("foobarbaz dolores")
			b := bytes.NewReader(txt)
			tracker := bar.TrackProgress("file,", 1, 42, ioutil.NopCloser(b))

			for i := 0; i < 42; i++ {
				tracker.Read([]byte("i"))
			}
			return tracker.Close()
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}
