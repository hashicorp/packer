package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
)

// MuxConn is a connection that can be used bi-directionally for RPC. Normally,
// Go RPC only allows client-to-server connections. This allows the client
// to actually act as a server as well.
//
// MuxConn works using a fairly dumb multiplexing technique of simply
// prefixing each message with what stream it is on along with the length
// of the data.
//
// This can likely be abstracted to N streams, but by choosing only two
// we decided to cut a lot of corners and make this easily usable for Packer.
type MuxConn struct {
	rwc     io.ReadWriteCloser
	streams map[byte]io.WriteCloser
	mu      sync.RWMutex
	wlock   sync.Mutex
}

func NewMuxConn(rwc io.ReadWriteCloser) *MuxConn {
	m := &MuxConn{
		rwc:     rwc,
		streams: make(map[byte]io.WriteCloser),
	}

	go m.loop()

	return m
}

// Close closes the underlying io.ReadWriteCloser. This will also close
// all streams that are open.
func (m *MuxConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all the streams
	for _, w := range m.streams {
		w.Close()
	}
	m.streams = make(map[byte]io.WriteCloser)

	return m.rwc.Close()
}

// Stream returns a io.ReadWriteCloser that will only read/write to the
// given stream ID. No handshake is done so if the remote end does not
// have a stream open with the same ID, then the messages will simply
// be dropped.
//
// This is one of those cases where we cut corners. Since Packer only does
// local connections, we can assume that both ends are ready at a certain
// point. In a real muxer, we'd probably want a handshake here.
func (m *MuxConn) Stream(id byte) (io.ReadWriteCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.streams[id]; ok {
		return nil, fmt.Errorf("Stream %d already exists", id)
	}

	// Create the stream object and channel where data will be sent to
	dataR, dataW := io.Pipe()
	stream := &Stream{
		id:     id,
		mux:    m,
		reader: dataR,
	}

	// Set the data channel so we can write to it.
	m.streams[id] = dataW

	return stream, nil
}

func (m *MuxConn) loop() {
	defer m.Close()

	for {
		var id byte
		var length int32

		if err := binary.Read(m.rwc, binary.BigEndian, &id); err != nil {
			log.Printf("[ERR] Error reading stream ID: %s", err)
			return
		}
		if err := binary.Read(m.rwc, binary.BigEndian, &length); err != nil {
			log.Printf("[ERR] Error reading length: %s", err)
			return
		}

		// TODO(mitchellh): probably would be better to re-use a buffer...
		data := make([]byte, length)
		if _, err := m.rwc.Read(data); err != nil {
			log.Printf("[ERR] Error reading data: %s", err)
			return
		}

		m.mu.RLock()
		w, ok := m.streams[id]
		if ok {
			// Note that if this blocks, it'll block the whole read loop.
			// Danger here... not sure how to handle it though.
			w.Write(data)
		}
		m.mu.RUnlock()
	}
}

func (m *MuxConn) write(id byte, p []byte) (int, error) {
	m.wlock.Lock()
	defer m.wlock.Unlock()

	if err := binary.Write(m.rwc, binary.BigEndian, id); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, int32(len(p))); err != nil {
		return 0, err
	}
	return m.rwc.Write(p)
}

// Stream is a single stream of data and implements io.ReadWriteCloser
type Stream struct {
	id     byte
	mux    *MuxConn
	reader io.Reader
}

func (s *Stream) Close() error {
	// Not functional yet, does it ever have to be?
	return nil
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *Stream) Write(p []byte) (int, error) {
	return s.mux.write(s.id, p)
}
