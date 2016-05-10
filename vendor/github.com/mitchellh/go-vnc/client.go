// Package vnc implements a VNC client.
//
// References:
//   [PROTOCOL]: http://tools.ietf.org/html/rfc6143
package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"unicode"
)

type ClientConn struct {
	c      net.Conn
	config *ClientConfig

	// If the pixel format uses a color map, then this is the color
	// map that is used. This should not be modified directly, since
	// the data comes from the server.
	ColorMap [256]Color

	// Encodings supported by the client. This should not be modified
	// directly. Instead, SetEncodings should be used.
	Encs []Encoding

	// Width of the frame buffer in pixels, sent from the server.
	FrameBufferWidth uint16

	// Height of the frame buffer in pixels, sent from the server.
	FrameBufferHeight uint16

	// Name associated with the desktop, sent from the server.
	DesktopName string

	// The pixel format associated with the connection. This shouldn't
	// be modified. If you wish to set a new pixel format, use the
	// SetPixelFormat method.
	PixelFormat PixelFormat
}

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	// A slice of ClientAuth methods. Only the first instance that is
	// suitable by the server will be used to authenticate.
	Auth []ClientAuth

	// Exclusive determines whether the connection is shared with other
	// clients. If true, then all other clients connected will be
	// disconnected when a connection is established to the VNC server.
	Exclusive bool

	// The channel that all messages received from the server will be
	// sent on. If the channel blocks, then the goroutine reading data
	// from the VNC server may block indefinitely. It is up to the user
	// of the library to ensure that this channel is properly read.
	// If this is not set, then all messages will be discarded.
	ServerMessageCh chan<- ServerMessage

	// A slice of supported messages that can be read from the server.
	// This only needs to contain NEW server messages, and doesn't
	// need to explicitly contain the RFC-required messages.
	ServerMessages []ServerMessage
}

func Client(c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	conn := &ClientConn{
		c:      c,
		config: cfg,
	}

	if err := conn.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	go conn.mainLoop()

	return conn, nil
}

func (c *ClientConn) Close() error {
	return c.c.Close()
}

// CutText tells the server that the client has new text in its cut buffer.
// The text string MUST only contain Latin-1 characters. This encoding
// is compatible with Go's native string format, but can only use up to
// unicode.MaxLatin values.
//
// See RFC 6143 Section 7.5.6
func (c *ClientConn) CutText(text string) error {
	var buf bytes.Buffer

	// This is the fixed size data we'll send
	fixedData := []interface{}{
		uint8(6),
		uint8(0),
		uint8(0),
		uint8(0),
		uint32(len(text)),
	}

	for _, val := range fixedData {
		if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
			return err
		}
	}

	for _, char := range text {
		if char > unicode.MaxLatin1 {
			return fmt.Errorf("Character '%s' is not valid Latin-1", char)
		}

		if err := binary.Write(&buf, binary.BigEndian, uint8(char)); err != nil {
			return err
		}
	}

	dataLength := 8 + len(text)
	if _, err := c.c.Write(buf.Bytes()[0:dataLength]); err != nil {
		return err
	}

	return nil
}

// Requests a framebuffer update from the server. There may be an indefinite
// time between the request and the actual framebuffer update being
// received.
//
// See RFC 6143 Section 7.5.3
func (c *ClientConn) FramebufferUpdateRequest(incremental bool, x, y, width, height uint16) error {
	var buf bytes.Buffer
	var incrementalByte uint8 = 0

	if incremental {
		incrementalByte = 1
	}

	data := []interface{}{
		uint8(3),
		incrementalByte,
		x, y, width, height,
	}

	for _, val := range data {
		if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
			return err
		}
	}

	if _, err := c.c.Write(buf.Bytes()[0:10]); err != nil {
		return err
	}

	return nil
}

