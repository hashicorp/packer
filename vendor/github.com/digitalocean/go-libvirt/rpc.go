// Copyright 2018 The go-libvirt Authors.
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

package libvirt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"sync/atomic"

	"github.com/digitalocean/go-libvirt/internal/constants"
	xdr "github.com/digitalocean/go-libvirt/internal/go-xdr/xdr2"
)

// ErrUnsupported is returned if a procedure is not supported by libvirt
var ErrUnsupported = errors.New("unsupported procedure requested")

// request and response types
const (
	// Call is used when making calls to the remote server.
	Call = iota

	// Reply indicates a server reply.
	Reply

	// Message is an asynchronous notification.
	Message

	// Stream represents a stream data packet.
	Stream

	// CallWithFDs is used by a client to indicate the request has
	// arguments with file descriptors.
	CallWithFDs

	// ReplyWithFDs is used by a server to indicate the request has
	// arguments with file descriptors.
	ReplyWithFDs
)

// request and response statuses
const (
	// StatusOK is always set for method calls or events.
	// For replies it indicates successful completion of the method.
	// For streams it indicates confirmation of the end of file on the stream.
	StatusOK = iota

	// StatusError for replies indicates that the method call failed
	// and error information is being returned. For streams this indicates
	// that not all data was sent and the stream has aborted.
	StatusError

	// StatusContinue is only used for streams.
	// This indicates that further data packets will be following.
	StatusContinue
)

// header is a libvirt rpc packet header
type header struct {
	// Program identifier
	Program uint32

	// Program version
	Version uint32

	// Remote procedure identifier
	Procedure uint32

	// Call type, e.g., Reply
	Type uint32

	// Call serial number
	Serial uint32

	// Request status, e.g., StatusOK
	Status uint32
}

// packet represents a RPC request or response.
type packet struct {
	// Size of packet, in bytes, including length.
	// Len + Header + Payload
	Len    uint32
	Header header
}

// internal rpc response
type response struct {
	Payload []byte
	Status  uint32
}

// event stream associated with a program and a procedure
type eventStream struct {
	// Channel of events sent by libvirt
	Events chan event

	// Remote procedure identifier used to unregister callback
	DeregisterProc uint32

	// Program identifier
	Program uint32
}

// helper to create an event stream
func newEventStream(deregisterProc, program uint32) eventStream {
	return eventStream{Events: make(chan event), DeregisterProc: deregisterProc, Program: program}
}

// libvirt error response
type libvirtError struct {
	Code     uint32
	DomainID uint32
	Padding  uint8
	Message  string
	Level    uint32
}

func (e libvirtError) Error() string {
	return e.Message
}

// checkError is used to check whether an error is a libvirtError, and if it is,
// whether its error code matches the one passed in. It will return false if
// these conditions are not met.
func checkError(err error, expectedError errorNumber) bool {
	e, ok := err.(libvirtError)
	if ok {
		return e.Code == uint32(expectedError)
	}
	return false
}

// IsNotFound detects libvirt's ERR_NO_DOMAIN.
func IsNotFound(err error) bool {
	return checkError(err, errNoDomain)
}

// listen processes incoming data and routes
// responses to their respective callback handler.
func (l *Libvirt) listen() {
	for {
		// response packet length
		length, err := pktlen(l.r)
		if err != nil {
			// When the underlying connection EOFs or is closed, stop
			// this goroutine
			if err == io.EOF || strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			// invalid packet
			continue
		}

		// response header
		h, err := extractHeader(l.r)
		if err != nil {
			// invalid packet
			continue
		}

		// payload: packet length minus what was previously read
		size := int(length) - (constants.PacketLengthSize + constants.HeaderSize)
		buf := make([]byte, size)
		_, err = io.ReadFull(l.r, buf)
		if err != nil {
			// invalid packet
			continue
		}

		// route response to caller
		l.route(h, buf)
	}
}

// callback sends rpc responses to their respective caller.
func (l *Libvirt) callback(id uint32, res response) {
	l.cm.Lock()
	c, ok := l.callbacks[id]
	l.cm.Unlock()
	if ok {
		// we close the channel in deregister() so that we don't block here
		// forever without a receiver. If that happens, this write will panic.
		defer func() {
			recover()
		}()
		c <- res
	}
}

