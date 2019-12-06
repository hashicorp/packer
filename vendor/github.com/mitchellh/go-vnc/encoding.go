package vnc

import (
	"encoding/binary"
	"io"
)

// An Encoding implements a method for encoding pixel data that is
// sent by the server to the client.
type Encoding interface {
	// The number that uniquely identifies this encoding type.
	Type() int32

	// Read reads the contents of the encoded pixel data from the reader.
	// This should return a new Encoding implementation that contains
	// the proper data.
	Read(*ClientConn, *Rectangle, io.Reader) (Encoding, error)
}

// RawEncoding is raw pixel data sent by the server.
//
// See RFC 6143 Section 7.7.1
type RawEncoding struct {
	Colors []Color
}

func (*RawEncoding) Type() int32 {
	return 0
}

func (*RawEncoding) Read(c *ClientConn, rect *Rectangle, r io.Reader) (Encoding, error) {
	bytesPerPixel := c.PixelFormat.BPP / 8
	pixelBytes := make([]uint8, bytesPerPixel)

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if c.PixelFormat.BigEndian {
		byteOrder = binary.BigEndian
	}

	colors := make([]Color, int(rect.Height)*int(rect.Width))

	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			if _, err := io.ReadFull(r, pixelBytes); err != nil {
				return nil, err
			}

			var rawPixel uint32
			if c.PixelFormat.BPP == 8 {
				rawPixel = uint32(pixelBytes[0])
			} else if c.PixelFormat.BPP == 16 {
				rawPixel = uint32(byteOrder.Uint16(pixelBytes))
			} else if c.PixelFormat.BPP == 32 {
				rawPixel = byteOrder.Uint32(pixelBytes)
			}

			color := &colors[int(y)*int(rect.Width)+int(x)]
			if c.PixelFormat.TrueColor {
				color.R = uint16((rawPixel >> c.PixelFormat.RedShift) & uint32(c.PixelFormat.RedMax))
				color.G = uint16((rawPixel >> c.PixelFormat.GreenShift) & uint32(c.PixelFormat.GreenMax))
				color.B = uint16((rawPixel >> c.PixelFormat.BlueShift) & uint32(c.PixelFormat.BlueMax))
			} else {
				*color = c.ColorMap[rawPixel]
			}
		}
	}

	return &RawEncoding{colors}, nil
}