// KeyEvent indiciates a key press or release and sends it to the server.
// The key is indicated using the X Window System "keysym" value. Use
// Google to find a reference of these values. To simulate a key press,
// you must send a key with both a down event, and a non-down event.
//
// See 7.5.4.
func (c *ClientConn) KeyEvent(keysym uint32, down bool) error {
	var downFlag uint8 = 0
	if down {
		downFlag = 1
	}

	data := []interface{}{
		uint8(4),
		downFlag,
		uint8(0),
		uint8(0),
		keysym,
	}

	for _, val := range data {
		if err := binary.Write(c.c, binary.BigEndian, val); err != nil {
			return err
		}
	}

	return nil
}

// PointerEvent indicates that pointer movement or a pointer button
// press or release.
//
// The mask is a bitwise mask of various ButtonMask values. When a button
// is set, it is pressed, when it is unset, it is released.
//
// See RFC 6143 Section 7.5.5
func (c *ClientConn) PointerEvent(mask ButtonMask, x, y uint16) error {
	var buf bytes.Buffer

	data := []interface{}{
		uint8(5),
		uint8(mask),
		x,
		y,
	}

	for _, val := range data {
		if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
			return err
		}
	}

	if _, err := c.c.Write(buf.Bytes()[0:6]); err != nil {
		return err
	}

	return nil
}

// SetEncodings sets the encoding types in which the pixel data can
// be sent from the server. After calling this method, the encs slice
// given should not be modified.
//
// See RFC 6143 Section 7.5.2
func (c *ClientConn) SetEncodings(encs []Encoding) error {
	data := make([]interface{}, 3+len(encs))
	data[0] = uint8(2)
	data[1] = uint8(0)
	data[2] = uint16(len(encs))

	for i, enc := range encs {
		data[3+i] = int32(enc.Type())
	}

	var buf bytes.Buffer
	for _, val := range data {
		if err := binary.Write(&buf, binary.BigEndian, val); err != nil {
			return err
		}
	}

	dataLength := 4 + (4 * len(encs))
	if _, err := c.c.Write(buf.Bytes()[0:dataLength]); err != nil {
		return err
	}

	c.Encs = encs

	return nil
}

// SetPixelFormat sets the format in which pixel values should be sent
// in FramebufferUpdate messages from the server.
//
// See RFC 6143 Section 7.5.1
func (c *ClientConn) SetPixelFormat(format *PixelFormat) error {
	var keyEvent [20]byte
	keyEvent[0] = 0

	pfBytes, err := writePixelFormat(format)
	if err != nil {
		return err
	}

	// Copy the pixel format bytes into the proper slice location
	copy(keyEvent[4:], pfBytes)

	// Send the data down the connection
	if _, err := c.c.Write(keyEvent[:]); err != nil {
		return err
	}

	// Reset the color map as according to RFC.
	var newColorMap [256]Color
	c.ColorMap = newColorMap

	return nil
}

const pvLen = 12 // ProtocolVersion message length.

func parseProtocolVersion(pv []byte) (uint, uint, error) {
	var major, minor uint

	if len(pv) < pvLen {
		return 0, 0, fmt.Errorf("ProtocolVersion message too short (%v < %v)", len(pv), pvLen)
	}

	l, err := fmt.Sscanf(string(pv), "RFB %d.%d\n", &major, &minor)
	if l != 2 {
		return 0, 0, fmt.Errorf("error parsing ProtocolVersion.")
	}
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}

