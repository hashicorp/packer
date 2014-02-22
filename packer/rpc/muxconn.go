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
	curId         uint32
	rwc           io.ReadWriteCloser
	streamsAccept map[uint32]*Stream
	streamsDial   map[uint32]*Stream
	muAccept      sync.RWMutex
	muDial        sync.RWMutex
	wlock         sync.Mutex
	doneCh        chan struct{}
}

type muxPacketFrom byte
type muxPacketType byte

const (
	muxPacketFromAccept muxPacketFrom = iota
	muxPacketFromDial
)

const (
	muxPacketSyn muxPacketType = iota
	muxPacketSynAck
	muxPacketAck
	muxPacketFin
	muxPacketData
)

func (f muxPacketFrom) String() string {
	switch f {
	case muxPacketFromAccept:
		return "accept"
	case muxPacketFromDial:
		return "dial"
	default:
		panic("unknown from type")
	}
}

// Create a new MuxConn around any io.ReadWriteCloser.
func NewMuxConn(rwc io.ReadWriteCloser) *MuxConn {
	m := &MuxConn{
		rwc:           rwc,
		streamsAccept: make(map[uint32]*Stream),
		streamsDial:   make(map[uint32]*Stream),
		doneCh:        make(chan struct{}),
	}

	go m.cleaner()
	go m.loop()

	return m
}

// Close closes the underlying io.ReadWriteCloser. This will also close
// all streams that are open.
func (m *MuxConn) Close() error {
	m.muAccept.Lock()
	m.muDial.Lock()
	defer m.muAccept.Unlock()
	defer m.muDial.Unlock()

	// Close all the streams
	for _, w := range m.streamsAccept {
		w.Close()
	}
	for _, w := range m.streamsDial {
		w.Close()
	}
	m.streamsAccept = make(map[uint32]*Stream)
	m.streamsDial = make(map[uint32]*Stream)

	// Close the actual connection. This will also force the loop
	// to end since it'll read EOF or closed connection.
	return m.rwc.Close()
}

