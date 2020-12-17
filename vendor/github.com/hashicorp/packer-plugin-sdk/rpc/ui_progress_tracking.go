package rpc

import (
	"io"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer-plugin-sdk/random"
)

// TrackProgress starts a pair of ProgressTrackingClient and ProgressProgressTrackingServer
// that will send the size of each read bytes of stream.
// In order to track an operation on the terminal side.
func (u *Ui) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	pl := &TrackProgressParameters{
		Src:         src,
		CurrentSize: currentSize,
		TotalSize:   totalSize,
	}
	var trackingID string
	if err := u.client.Call("Ui.NewTrackProgress", pl, &trackingID); err != nil {
		log.Printf("Error in Ui.NewTrackProgress RPC call: %s", err)
		return stream
	}
	cli := &ProgressTrackingClient{
		id:     trackingID,
		client: u.client,
		stream: stream,
	}
	return cli
}

type ProgressTrackingClient struct {
	id     string
	client *rpc.Client
	stream io.ReadCloser
}

// Read will send len(b) over the wire instead of it's content
func (u *ProgressTrackingClient) Read(b []byte) (read int, err error) {
	defer func() {
		if err := u.client.Call("Ui"+u.id+".Add", read, new(interface{})); err != nil {
			log.Printf("Error in ProgressTrackingClient.Read RPC call: %s", err)
		}
	}()
	return u.stream.Read(b)
}

func (u *ProgressTrackingClient) Close() error {
	log.Printf("closing")
	if err := u.client.Call("Ui"+u.id+".Close", nil, new(interface{})); err != nil {
		log.Printf("Error in ProgressTrackingClient.Close RPC call: %s", err)
	}
	return u.stream.Close()
}

type TrackProgressParameters struct {
	Src         string
	TotalSize   int64
	CurrentSize int64
}

func (ui *UiServer) NewTrackProgress(pl *TrackProgressParameters, reply *string) error {
	// keep identifier as is for now
	srvr := &ProgressTrackingServer{
		id: *reply,
	}

	*reply = pl.Src + random.AlphaNum(6)
	srvr.stream = ui.ui.TrackProgress(pl.Src, pl.CurrentSize, pl.TotalSize, nopReadCloser{})
	err := ui.register("Ui"+*reply, srvr)
	if err != nil {
		log.Printf("failed to register ProgressTrackingServer at %s: %s", *reply, err)
		return err
	}
	return nil
}

type ProgressTrackingServer struct {
	id     string
	stream io.ReadCloser
}

func (t *ProgressTrackingServer) Add(size int, _ *interface{}) error {
	stubBytes := make([]byte, size, size)
	t.stream.Read(stubBytes)
	return nil
}

func (t *ProgressTrackingServer) Close(_, _ *interface{}) error {
	t.stream.Close()
	return nil
}

type nopReadCloser struct {
}

func (nopReadCloser) Close() error               { return nil }
func (nopReadCloser) Read(b []byte) (int, error) { return len(b), nil }
