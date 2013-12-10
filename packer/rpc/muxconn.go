package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// MuxConn is a connection that can be used bi-directionally for RPC. Normally,
// Go RPC only allows client-to-server connections. This allows the client
// to actually act as a server as well.
//
// MuxConn works using a fairly dumb multiplexing technique of simply
// framing every piece of data sent into a prefix + data format. Streams
// are established using a subset of the TCP protocol. Only a subset is
// necessary since we assume ordering on the underlying RWC.
type MuxConn struct {
	curId   uint32
	rwc     io.ReadWriteCloser
	streams map[uint32]*Stream
	mu      sync.RWMutex
	wlock   sync.Mutex
}

type muxPacketType byte

const (
	muxPacketSyn muxPacketType = iota
	muxPacketAck
	muxPacketFin
	muxPacketData
)

func NewMuxConn(rwc io.ReadWriteCloser) *MuxConn {
	m := &MuxConn{
		rwc:     rwc,
		streams: make(map[uint32]*Stream),
	}

	go m.loop()

	return m
}

// Close closes the underlying io.ReadWriteCloser. This will also close
// all streams that are open.
func (m *MuxConn) Close() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Close all the streams
	for _, w := range m.streams {
		w.Close()
	}
	m.streams = make(map[uint32]*Stream)

	return m.rwc.Close()
}

// Accept accepts a multiplexed connection with the given ID. This
// will block until a request is made to connect.
func (m *MuxConn) Accept(id uint32) (io.ReadWriteCloser, error) {
	stream, err := m.openStream(id)
	if err != nil {
		return nil, err
	}

	// If the stream isn't closed, then it is already open somehow
	stream.mu.Lock()
	if stream.state != streamStateSynRecv && stream.state != streamStateClosed {
		stream.mu.Unlock()
		return nil, fmt.Errorf("Stream %d already open in bad state: %d", id, stream.state)
	}

	if stream.state == streamStateSynRecv {
		// Fast track establishing since we already got the syn
		stream.setState(streamStateEstablished)
		stream.mu.Unlock()
	}

	if stream.state != streamStateEstablished {
		// Go into the listening state
		stream.setState(streamStateListen)
		stream.mu.Unlock()

		// Wait for the connection to establish
	ACCEPT_ESTABLISH_LOOP:
		for {
			time.Sleep(50 * time.Millisecond)
			stream.mu.Lock()
			switch stream.state {
			case streamStateListen:
				stream.mu.Unlock()
			case streamStateClosed:
				// This can happen if it becomes established, some data is sent,
				// and it closed all within the time period we wait above.
				// This case will be fixed when we have edge-triggered checks.
				fallthrough
			case streamStateEstablished:
				stream.mu.Unlock()
				break ACCEPT_ESTABLISH_LOOP
			default:
				defer stream.mu.Unlock()
				return nil, fmt.Errorf("Stream %d went to bad state: %d", id, stream.state)
			}
		}
	}

	// Send the ack down
	if _, err := m.write(stream.id, muxPacketAck, nil); err != nil {
		return nil, err
	}

	return stream, nil
}

// Dial opens a connection to the remote end using the given stream ID.
// An Accept on the remote end will only work with if the IDs match.
func (m *MuxConn) Dial(id uint32) (io.ReadWriteCloser, error) {
	stream, err := m.openStream(id)
	if err != nil {
		return nil, err
	}

	// If the stream isn't closed, then it is already open somehow
	stream.mu.Lock()
	if stream.state != streamStateClosed {
		stream.mu.Unlock()
		return nil, fmt.Errorf("Stream %d already open in bad state: %d", id, stream.state)
	}

	// Open a connection
	if _, err := m.write(stream.id, muxPacketSyn, nil); err != nil {
		return nil, err
	}
	stream.setState(streamStateSynSent)
	stream.mu.Unlock()

	for {
		time.Sleep(50 * time.Millisecond)
		stream.mu.Lock()
		switch stream.state {
		case streamStateSynSent:
			stream.mu.Unlock()
		case streamStateClosed:
			// This can happen if it becomes established, some data is sent,
			// and it closed all within the time period we wait above.
			// This case will be fixed when we have edge-triggered checks.
			fallthrough
		case streamStateEstablished:
			stream.mu.Unlock()
			return stream, nil
		default:
			defer stream.mu.Unlock()
			return nil, fmt.Errorf("Stream %d went to bad state: %d", id, stream.state)
		}
	}
}

// NextId returns the next available stream ID that isn't currently
// taken.
func (m *MuxConn) NextId() uint32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	for {
		result := m.curId
		m.curId++
		if _, ok := m.streams[result]; !ok {
			return result
		}
	}
}