// Accept accepts a multiplexed connection with the given ID. This
// will block until a request is made to connect.
func (m *MuxConn) Accept(id uint32) (io.ReadWriteCloser, error) {
	//log.Printf("[TRACE] %p: Accept on stream ID: %d", m, id)

	// Get the stream. It is okay if it is already in the list of streams
	// because we may have prematurely received a syn for it.
	m.muAccept.Lock()
	stream, ok := m.streamsAccept[id]
	if !ok {
		stream = newStream(muxPacketFromAccept, id, m)
		m.streamsAccept[id] = stream
	}
	m.muAccept.Unlock()

	stream.mu.Lock()
	defer stream.mu.Unlock()

	// If the stream isn't closed, then it is already open somehow
	if stream.state != streamStateSynRecv && stream.state != streamStateClosed {
		panic(fmt.Sprintf(
			"Stream %d already open in bad state: %d", id, stream.state))
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
		if _, err := stream.write(muxPacketSynAck, nil); err != nil {
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
	//log.Printf("[TRACE] %p: Dial on stream ID: %d", m, id)

	m.muDial.Lock()

	// If we have any streams with this ID, then it is a failure. The
	// reaper should clear out old streams once in awhile.
	if stream, ok := m.streamsDial[id]; ok {
		m.muDial.Unlock()
		panic(fmt.Sprintf(
			"Stream %d already open for dial. State: %d",
			id, stream.state))
	}

	// Create the new stream and put it in our list. We can then
	// unlock because dialing will no longer be allowed on that ID.
	stream := newStream(muxPacketFromDial, id, m)
	m.streamsDial[id] = stream

	// Don't let anyone else mess with this stream
	stream.mu.Lock()
	defer stream.mu.Unlock()

	m.muDial.Unlock()

	// Open a connection
	if _, err := stream.write(muxPacketSyn, nil); err != nil {
		return nil, err
	}

	// It is safe to set the state after the write above because
	// we hold the stream lock.
	stream.setState(streamStateSynSent)

	if err := stream.waitState(streamStateEstablished); err != nil {
		return nil, err
	}

	stream.write(muxPacketAck, nil)
	return stream, nil
}

// NextId returns the next available listen stream ID that isn't currently
// taken.
func (m *MuxConn) NextId() uint32 {
	m.muAccept.Lock()
	defer m.muAccept.Unlock()

	for {
		// We never use stream ID 0 because 0 is the zero value of a uint32
		// and we want to reserve that for "not in use"
		if m.curId == 0 {
			m.curId = 1
		}

		result := m.curId
		m.curId += 1
		if _, ok := m.streamsAccept[result]; !ok {
			return result
		}
	}
}

func (m *MuxConn) cleaner() {
	checks := []struct {
		Map  *map[uint32]*Stream
		Lock *sync.RWMutex
	}{
		{&m.streamsAccept, &m.muAccept},
		{&m.streamsDial, &m.muDial},
	}

	for {
		done := false
		select {
		case <-time.After(500 * time.Millisecond):
		case <-m.doneCh:
			done = true
		}

		for _, check := range checks {
			check.Lock.Lock()
			for id, s := range *check.Map {
				s.mu.Lock()

				if done && s.state != streamStateClosed {
					s.closeWriter()
				}

				if s.state == streamStateClosed {
					// Only clean up the streams that have been closed
					// for a certain amount of time.
					since := time.Now().UTC().Sub(s.stateUpdated)
					if since > 2*time.Second {
						delete(*check.Map, id)
					}
				}

				s.mu.Unlock()
			}
			check.Lock.Unlock()
		}

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

	var from muxPacketFrom
	var id uint32
	var packetType muxPacketType
	var length int32
	for {
		if err := binary.Read(m.rwc, binary.BigEndian, &from); err != nil {
			log.Printf("[ERR] Error reading stream direction: %s", err)
			return
		}
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
		n := 0
		for n < int(length) {
			if n2, err := m.rwc.Read(data[n:]); err != nil {
				log.Printf("[ERR] Error reading data: %s", err)
				return
			} else {
				n += n2
			}
		}

		// Get the proper stream. Note that the map we look into is
		// opposite the "from" because if the dial side is talking to
		// us, we need to look into the accept map, and so on.
		//
		// Note: we also switch the "from" value so that logging
		// below is correct.
		var stream *Stream
		switch from {
		case muxPacketFromDial:
			m.muAccept.Lock()
			stream = m.streamsAccept[id]
			m.muAccept.Unlock()

			from = muxPacketFromAccept
		case muxPacketFromAccept:
			m.muDial.Lock()
			stream = m.streamsDial[id]
			m.muDial.Unlock()

			from = muxPacketFromDial
		default:
			panic(fmt.Sprintf("Unknown stream direction: %d", from))
		}

		if stream == nil && packetType != muxPacketSyn {
			log.Printf(
				"[WARN] %p: Non-existent stream %d (%s) received packer %d",
				m, id, from, packetType)
			continue
		}

		//log.Printf("[TRACE] %p: Stream %d (%s) received packet %d", m, id, from, packetType)
		switch packetType {
		case muxPacketSyn:
			// If the stream is nil, this is the only case where we'll
			// automatically create the stream struct.
			if stream == nil {
				var ok bool

				m.muAccept.Lock()
				stream, ok = m.streamsAccept[id]
				if !ok {
					stream = newStream(muxPacketFromAccept, id, m)
					m.streamsAccept[id] = stream
				}
				m.muAccept.Unlock()
			}

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
				stream.write(muxPacketAck, nil)
			case streamStateFinWait2:
				stream.closeWriter()
				stream.setState(streamStateClosed)
				stream.write(muxPacketAck, nil)
			case streamStateFinWait1:
				stream.closeWriter()
				stream.setState(streamStateClosing)
				stream.write(muxPacketAck, nil)
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
				if len(data) > 0 && stream.writeCh != nil {
					//log.Printf("[TRACE] %p: Stream %d (%s) WRITE-START", m, id, from)
					stream.writeCh <- data
					//log.Printf("[TRACE] %p: Stream %d (%s) WRITE-END", m, id, from)
				}
			default:
				log.Printf("[ERR] Data received for stream in state: %d", stream.state)
			}
			stream.mu.Unlock()
		}
	}
}

func (m *MuxConn) write(from muxPacketFrom, id uint32, dataType muxPacketType, p []byte) (int, error) {
	m.wlock.Lock()
	defer m.wlock.Unlock()

	if err := binary.Write(m.rwc, binary.BigEndian, from); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, id); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, byte(dataType)); err != nil {
		return 0, err
	}
	if err := binary.Write(m.rwc, binary.BigEndian, int32(len(p))); err != nil {
		return 0, err
	}

	// Write all the bytes. If we don't write all the bytes, report an error
	var err error = nil
	n := 0
	for n < len(p) {
		var n2 int
		n2, err = m.rwc.Write(p[n:])
		n += n2
		if err != nil {
			log.Printf("[ERR] %p: Stream %d (%s) write error: %s", m, id, from, err)
			break
		}
	}

	return n, err
}

// Stream is a single stream of data and implements io.ReadWriteCloser.
// A Stream is full-duplex so you can write data as well as read data.
type Stream struct {
	from         muxPacketFrom
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

func newStream(from muxPacketFrom, id uint32, m *MuxConn) *Stream {
	// Create the stream object and channel where data will be sent to
	dataR, dataW := io.Pipe()
	writeCh := make(chan []byte, 4096)

	// Set the data channel so we can write to it.
	stream := &Stream{
		from:        from,
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

		drain := false
		for {
			data := <-writeCh
			if data == nil {
				// A nil is a tombstone letting us know we're done
				// accepting data.
				return
			}

			if drain {
				// We're draining, meaning we're just waiting for the
				// write channel to close.
				continue
			}

			if _, err := dataW.Write(data); err != nil {
				drain = true
			}
		}
	}()

	return stream
}

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

	s.write(muxPacketFin, nil)
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

	return s.write(muxPacketData, p)
}

func (s *Stream) closeWriter() {
	if s.writeCh != nil {
		s.writeCh <- nil
		s.writeCh = nil
	}
}

func (s *Stream) setState(state streamState) {
	//log.Printf("[TRACE] %p: Stream %d (%s) went to state %d", s.mux, s.id, s.from, state)
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

	//log.Printf("[TRACE] %p: Stream %d (%s) waiting for state: %d", s.mux, s.id, s.from, target)
	state := <-stateCh
	if state == target {
		return nil
	} else {
		return fmt.Errorf("Stream %d went to bad state: %d", s.id, state)
	}
}

func (s *Stream) write(dataType muxPacketType, p []byte) (int, error) {
	return s.mux.write(s.from, s.id, dataType, p)
}