// route sends incoming packets to their listeners.
func (l *Libvirt) route(h *header, buf []byte) {
	// route events to their respective listener
	var streamEvent event
	switch {
	case h.Program == constants.ProgramQEMU && h.Procedure == constants.QEMUDomainMonitorEvent:
		streamEvent = &DomainEvent{}
	case h.Program == constants.Program && h.Procedure == constants.ProcDomainEventCallbackLifecycle:
		streamEvent = &DomainEventCallbackLifecycleMsg{}
	}

	if streamEvent != nil {
		err := eventDecoder(buf, streamEvent)
		if err != nil { // event was malformed, drop.
			return
		}
		l.stream(streamEvent)
		return
	}

	// send responses to caller
	res := response{
		Payload: buf,
		Status:  h.Status,
	}
	l.callback(h.Serial, res)
}

// serial provides atomic access to the next sequential request serial number.
func (l *Libvirt) serial() uint32 {
	return atomic.AddUint32(&l.s, 1)
}

// stream decodes domain events and sends them
// to the respective event listener.
func (l *Libvirt) stream(e event) error {
	// send to event listener
	l.em.Lock()
	c, ok := l.events[e.GetCallbackID()]
	l.em.Unlock()

	if ok {
		// we close the channel in deregister() so that we don't block here
		// forever without a receiver. If that happens, this write will panic.
		defer func() {
			recover()
		}()
		c.Events <- e
	}
	return nil
}

// addStream configures the routing for an event stream.
func (l *Libvirt) addStream(id uint32, s eventStream) {
	l.em.Lock()
	l.events[id] = s
	l.em.Unlock()
}

// removeStream notifies the libvirt server to stop sending events
// for the provided callback id. Upon successful de-registration the
// callback handler is destroyed.
func (l *Libvirt) removeStream(id uint32) error {
	stream := l.events[id]
	close(stream.Events)

	payload := struct {
		CallbackID uint32
	}{
		CallbackID: id,
	}

	buf, err := encode(&payload)
	if err != nil {
		return err
	}

	_, err = l.request(stream.DeregisterProc, stream.Program, buf)
	if err != nil {
		return err
	}

	l.em.Lock()
	delete(l.events, id)
	l.em.Unlock()

	return nil
}

// register configures a method response callback
func (l *Libvirt) register(id uint32, c chan response) {
	l.cm.Lock()
	l.callbacks[id] = c
	l.cm.Unlock()
}

// deregister destroys a method response callback
func (l *Libvirt) deregister(id uint32) {
	l.cm.Lock()
	if _, ok := l.callbacks[id]; ok {
		close(l.callbacks[id])
		delete(l.callbacks, id)
	}
	l.cm.Unlock()
}

// deregisterAll closes all the waiting callback channels. This is used to clean
// up if the connection to libvirt is lost. Callers waiting for responses will
// return an error when the response channel is closed, rather than just
// hanging.
func (l *Libvirt) deregisterAll() {
	l.cm.Lock()
	for id := range l.callbacks {
		// can't call deregister() here because we're already holding the lock.
		close(l.callbacks[id])
		delete(l.callbacks, id)
	}
	l.cm.Unlock()
}

// request performs a libvirt RPC request.
// returns response returned by server.
// if response is not OK, decodes error from it and returns it.
func (l *Libvirt) request(proc uint32, program uint32, payload []byte) (response, error) {
	return l.requestStream(proc, program, payload, nil, nil)
}

func (l *Libvirt) processIncomingStream(c chan response, inStream io.Writer) (response, error) {
	for {
		resp, err := l.getResponse(c)
		if err != nil {
			return resp, err
		}
		// StatusOK here means end of stream
		if resp.Status == StatusOK {
			return resp, nil
		}
		// StatusError is handled in getResponse, so this is StatusContinue
		// StatusContinue is valid here only for stream packets
		// libvirtd breaks protocol and returns StatusContinue with empty Payload when stream finishes
		if len(resp.Payload) == 0 {
			return resp, nil
		}
		if inStream != nil {
			_, err = inStream.Write(resp.Payload)
			if err != nil {
				return response{}, err
			}
		}
	}
}

// requestStream performs a libvirt RPC request. The outStream and inStream
// parameters are optional, and should be nil for RPC endpoints that don't
// return a stream.
func (l *Libvirt) requestStream(proc uint32, program uint32, payload []byte,
	outStream io.Reader, inStream io.Writer) (response, error) {
	serial := l.serial()
	c := make(chan response)

	l.register(serial, c)
	defer l.deregister(serial)

	err := l.sendPacket(serial, proc, program, payload, Call, StatusOK)
	if err != nil {
		return response{}, err
	}

	resp, err := l.getResponse(c)
	if err != nil {
		return resp, err
	}

	if outStream != nil {
		abortOutStream := make(chan bool)
		outStreamErr := make(chan error)
		go func() {
			outStreamErr <- l.sendStream(serial, proc, program, outStream, abortOutStream)
		}()

		// Even without incoming stream server sends confirmation once all data is received
		resp, err = l.processIncomingStream(c, inStream)
		if err != nil {
			abortOutStream <- true
			return resp, err
		}

		err = <-outStreamErr
		if err != nil {
			return response{}, err
		}
	} else if inStream != nil {
		return l.processIncomingStream(c, inStream)
	}

	return resp, nil
}

