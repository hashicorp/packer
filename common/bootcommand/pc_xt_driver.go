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
type scMap map[string]*scancode

type pcXTDriver struct {
	interval    time.Duration
	sendImpl    SendCodeFunc
	specialMap  scMap
	scancodeMap map[rune]byte
	buffer      [][]string
	// TODO: set from env
	scancodeChunkSize int
}

type scancode struct {
	make   []string
	break_ []string
}

func (sc *scancode) makeBreak() []string {
	return append(sc.make, sc.break_...)
}

// NewPCXTDriver creates a new boot command driver for VMs that expect PC-XT
// keyboard codes. `send` should send its argument to the VM. `chunkSize` should
// be the maximum number of keyboard codes to send to `send` at one time.
func NewPCXTDriver(send SendCodeFunc, chunkSize int, interval time.Duration) *pcXTDriver {
	// We delay (default 100ms) between each input event to allow for CPU or
	// network latency. See PackerKeyEnv for tuning.
	keyInterval := common.PackerKeyDefault
	if delay, err := time.ParseDuration(os.Getenv(common.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}
	// Override interval based on builder-specific override
	if interval > time.Duration(0) {
		keyInterval = interval
	}
	// Scancodes reference: https://www.win.tue.nl/~aeb/linux/kbd/scancodes-1.html
	//						https://www.win.tue.nl/~aeb/linux/kbd/scancodes-10.html
	//
	// Scancodes are recorded here in pairs. The first entry represents
	// the key press and the second entry represents the key release and is
	// derived from the first by the addition of 0x80.
	sMap := make(scMap)
	sMap["bs"] = &scancode{[]string{"0e"}, []string{"8e"}}
	sMap["del"] = &scancode{[]string{"e0", "53"}, []string{"e0", "d3"}}
	sMap["down"] = &scancode{[]string{"e0", "50"}, []string{"e0", "d0"}}
	sMap["end"] = &scancode{[]string{"e0", "4f"}, []string{"e0", "cf"}}
	sMap["enter"] = &scancode{[]string{"1c"}, []string{"9c"}}
	sMap["esc"] = &scancode{[]string{"01"}, []string{"81"}}
	sMap["f1"] = &scancode{[]string{"3b"}, []string{"bb"}}
	sMap["f2"] = &scancode{[]string{"3c"}, []string{"bc"}}
	sMap["f3"] = &scancode{[]string{"3d"}, []string{"bd"}}
	sMap["f4"] = &scancode{[]string{"3e"}, []string{"be"}}
	sMap["f5"] = &scancode{[]string{"3f"}, []string{"bf"}}
	sMap["f6"] = &scancode{[]string{"40"}, []string{"c0"}}
	sMap["f7"] = &scancode{[]string{"41"}, []string{"c1"}}
	sMap["f8"] = &scancode{[]string{"42"}, []string{"c2"}}
	sMap["f9"] = &scancode{[]string{"43"}, []string{"c3"}}
	sMap["f10"] = &scancode{[]string{"44"}, []string{"c4"}}
	sMap["f11"] = &scancode{[]string{"57"}, []string{"d7"}}
	sMap["f12"] = &scancode{[]string{"58"}, []string{"d8"}}
	sMap["home"] = &scancode{[]string{"e0", "47"}, []string{"e0", "c7"}}
	sMap["insert"] = &scancode{[]string{"e0", "52"}, []string{"e0", "d2"}}
	sMap["left"] = &scancode{[]string{"e0", "4b"}, []string{"e0", "cb"}}
	sMap["leftalt"] = &scancode{[]string{"38"}, []string{"b8"}}
	sMap["leftctrl"] = &scancode{[]string{"1d"}, []string{"9d"}}
	sMap["leftshift"] = &scancode{[]string{"2a"}, []string{"aa"}}
	sMap["leftsuper"] = &scancode{[]string{"e0", "5b"}, []string{"e0", "db"}}
	sMap["menu"] = &scancode{[]string{"e0", "5d"}, []string{"e0", "dd"}}
	sMap["pagedown"] = &scancode{[]string{"e0", "51"}, []string{"e0", "d1"}}
	sMap["pageup"] = &scancode{[]string{"e0", "49"}, []string{"e0", "c9"}}
	sMap["return"] = &scancode{[]string{"1c"}, []string{"9c"}}
	sMap["right"] = &scancode{[]string{"e0", "4d"}, []string{"e0", "cd"}}
	sMap["rightalt"] = &scancode{[]string{"e0", "38"}, []string{"e0", "b8"}}
	sMap["rightctrl"] = &scancode{[]string{"e0", "1d"}, []string{"e0", "9d"}}
	sMap["rightshift"] = &scancode{[]string{"36"}, []string{"b6"}}
	sMap["rightsuper"] = &scancode{[]string{"e0", "5c"}, []string{"e0", "dc"}}
	sMap["spacebar"] = &scancode{[]string{"39"}, []string{"b9"}}
	sMap["tab"] = &scancode{[]string{"0f"}, []string{"8f"}}
	sMap["up"] = &scancode{[]string{"e0", "48"}, []string{"e0", "c8"}}

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

	return &pcXTDriver{
		interval:          keyInterval,
		sendImpl:          send,
		specialMap:        sMap,
		scancodeMap:       scancodeMap,
		scancodeChunkSize: chunkSize,
	}
}

// Flush send all scanecodes.
func (d *pcXTDriver) Flush() error {
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

func (d *pcXTDriver) SendKey(key rune, action KeyAction) error {
	keyShift := unicode.IsUpper(key) || strings.ContainsRune(shiftedChars, key)

	var sc []string

	if action&(KeyOn|KeyPress) != 0 {
		scInt := d.scancodeMap[key]
		if keyShift {
			sc = append(sc, "2a")
		}
		sc = append(sc, fmt.Sprintf("%02x", scInt))
	}

	if action&(KeyOff|KeyPress) != 0 {
		scInt := d.scancodeMap[key] + 0x80
		if keyShift {
			sc = append(sc, "aa")
		}
		sc = append(sc, fmt.Sprintf("%02x", scInt))
	}

	log.Printf("Sending char '%c', code '%s', shift %v",
		key, strings.Join(sc, ""), keyShift)

	d.send(sc)
	return nil
}

func (d *pcXTDriver) SendSpecial(special string, action KeyAction) error {
	keyCode, ok := d.specialMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}
	log.Printf("Special code '%s' '<%s>' found, replacing with: %v", action.String(), special, keyCode)

	switch action {
	case KeyOn:
		d.send(keyCode.make)
	case KeyOff:
		d.send(keyCode.break_)
	case KeyPress:
		d.send(keyCode.makeBreak())
	}
	return nil
}

// send stores the codes in an internal buffer. Use Flush to send them.
func (d *pcXTDriver) send(codes []string) {
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