func (m *MuxConn) openStream(id uint32) (*Stream, error) {
	// First grab a read-lock if we have the stream already we can
	// cheaply return it.
	m.mu.RLock()
	if stream, ok := m.streams[id]; ok {
		m.mu.RUnlock()
		return stream, nil
	}

	// Now acquire a full blown write lock so we can create the stream
	m.mu.RUnlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	// We have to check this again because there is a time period
	// above where we couldn't lost this lock.
	if stream, ok := m.streams[id]; ok {
		return stream, nil
	}

	// Create the stream object and channel where data will be sent to
	dataR, dataW := io.Pipe()
	writeCh := make(chan []byte, 10)

	// Set the data channel so we can write to it.
	stream := &Stream{
		id:      id,
		mux:     m,
		reader:  dataR,
		writeCh: writeCh,
	}
	stream.setState(streamStateClosed)

	// Start the goroutine that will read from the queue and write
	// data out.
	go func() {
		defer dataW.Close()

		for {
			data := <-writeCh
			if data == nil {
				// A nil is a tombstone letting us know we're done
				// accepting data.
				return
			}

			if _, err := dataW.Write(data); err != nil {
				return
			}
		}
	}()

	m.streams[id] = stream
	return m.streams[id], nil
}

func (m *MuxConn) loop() {
	defer m.Close()

	var id uint32
	var packetType muxPacketType
	var length int32
	for {
		if err := binary.Read(m.rwc, binary.BigEndian, &id); err != nil {
			log.Printf("[ERR] Error reading stream ID: %s", err)
			return
		}
		if err := binary.Read(m.rwc, binary.BigEndian, &packetType); err != nil {
			log.Printf("[ERR] Error reading packet type: %s", err)
			return
		}
		if err := binary.Read(m.rwc, binary.BigEndian, &length); err != nil {
			log.Printf("[ERR] Error reading length: %s", err)
			return
		}

		// TODO(mitchellh): probably would be better to re-use a buffer...
		data := make([]byte, length)
		if length > 0 {
			if _, err := m.rwc.Read(data); err != nil {
				log.Printf("[ERR] Error reading data: %s", err)
				return
			}
		}

		stream, err := m.openStream(id)
		if err != nil {
			log.Printf("[ERR] Error opening stream %d: %s", id, err)
			return
		}

		log.Printf("[DEBUG] Stream %d received packet %d", id, packetType)
		switch packetType {
		case muxPacketAck:
			stream.mu.Lock()
			switch stream.state {
			case streamStateSynSent:
				stream.setState(streamStateEstablished)
			case streamStateFinWait1:
				stream.remoteClose()
			default:
				log.Printf("[ERR] Ack received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		case muxPacketSyn:
			stream.mu.Lock()
			switch stream.state {
			case streamStateClosed:
				stream.setState(streamStateSynRecv)
			case streamStateListen:
				stream.setState(streamStateEstablished)
			default:
				log.Printf("[ERR] Syn received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		case muxPacketFin:
			stream.mu.Lock()
			switch stream.state {
			case streamStateEstablished:
				m.write(id, muxPacketAck, nil)
				fallthrough
			case streamStateFinWait1:
				stream.remoteClose()

				// Remove this stream from being active so that it
				// can be re-used
				m.mu.Lock()
				delete(m.streams, stream.id)
				m.mu.Unlock()
			default:
				log.Printf("[ERR] Fin received for stream %d in state: %d", id, stream.state)
			}
			stream.mu.Unlock()

		case muxPacketData:
			stream.mu.Lock()
			if stream.state == streamStateEstablished {
				select {
				case stream.writeCh <- data:
				default:
					log.Printf("[ERR] Failed to write data, buffer full: %d", id)
				}
			} else {
				log.Printf("[ERR] Data received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		}
	}
}

func (m *MuxConn) write(id uint32, dataType muxPacketType, p []byte) (int, error) {
	m.wlock.Lock()
	defer m.wlock.Unlock()

	if err := binary.Write(m.rwc, binary.BigEndian, id); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, byte(dataType)); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, int32(len(p))); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	return m.rwc.Write(p)
}

// Stream is a single stream of data and implements io.ReadWriteCloser
type Stream struct {
	id           uint32
	mux          *MuxConn
	reader       io.Reader
	state        streamState
	stateUpdated time.Time
	mu           sync.Mutex
	writeCh      chan<- []byte
}

type streamState byte

const (
	streamStateClosed streamState = iota
	streamStateListen
	streamStateSynRecv
	streamStateSynSent
	streamStateEstablished
	streamStateFinWait1
)

func (s *Stream) Close() error {
	s.mu.Lock()
	if s.state != streamStateEstablished {
		s.mu.Unlock()
		return fmt.Errorf("Stream in bad state: %d", s.state)
	}

	if _, err := s.mux.write(s.id, muxPacketFin, nil); err != nil {
		return err
	}
	s.setState(streamStateFinWait1)
	s.mu.Unlock()

	for {
		time.Sleep(50 * time.Millisecond)
		s.mu.Lock()
		switch s.state {
		case streamStateFinWait1:
			s.mu.Unlock()
		case streamStateClosed:
			s.mu.Unlock()
			return nil
		default:
			defer s.mu.Unlock()
			return fmt.Errorf("Stream %d went to bad state: %d", s.id, s.state)
		}
	}

	return nil
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *Stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	state := s.state
	s.mu.Unlock()

	if state != streamStateEstablished {
		return 0, fmt.Errorf("Stream in bad state to send: %d", state)
	}

	return s.mux.write(s.id, muxPacketData, p)
}

func (s *Stream) remoteClose() {
	s.setState(streamStateClosed)
	s.writeCh <- nil
}

func (s *Stream) setState(state streamState) {
	s.state = state
	s.stateUpdated = time.Now().UTC()
}
