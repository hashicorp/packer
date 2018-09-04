package packer

import (
	"github.com/cheggaaa/pb"
)

// ProgressBar allows to graphically display
// a self refreshing progress bar.
// No-op When in machine readable mode.
type ProgressBar interface {
	Start(total uint64)
	Set(current uint64)
	Finish()
}

type BasicProgressBar struct {
	*pb.ProgressBar
}

func (bpb *BasicProgressBar) Start(total uint64) {
	bpb.SetTotal64(int64(total))
	bpb.ProgressBar.Start()
}

func (bpb *BasicProgressBar) Set(current uint64) {
	bpb.ProgressBar.Set64(int64(current))
}

var _ ProgressBar = new(BasicProgressBar)

// NoopProgressBar is a silent progress bar
type NoopProgressBar struct {
}

func (bpb *NoopProgressBar) Start(_ uint64) {}
func (bpb *NoopProgressBar) Set(_ uint64)   {}
func (bpb *NoopProgressBar) Finish()        {}

var _ ProgressBar = new(NoopProgressBar)
