package bootcommand

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/packer/common"
)

// SendCodeFunc will be called to send codes to the VM
type SendCodeFunc func([]string) error

type pcATDriver struct {
	interval    time.Duration
	sendImpl    SendCodeFunc
	specialMap  map[string][]string
	scancodeMap map[rune]byte
	buffer      [][]string
	// TODO: set from env
	scancodeChunkSize int
}

// NewPCATDriver creates a new boot command driver for VMs that expect PC-AT
// keyboard codes. `send` should send its argument to the VM. `chunkSize` should
// be the maximum number of keyboard codes to send to `send` at one time.
func NewPCATDriver(send SendCodeFunc, chunkSize int) *pcATDriver {
	// We delay (default 100ms) between each input event to allow for CPU or
	// network latency. See PackerKeyEnv for tuning.
	keyInterval := common.PackerKeyDefault
	if delay, err := time.ParseDuration(os.Getenv(common.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}
	// Scancodes reference: http://www.win.tue.nl/~aeb/linux/kbd/scancodes-1.html
	//
	// Scancodes are recorded here in pairs. The first entry represents
	// the key press and the second entry represents the key release and is
	// derived from the first by the addition of 0x80.
	sMap := make(map[string][]string)
	sMap["bs"] = []string{"0e", "8e"}
	sMap["del"] = []string{"e053", "e0d3"}
	sMap["enter"] = []string{"1c", "9c"}
	sMap["esc"] = []string{"01", "81"}
	sMap["f1"] = []string{"3b", "bb"}
	sMap["f2"] = []string{"3c", "bc"}
	sMap["f3"] = []string{"3d", "bd"}
	sMap["f4"] = []string{"3e", "be"}
	sMap["f5"] = []string{"3f", "bf"}
	sMap["f6"] = []string{"40", "c0"}
	sMap["f7"] = []string{"41", "c1"}
	sMap["f8"] = []string{"42", "c2"}
	sMap["f9"] = []string{"43", "c3"}
	sMap["f10"] = []string{"44", "c4"}
	sMap["f11"] = []string{"57", "d7"}
	sMap["f12"] = []string{"58", "d8"}
	sMap["return"] = []string{"1c", "9c"}
	sMap["tab"] = []string{"0f", "8f"}
	sMap["up"] = []string{"e048", "e0c8"}
	sMap["down"] = []string{"e050", "e0d0"}
	sMap["left"] = []string{"e04b", "e0cb"}
	sMap["right"] = []string{"e04d", "e0cd"}
	sMap["spacebar"] = []string{"39", "b9"}
	sMap["insert"] = []string{"e052", "e0d2"}
	sMap["home"] = []string{"e047", "e0c7"}
	sMap["end"] = []string{"e04f", "e0cf"}
	sMap["pageUp"] = []string{"e049", "e0c9"}
	sMap["pageDown"] = []string{"e051", "e0d1"}
	sMap["leftAlt"] = []string{"38", "b8"}
	sMap["leftCtrl"] = []string{"1d", "9d"}
	sMap["leftShift"] = []string{"2a", "aa"}
	sMap["rightAlt"] = []string{"e038", "e0b8"}
	sMap["rightCtrl"] = []string{"e01d", "e09d"}
	sMap["rightShift"] = []string{"36", "b6"}
	sMap["leftSuper"] = []string{"e05b", "e0db"}
	sMap["rightSuper"] = []string{"e05c", "e0dc"}

	scancodeIndex := make(map[string]byte)
	scancodeIndex["1234567890-="] = 0x02
	scancodeIndex["!@#$%^&*()_+"] = 0x02
	scancodeIndex["qwertyuiop[]"] = 0x10
	scancodeIndex["QWERTYUIOP{}"] = 0x10
	scancodeIndex["asdfghjkl;'`"] = 0x1e
	scancodeIndex[`ASDFGHJKL:"~`] = 0x1e
	scancodeIndex[`\zxcvbnm,./`] = 0x2b
	scancodeIndex["|ZXCVBNM<>?"] = 0x2b
	scancodeIndex[" "] = 0x39

	scancodeMap := make(map[rune]byte)
	for chars, start := range scancodeIndex {
		var i byte = 0
		for len(chars) > 0 {
			r, size := utf8.DecodeRuneInString(chars)
			chars = chars[size:]
			scancodeMap[r] = start + i
			i += 1
		}
	}

	return &pcATDriver{
		interval:          keyInterval,
		sendImpl:          send,
		specialMap:        sMap,
		scancodeMap:       scancodeMap,
		scancodeChunkSize: chunkSize,
	}
}

// Finalize flushes all scanecodes.
func (d *pcATDriver) Finalize() error {
	defer func() {
		d.buffer = nil
	}()
	sc, err := chunkScanCodes(d.buffer, d.scancodeChunkSize)
	if err != nil {
		return err
	}
	for _, b := range sc {
		if err := d.sendImpl(b); err != nil {
			return err
		}
		time.Sleep(d.interval)
	}
	return nil
}

func (d *pcATDriver) SendKey(key rune, action KeyAction) error {

	keyShift := unicode.IsUpper(key) || strings.ContainsRune(shiftedChars, key)

	var scancode []string

	if action&(KeyOn|KeyPress) != 0 {
		scancodeInt := d.scancodeMap[key]
		if keyShift {
			scancode = append(scancode, "2a")
		}
		scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))
	}

	if action&(KeyOff|KeyPress) != 0 {
		scancodeInt := d.scancodeMap[key] + 0x80
		if keyShift {
			scancode = append(scancode, "aa")
		}
		scancode = append(scancode, fmt.Sprintf("%02x", scancodeInt))
	}

	for _, sc := range scancode {
		log.Printf("Sending char '%c', code '%s', shift %v", key, sc, keyShift)
	}

	d.send(scancode)
	return nil
}

func (d *pcATDriver) SendSpecial(special string, action KeyAction) error {
	keyCode, ok := d.specialMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}

	switch action {
	case KeyOn:
		d.send([]string{keyCode[0]})
	case KeyOff:
		d.send([]string{keyCode[1]})
	case KeyPress:
		d.send(keyCode)
	}
	return nil
}

func (d *pcATDriver) send(codes []string) {
	d.buffer = append(d.buffer, codes)
}

func chunkScanCodes(sc [][]string, size int) (out [][]string, err error) {
	var running []string
	for _, codes := range sc {
		if size > 0 {
			if len(codes) > size {
				return nil, fmt.Errorf("chunkScanCodes: size cannot be smaller than sc width.")
			}
			if len(running)+len(codes) > size {
				out = append(out, running)
				running = nil
			}
		}
		running = append(running, codes...)
	}
	if running != nil {
		out = append(out, running)
	}
	return
}
