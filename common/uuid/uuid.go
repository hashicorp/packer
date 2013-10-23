package uuid

import (
	"fmt"
	"crypto/rand"
	"encoding/binary"
	"time"
)

func uint32rand() (value uint32) {
	err := binary.Read(rand.Reader, binary.LittleEndian, &value)
	if err != nil {
		panic(err)
	}
	return
}

// Generates a time ordered UUID. Top 32 bits are a timestamp,
// bottom 96 are random.
func TimeOrderedUUID() string {
	unix := uint32(time.Now().UTC().Unix())
	rand1 := uint32rand()
	rand2 := uint32rand()
	rand3 := uint32rand()
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		unix,
		uint16(rand1>>16),
		uint16(rand1&0xffff),
		uint16(rand2>>16),
		uint16(rand2&0xffff),
		rand3)
}
