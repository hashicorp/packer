// Copyright 2016 The go-qemu Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qmp

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// A SocketMonitor is a Monitor which speaks directly to a QEMU Machine Protocol
// (QMP) socket. Communication is performed directly using a QEMU monitor socket,
// typically using a UNIX socket or TCP connection.  Multiple connections to the
// same domain are not permitted, and will result in the monitor blocking until
// the existing connection is closed.
type SocketMonitor struct {
	// QEMU version reported by a connected monitor socket.
	Version *Version

	// Underlying connection
	c net.Conn

	// Serialize running command against domain
	mu sync.Mutex

	// Send command responses and errors
	stream <-chan streamResponse

	// Send domain events to listeners when available
	listeners *int32
	events    <-chan Event
}

// NewSocketMonitor configures a connection to the provided QEMU monitor socket.
// An error is returned if the socket cannot be successfully dialed, or the
// dial attempt times out.
//
// NewSocketMonitor may dial the QEMU socket using a variety of connection types:
//	NewSocketMonitor("unix", "/var/lib/qemu/example.monitor", 2 * time.Second)
//	NewSocketMonitor("tcp", "8.8.8.8:4444", 2 * time.Second)
func NewSocketMonitor(network, addr string, timeout time.Duration) (*SocketMonitor, error) {
	c, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}

	mon := &SocketMonitor{
		c:         c,
		listeners: new(int32),
	}

	return mon, nil
}

// Listen creates a new SocketMonitor listening for a single connection to the provided socket file or address.
// An error is returned if unable to listen at the specified file path or port.
//
// Listen will wait for a QEMU socket connection using a variety connection types:
//	Listen("unix", "/var/lib/qemu/example.monitor")
//	Listen("tcp", "0.0.0.0:4444")
func Listen(network, addr string) (*SocketMonitor, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	c, err := l.Accept()
	defer l.Close()
	if err != nil {
		return nil, err
	}

	mon := &SocketMonitor{
		c:         c,
		listeners: new(int32),
	}

	return mon, nil
}

// Disconnect closes the QEMU monitor socket connection.
func (mon *SocketMonitor) Disconnect() error {
	atomic.StoreInt32(mon.listeners, 0)
	err := mon.c.Close()

	return err
}

// qmpCapabilities is the command which must be executed to perform the
// QEMU QMP handshake.
const qmpCapabilities = "qmp_capabilities"

// Connect sets up a QEMU QMP connection by connecting directly to the QEMU
// monitor socket.  An error is returned if the capabilities handshake does
// not succeed.
func (mon *SocketMonitor) Connect() error {
	enc := json.NewEncoder(mon.c)
	dec := json.NewDecoder(mon.c)

	// Check for banner on startup
	var ban banner
	if err := dec.Decode(&ban); err != nil {
		return err
	}
	mon.Version = &ban.QMP.Version

	// Issue capabilities handshake
	cmd := Command{Execute: qmpCapabilities}
	if err := enc.Encode(cmd); err != nil {
		return err
	}

	// Check for no error on return
	var r response
	if err := dec.Decode(&r); err != nil {
		return err
	}
	if err := r.Err(); err != nil {
		return err
	}

	// Initialize socket listener for command responses and asynchronous
	// events
	events := make(chan Event)
	stream := make(chan streamResponse)
	go mon.listen(mon.c, events, stream)

	mon.events = events
	mon.stream = stream

	return nil
}

// Events streams QEMU QMP Events.
// Events should only be called once per Socket.  If used with a qemu.Domain,
// qemu.Domain.Events should be called to retrieve events instead.
func (mon *SocketMonitor) Events() (<-chan Event, error) {
	atomic.AddInt32(mon.listeners, 1)
	return mon.events, nil
}

// listen listens for incoming data from a QEMU monitor socket.  It determines
// if the data is an asynchronous event or a response to a command, and returns
// the data on the appropriate channel.
func (mon *SocketMonitor) listen(r io.Reader, events chan<- Event, stream chan<- streamResponse) {
	defer close(events)
	defer close(stream)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var e Event

		b := scanner.Bytes()
		if err := json.Unmarshal(b, &e); err != nil {
			continue
		}

		// If data does not have an event type, it must be in response to a command.
		if e.Event == "" {
			stream <- streamResponse{buf: b}
			continue
		}

		// If nobody is listening for events, do not bother sending them.
		if atomic.LoadInt32(mon.listeners) == 0 {
			continue
		}

		events <- e
	}

	if err := scanner.Err(); err != nil {
		stream <- streamResponse{err: err}
	}
}

// Run executes the given QAPI command against a domain's QEMU instance.
// For a list of available QAPI commands, see:
//	http://git.qemu.org/?p=qemu.git;a=blob;f=qapi-schema.json;hb=HEAD
func (mon *SocketMonitor) Run(command []byte) ([]byte, error) {
	// Only allow a single command to be run at a time to ensure that responses
	// to a command cannot be mixed with responses from another command
	mon.mu.Lock()
	defer mon.mu.Unlock()

	if _, err := mon.c.Write(command); err != nil {
		return nil, err
	}

	// Wait for a response or error to our command
	res := <-mon.stream
	if res.err != nil {
		return nil, res.err
	}

	// Check for QEMU errors
	var r response
	if err := json.Unmarshal(res.buf, &r); err != nil {
		return nil, err
	}
	if err := r.Err(); err != nil {
		return nil, err
	}

	return res.buf, nil
}

// banner is a wrapper type around a Version.
type banner struct {
	QMP struct {
		Version Version `json:"version"`
	} `json:"QMP"`
}

// streamResponse is a struct sent over a channel in response to a command.
type streamResponse struct {
	buf []byte
	err error
}
