package pktline

// pkt-line Format
// ---------------
//
// Much (but not all) of the payload is described around pkt-lines.
//
// A pkt-line is a variable length binary string.  The first four bytes
// of the line, the pkt-len, indicates the total length of the line,
// in hexadecimal.  The pkt-len includes the 4 bytes used to contain
// the length's hexadecimal representation.
//
// A pkt-line MAY contain binary data, so implementors MUST ensure
// pkt-line parsing/formatting routines are 8-bit clean.
//
// A non-binary line SHOULD BE terminated by an LF, which if present
// MUST be included in the total length.
//
// The maximum length of a pkt-line's data component is 65520 bytes.
// Implementations MUST NOT send pkt-line whose length exceeds 65524
// (65520 bytes of payload + 4 bytes of length data).
//
// Implementations SHOULD NOT send an empty pkt-line ("0004").
//
// A pkt-line with a length field of 0 ("0000"), called a flush-pkt,
// is a special case and MUST be handled differently than an empty
// pkt-line ("0004").
//
// ----
//   pkt-line     =  data-pkt / flush-pkt
//
//   data-pkt     =  pkt-len pkt-payload
//   pkt-len      =  4*(HEXDIG)
//   pkt-payload  =  (pkt-len - 4)*(OCTET)
//
//   flush-pkt    = "0000"
// ----
//
// Examples (as C-style strings):
//
// ----
//   pkt-line          actual value
//   ---------------------------------
//   "0006a\n"         "a\n"
//   "0005a"           "a"
//   "000bfoobar\n"    "foobar\n"
//   "0004"            ""
// ----
//
// Extracted from:
// https://github.com/git/git/blob/master/Documentation/technical/protocol-common.txt

const (
	HeaderLength = 4
	MaxLength    = 65524
)