func (l *Libvirt) sendStream(serial uint32, proc uint32, program uint32, stream io.Reader, abort chan bool) error {
	// Keep total packet length under 4 MiB to follow possible limitation in libvirt server code
	buf := make([]byte, 4*MiB-constants.HeaderSize)
	for {
		select {
		case <-abort:
			return l.sendPacket(serial, proc, program, nil, Stream, StatusError)
		default:
		}
		n, err := stream.Read(buf)
		if n > 0 {
			err2 := l.sendPacket(serial, proc, program, buf[:n], Stream, StatusContinue)
			if err2 != nil {
				return err2
			}
		}
		if err != nil {
			if err == io.EOF {
				return l.sendPacket(serial, proc, program, nil, Stream, StatusOK)
			}
			// keep original error
			err2 := l.sendPacket(serial, proc, program, nil, Stream, StatusError)
			if err2 != nil {
				return err2
			}
			return err
		}
	}
}

func (l *Libvirt) sendPacket(serial uint32, proc uint32, program uint32, payload []byte, typ uint32, status uint32) error {
	size := constants.PacketLengthSize + constants.HeaderSize
	if payload != nil {
		size += len(payload)
	}

	p := packet{
		Len: uint32(size),
		Header: header{
			Program:   program,
			Version:   constants.ProtocolVersion,
			Procedure: proc,
			Type:      typ,
			Serial:    serial,
			Status:    status,
		},
	}

	// write header
	l.mu.Lock()
	defer l.mu.Unlock()
	err := binary.Write(l.w, binary.BigEndian, p)
	if err != nil {
		return err
	}

	// write payload
	if payload != nil {
		err = binary.Write(l.w, binary.BigEndian, payload)
		if err != nil {
			return err
		}
	}

	return l.w.Flush()
}

func (l *Libvirt) getResponse(c chan response) (response, error) {
	resp := <-c
	if resp.Status == StatusError {
		return resp, decodeError(resp.Payload)
	}

	return resp, nil
}

// encode XDR encodes the provided data.
func encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	_, err := xdr.Marshal(&buf, data)

	return buf.Bytes(), err
}

// decodeError extracts an error message from the provider buffer.
func decodeError(buf []byte) error {
	var e libvirtError

	dec := xdr.NewDecoder(bytes.NewReader(buf))
	_, err := dec.Decode(&e)
	if err != nil {
		return err
	}

	if strings.Contains(e.Message, "unknown procedure") {
		return ErrUnsupported
	}
	// if libvirt returns ERR_OK, ignore the error
	if checkError(e, errOk) {
		return nil
	}

	return e
}

// eventDecoder decoder an event from a xdr buffer.
func eventDecoder(buf []byte, e interface{}) error {
	dec := xdr.NewDecoder(bytes.NewReader(buf))
	_, err := dec.Decode(e)
	return err
}

// pktlen determines the length of an incoming rpc response.
// If an error is encountered reading the provided Reader, the
// error is returned and response length will be 0.
func pktlen(r io.Reader) (uint32, error) {
	buf := make([]byte, constants.PacketLengthSize)

	for n := 0; n < cap(buf); {
		nn, err := r.Read(buf)
		if err != nil {
			return 0, err
		}

		n += nn
	}

	return binary.BigEndian.Uint32(buf), nil
}

// extractHeader returns the decoded header from an incoming response.
func extractHeader(r io.Reader) (*header, error) {
	buf := make([]byte, constants.HeaderSize)

	for n := 0; n < cap(buf); {
		nn, err := r.Read(buf)
		if err != nil {
			return nil, err
		}

		n += nn
	}

	h := &header{
		Program:   binary.BigEndian.Uint32(buf[0:4]),
		Version:   binary.BigEndian.Uint32(buf[4:8]),
		Procedure: binary.BigEndian.Uint32(buf[8:12]),
		Type:      binary.BigEndian.Uint32(buf[12:16]),
		Serial:    binary.BigEndian.Uint32(buf[16:20]),
		Status:    binary.BigEndian.Uint32(buf[20:24]),
	}

	return h, nil
}
