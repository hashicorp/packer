package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// MuxConn is able to multiplex multiple streams on top of any
// io.ReadWriteCloser. These streams act like TCP connections (Dial, Accept,
// Close, full duplex, etc.).
//
// The underlying io.ReadWriteCloser is expected to guarantee delivery
// and ordering, such as TCP. Congestion control and such aren't implemented
// by the streams, so that is also up to the underlying connection.
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
	doneCh  chan struct{}
}

type muxPacketType byte

const (
	muxPacketSyn muxPacketType = iota
	muxPacketSynAck
	muxPacketAck
	muxPacketFin
	muxPacketData
)

// Create a new MuxConn around any io.ReadWriteCloser.
func NewMuxConn(rwc io.ReadWriteCloser) *MuxConn {
	m := &MuxConn{
		rwc:     rwc,
		streams: make(map[uint32]*Stream),
		doneCh:  make(chan struct{}),
	}

	go m.cleaner()
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
	m.streams = make(map[uint32]*Stream)

	// Close the actual connection. This will also force the loop
	// to end since it'll read EOF or closed connection.
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
	defer stream.mu.Unlock()
	if stream.state != streamStateSynRecv && stream.state != streamStateClosed {
		return nil, fmt.Errorf("Stream %d already open in bad state: %d", id, stream.state)
	}

	if stream.state == streamStateClosed {
		// Go into the listening state and wait for a syn
		stream.setState(streamStateListen)
		if err := stream.waitState(streamStateSynRecv); err != nil {
			return nil, err
		}
	}

	if stream.state == streamStateSynRecv {
		// Send a syn-ack
		if _, err := m.write(stream.id, muxPacketSynAck, nil); err != nil {
			return nil, err
		}
	}

	if err := stream.waitState(streamStateEstablished); err != nil {
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
	defer stream.mu.Unlock()
	if stream.state != streamStateClosed {
		return nil, fmt.Errorf("Stream %d already open in bad state: %d", id, stream.state)
	}

	// Open a connection
	if _, err := m.write(stream.id, muxPacketSyn, nil); err != nil {
		return nil, err
	}
	stream.setState(streamStateSynSent)

	if err := stream.waitState(streamStateEstablished); err != nil {
		return nil, err
	}

	m.write(id, muxPacketAck, nil)
	return stream, nil
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

	// Make sure we attempt to use the next biggest stream ID
	if id >= m.curId {
		m.curId = id + 1
	}

	// We have to check this again because there is a time period
	// above where we couldn't lost this lock.
	if stream, ok := m.streams[id]; ok {
		return stream, nil
	}

	// Create the stream object and channel where data will be sent to
	dataR, dataW := io.Pipe()
	writeCh := make(chan []byte, 256)

	// Set the data channel so we can write to it.
	stream := &Stream{
		id:          id,
		mux:         m,
		reader:      dataR,
		writeCh:     writeCh,
		stateChange: make(map[chan<- streamState]struct{}),
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

func (m *MuxConn) cleaner() {
	for {
		done := false
		select {
		case <-time.After(500 * time.Millisecond):
		case <-m.doneCh:
			done = true
		}

		m.mu.Lock()
		for id, s := range m.streams {
			s.mu.Lock()
			if s.state == streamStateClosed {
				delete(m.streams, id)
			}
			s.mu.Unlock()
		}

		if done {
			for _, s := range m.streams {
				s.mu.Lock()
				s.closeWriter()
				s.mu.Unlock()
			}
		}
		m.mu.Unlock()

		if done {
			return
		}
	}
}

func (m *MuxConn) loop() {
	// Force close every stream that we know about when we exit so
	// that they all read EOF and don't block forever.
	defer func() {
		log.Printf("[INFO] Mux connection loop exiting")
		close(m.doneCh)
	}()

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

		//log.Printf("[TRACE] Stream %d received packet %d", id, packetType)
		switch packetType {
		case muxPacketSyn:
			stream.mu.Lock()
			switch stream.state {
			case streamStateClosed:
				fallthrough
			case streamStateListen:
				stream.setState(streamStateSynRecv)
			default:
				log.Printf("[ERR] Syn received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		case muxPacketAck:
			stream.mu.Lock()
			switch stream.state {
			case streamStateSynRecv:
				stream.setState(streamStateEstablished)
			case streamStateFinWait1:
				stream.setState(streamStateFinWait2)
			case streamStateLastAck:
				stream.closeWriter()
				fallthrough
			case streamStateClosing:
				stream.setState(streamStateClosed)
			default:
				log.Printf("[ERR] Ack received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		case muxPacketSynAck:
			stream.mu.Lock()
			switch stream.state {
			case streamStateSynSent:
				stream.setState(streamStateEstablished)
			default:
				log.Printf("[ERR] SynAck received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		case muxPacketFin:
			stream.mu.Lock()
			switch stream.state {
			case streamStateEstablished:
				stream.closeWriter()
				stream.setState(streamStateCloseWait)
				m.write(id, muxPacketAck, nil)
			case streamStateFinWait2:
				stream.closeWriter()
				stream.setState(streamStateClosed)
				m.write(id, muxPacketAck, nil)
			case streamStateFinWait1:
				stream.closeWriter()
				stream.setState(streamStateClosing)
				m.write(id, muxPacketAck, nil)
			default:
				log.Printf("[ERR] Fin received for stream %d in state: %d", id, stream.state)
			}
			stream.mu.Unlock()

		case muxPacketData:
			stream.mu.Lock()
			switch stream.state {
			case streamStateFinWait1:
				fallthrough
			case streamStateFinWait2:
				fallthrough
			case streamStateEstablished:
				select {
				case stream.writeCh <- data:
				default:
					panic(fmt.Sprintf("Failed to write data, buffer full for stream %d", id))
				}
			default:
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

// Stream is a single stream of data and implements io.ReadWriteCloser.
// A Stream is full-duplex so you can write data as well as read data.
type Stream struct {
	id           uint32
	mux          *MuxConn
	reader       io.Reader
	state        streamState
	stateChange  map[chan<- streamState]struct{}
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
	streamStateFinWait2
	streamStateCloseWait
	streamStateClosing
	streamStateLastAck
)

func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != streamStateEstablished && s.state != streamStateCloseWait {
		return fmt.Errorf("Stream in bad state: %d", s.state)
	}

	if s.state == streamStateEstablished {
		s.setState(streamStateFinWait1)
	} else {
		s.setState(streamStateLastAck)
	}

	s.mux.write(s.id, muxPacketFin, nil)
	return nil
}

func (s *Stream) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *Stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	state := s.state
	s.mu.Unlock()

	if state != streamStateEstablished && state != streamStateCloseWait {
		return 0, fmt.Errorf("Stream %d in bad state to send: %d", s.id, state)
	}

	return s.mux.write(s.id, muxPacketData, p)
}

func (s *Stream) closeWriter() {
	s.writeCh <- nil
}

func (s *Stream) setState(state streamState) {
	//log.Printf("[TRACE] Stream %d went to state %d", s.id, state)
	s.state = state
	s.stateUpdated = time.Now().UTC()
	for ch, _ := range s.stateChange {
		select {
		case ch <- state:
		default:
		}
	}
}

func (s *Stream) waitState(target streamState) error {
	// Register a state change listener to wait for changes
	stateCh := make(chan streamState, 10)
	s.stateChange[stateCh] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.stateChange, stateCh)
	}()

	state := <-stateCh
	if state == target {
		return nil
	} else {
		return fmt.Errorf("Stream %d went to bad state: %d", s.id, state)
	}
}
