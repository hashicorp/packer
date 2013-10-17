package uuid

import (
	"fmt"
	"math/rand"
	"time"
)

// Generates a time ordered UUID. Top 32 bits are a timestamp,
// bottom 96 are random.
func TimeOrderedUUID() string {
	unix := uint32(time.Now().UTC().Unix())
	rand1 := rand.Uint32()
	rand2 := rand.Uint32()
	rand3 := rand.Uint32()
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		unix,
		uint16(rand1>>16),
		uint16(rand1&0xffff),
		uint16(rand2>>16),
		uint16(rand2&0xffff),
		rand3)
}