func (c *ClientConn) handshake() error {
	var protocolVersion [pvLen]byte

	// 7.1.1, read the ProtocolVersion message sent by the server.
	if _, err := io.ReadFull(c.c, protocolVersion[:]); err != nil {
		return err
	}

	maxMajor, maxMinor, err := parseProtocolVersion(protocolVersion[:])
	if err != nil {
		return err
	}
	if maxMajor < 3 {
		return fmt.Errorf("unsupported major version, less than 3: %d", maxMajor)
	}
	if maxMinor < 8 {
		return fmt.Errorf("unsupported minor version, less than 8: %d", maxMinor)
	}

	// Respond with the version we will support
	if _, err = c.c.Write([]byte("RFB 003.008\n")); err != nil {
		return err
	}

	// 7.1.2 Security Handshake from server
	var numSecurityTypes uint8
	if err = binary.Read(c.c, binary.BigEndian, &numSecurityTypes); err != nil {
		return err
	}

	if numSecurityTypes == 0 {
		return fmt.Errorf("no security types: %s", c.readErrorReason())
	}

	securityTypes := make([]uint8, numSecurityTypes)
	if err = binary.Read(c.c, binary.BigEndian, &securityTypes); err != nil {
		return err
	}

	clientSecurityTypes := c.config.Auth
	if clientSecurityTypes == nil {
		clientSecurityTypes = []ClientAuth{new(ClientAuthNone)}
	}

	var auth ClientAuth
FindAuth:
	for _, curAuth := range clientSecurityTypes {
		for _, securityType := range securityTypes {
			if curAuth.SecurityType() == securityType {
				// We use the first matching supported authentication
				auth = curAuth
				break FindAuth
			}
		}
	}

	if auth == nil {
		return fmt.Errorf("no suitable auth schemes found. server supported: %#v", securityTypes)
	}

	// Respond back with the security type we'll use
	if err = binary.Write(c.c, binary.BigEndian, auth.SecurityType()); err != nil {
		return err
	}

	if err = auth.Handshake(c.c); err != nil {
		return err
	}

	// 7.1.3 SecurityResult Handshake
	var securityResult uint32
	if err = binary.Read(c.c, binary.BigEndian, &securityResult); err != nil {
		return err
	}

	if securityResult == 1 {
		return fmt.Errorf("security handshake failed: %s", c.readErrorReason())
	}

	// 7.3.1 ClientInit
	var sharedFlag uint8 = 1
	if c.config.Exclusive {
		sharedFlag = 0
	}

	if err = binary.Write(c.c, binary.BigEndian, sharedFlag); err != nil {
		return err
	}

	// 7.3.2 ServerInit
	if err = binary.Read(c.c, binary.BigEndian, &c.FrameBufferWidth); err != nil {
		return err
	}

	if err = binary.Read(c.c, binary.BigEndian, &c.FrameBufferHeight); err != nil {
		return err
	}

	// Read the pixel format
	if err = readPixelFormat(c.c, &c.PixelFormat); err != nil {
		return err
	}

	var nameLength uint32
	if err = binary.Read(c.c, binary.BigEndian, &nameLength); err != nil {
		return err
	}

	nameBytes := make([]uint8, nameLength)
	if err = binary.Read(c.c, binary.BigEndian, &nameBytes); err != nil {
		return err
	}

	c.DesktopName = string(nameBytes)

	return nil
}

// mainLoop reads messages sent from the server and routes them to the
// proper channels for users of the client to read.
func (c *ClientConn) mainLoop() {
	defer c.Close()

	// Build the map of available server messages
	typeMap := make(map[uint8]ServerMessage)

	defaultMessages := []ServerMessage{
		new(FramebufferUpdateMessage),
		new(SetColorMapEntriesMessage),
		new(BellMessage),
		new(ServerCutTextMessage),
	}

	for _, msg := range defaultMessages {
		typeMap[msg.Type()] = msg
	}

	if c.config.ServerMessages != nil {
		for _, msg := range c.config.ServerMessages {
			typeMap[msg.Type()] = msg
		}
	}

	for {
		var messageType uint8
		if err := binary.Read(c.c, binary.BigEndian, &messageType); err != nil {
			break
		}

		msg, ok := typeMap[messageType]
		if !ok {
			// Unsupported message type! Bad!
			break
		}

		parsedMsg, err := msg.Read(c, c.c)
		if err != nil {
			break
		}

		if c.config.ServerMessageCh == nil {
			continue
		}

		c.config.ServerMessageCh <- parsedMsg
	}
}

func (c *ClientConn) readErrorReason() string {
	var reasonLen uint32
	if err := binary.Read(c.c, binary.BigEndian, &reasonLen); err != nil {
		return "<error>"
	}

	reason := make([]uint8, reasonLen)
	if err := binary.Read(c.c, binary.BigEndian, &reason); err != nil {
		return "<error>"
	}

	return string(reason)
}
