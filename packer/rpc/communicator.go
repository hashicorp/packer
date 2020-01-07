package rpc

import (
	"context"
	"encoding/gob"
	"io"
	"log"
	"net/rpc"
	"os"
	"sync"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Communicator where the communicator is actually
// executed over an RPC connection.
type communicator struct {
	commonClient
}

// CommunicatorServer wraps a packer.Communicator implementation and makes
// it exportable as part of a Golang RPC server.
type CommunicatorServer struct {
	commonServer
	c packer.Communicator
}

type CommandFinished struct {
	ExitStatus int
}

type CommunicatorStartArgs struct {
	Command          string
	StdinStreamId    uint32
	StdoutStreamId   uint32
	StderrStreamId   uint32
	ResponseStreamId uint32
}

type CommunicatorDownloadArgs struct {
	Path           string
	WriterStreamId uint32
}

type CommunicatorUploadArgs struct {
	Path           string
	ReaderStreamId uint32
	FileInfo       *fileInfo
}

type CommunicatorUploadDirArgs struct {
	Dst     string
	Src     string
	Exclude []string
}

type CommunicatorDownloadDirArgs struct {
	Dst     string
	Src     string
	Exclude []string
}

func Communicator(client *rpc.Client) *communicator {
	return &communicator{
		commonClient: commonClient{
			client:   client,
			endpoint: DefaultCommunicatorEndpoint,
		},
	}
}

func (c *communicator) Start(ctx context.Context, cmd *packer.RemoteCmd) (err error) {
	var args CommunicatorStartArgs
	args.Command = cmd.Command

	var wg sync.WaitGroup

	if cmd.Stdin != nil {
		args.StdinStreamId = c.mux.NextId()
		go func() {
			serveSingleCopy("stdin", c.mux, args.StdinStreamId, nil, cmd.Stdin)
		}()
	}

	if cmd.Stdout != nil {
		wg.Add(1)
		args.StdoutStreamId = c.mux.NextId()
		go func() {
			defer wg.Done()
			serveSingleCopy("stdout", c.mux, args.StdoutStreamId, cmd.Stdout, nil)
		}()
	}

	if cmd.Stderr != nil {
		wg.Add(1)
		args.StderrStreamId = c.mux.NextId()
		go func() {
			defer wg.Done()
			serveSingleCopy("stderr", c.mux, args.StderrStreamId, cmd.Stderr, nil)
		}()
	}

	responseStreamId := c.mux.NextId()
	args.ResponseStreamId = responseStreamId

	go func() {
		conn, err := c.mux.Accept(responseStreamId)
		wg.Wait()
		if err != nil {
			log.Printf("[ERR] Error accepting response stream %d: %s",
				responseStreamId, err)
			cmd.SetExited(123)
			return
		}
		defer conn.Close()

		var finished CommandFinished
		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(&finished); err != nil {
			log.Printf("[ERR] Error decoding response stream %d: %s",
				responseStreamId, err)
			cmd.SetExited(123)
			return
		}

		log.Printf("[INFO] RPC client: Communicator ended with: %d", finished.ExitStatus)
		cmd.SetExited(finished.ExitStatus)
	}()

	err = c.client.Call(c.endpoint+".Start", &args, new(interface{}))
	return
}

func (c *communicator) Upload(path string, r io.Reader, fi *os.FileInfo) (err error) {
	// Pipe the reader through to the connection
	streamId := c.mux.NextId()
	go serveSingleCopy("uploadData", c.mux, streamId, nil, r)

	args := CommunicatorUploadArgs{
		Path:           path,
		ReaderStreamId: streamId,
	}

	if fi != nil {
		args.FileInfo = NewFileInfo(*fi)
	}

	err = c.client.Call(c.endpoint+".Upload", &args, new(interface{}))
	return
}

func (c *communicator) UploadDir(dst string, src string, exclude []string) error {
	args := &CommunicatorUploadDirArgs{
		Dst:     dst,
		Src:     src,
		Exclude: exclude,
	}

	var reply error
	err := c.client.Call(c.endpoint+".UploadDir", args, &reply)
	if err == nil {
		err = reply
	}

	return err
}

func (c *communicator) DownloadDir(src string, dst string, exclude []string) error {
	args := &CommunicatorDownloadDirArgs{
		Dst:     dst,
		Src:     src,
		Exclude: exclude,
	}

	var reply error
	err := c.client.Call(c.endpoint+".DownloadDir", args, &reply)
	if err == nil {
		err = reply
	}

	return err
}

func (c *communicator) Download(path string, w io.Writer) (err error) {
	// Serve a single connection and a single copy
	streamId := c.mux.NextId()

	waitServer := make(chan struct{})
	go func() {
		serveSingleCopy("downloadWriter", c.mux, streamId, w, nil)
		close(waitServer)
	}()

	args := CommunicatorDownloadArgs{
		Path:           path,
		WriterStreamId: streamId,
	}

	// Start sending data to the RPC server
	err = c.client.Call(c.endpoint+".Download", &args, new(interface{}))

	// Wait for the RPC server to finish receiving the data before we return
	<-waitServer

	return
}

func (c *CommunicatorServer) Start(args *CommunicatorStartArgs, reply *interface{}) error {
	ctx := context.TODO()

	// Build the RemoteCmd on this side so that it all pipes over
	// to the remote side.
	var cmd packer.RemoteCmd
	cmd.Command = args.Command

	// Create a channel to signal we're done so that we can close
	// our stdin/stdout/stderr streams
	toClose := make([]io.Closer, 0)
	doneCh := make(chan struct{})
	go func() {
		<-doneCh
		for _, conn := range toClose {
			defer conn.Close()
		}
	}()

	if args.StdinStreamId > 0 {
		conn, err := c.mux.Dial(args.StdinStreamId)
		if err != nil {
			close(doneCh)
			return NewBasicError(err)
		}

		toClose = append(toClose, conn)
		cmd.Stdin = conn
	}

	if args.StdoutStreamId > 0 {
		conn, err := c.mux.Dial(args.StdoutStreamId)
		if err != nil {
			close(doneCh)
			return NewBasicError(err)
		}

		toClose = append(toClose, conn)
		cmd.Stdout = conn
	}

	if args.StderrStreamId > 0 {
		conn, err := c.mux.Dial(args.StderrStreamId)
		if err != nil {
			close(doneCh)
			return NewBasicError(err)
		}

		toClose = append(toClose, conn)
		cmd.Stderr = conn
	}

	// Connect to the response address so we can write our result to it
	// when ready.
	responseC, err := c.mux.Dial(args.ResponseStreamId)
	if err != nil {
		close(doneCh)
		return NewBasicError(err)
	}
	responseWriter := gob.NewEncoder(responseC)

	// Start the actual command
	err = c.c.Start(ctx, &cmd)
	if err != nil {
		close(doneCh)
		return NewBasicError(err)
	}

	// Start a goroutine to spin and wait for the process to actual
	// exit. When it does, report it back to caller...
	go func() {
		defer close(doneCh)
		defer responseC.Close()
		cmd.Wait()
		log.Printf("[INFO] RPC endpoint: Communicator ended with: %d", cmd.ExitStatus())
		responseWriter.Encode(&CommandFinished{cmd.ExitStatus()})
	}()

	return nil
}

func (c *CommunicatorServer) Upload(args *CommunicatorUploadArgs, reply *interface{}) (err error) {
	readerC, err := c.mux.Dial(args.ReaderStreamId)
	if err != nil {
		return
	}
	defer readerC.Close()

	var fi *os.FileInfo
	if args.FileInfo != nil {
		fi = new(os.FileInfo)
		*fi = *args.FileInfo
	}
	err = c.c.Upload(args.Path, readerC, fi)
	return
}

func (c *CommunicatorServer) UploadDir(args *CommunicatorUploadDirArgs, reply *error) error {
	return c.c.UploadDir(args.Dst, args.Src, args.Exclude)
}

func (c *CommunicatorServer) DownloadDir(args *CommunicatorUploadDirArgs, reply *error) error {
	return c.c.DownloadDir(args.Src, args.Dst, args.Exclude)
}

func (c *CommunicatorServer) Download(args *CommunicatorDownloadArgs, reply *interface{}) (err error) {
	writerC, err := c.mux.Dial(args.WriterStreamId)
	if err != nil {
		return
	}
	defer writerC.Close()

	err = c.c.Download(args.Path, writerC)
	return
}

func serveSingleCopy(name string, mux *muxBroker, id uint32, dst io.Writer, src io.Reader) {
	conn, err := mux.Accept(id)
	if err != nil {
		log.Printf("[ERR] '%s' accept error: %s", name, err)
		return
	}

	// Be sure to close the connection after we're done copying so
	// that an EOF will successfully be sent to the remote side
	defer conn.Close()

	// The connection is the destination/source that is nil
	if dst == nil {
		dst = conn
	} else {
		src = conn
	}

	written, err := io.Copy(dst, src)
	log.Printf("[INFO] %d bytes written for '%s'", written, name)
	if err != nil {
		log.Printf("[ERR] '%s' copy error: %s", name, err)
	}
}
