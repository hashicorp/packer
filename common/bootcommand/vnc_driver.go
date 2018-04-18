package bootcommand

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/packer/common"
)

const KeyLeftShift uint32 = 0xFFE1

type VNCKeyEvent interface {
	KeyEvent(uint32, bool) error
}

type vncDriver struct {
	c          VNCKeyEvent
	interval   time.Duration
	specialMap map[string]uint32
	// keyEvent can set this error which will prevent it from continuing
	err error
}

func NewVNCDriver(c VNCKeyEvent) *vncDriver {
	// We delay (default 100ms) between each key event to allow for CPU or
	// network latency. See PackerKeyEnv for tuning.
	keyInterval := common.PackerKeyDefault
	if delay, err := time.ParseDuration(os.Getenv(common.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}

	// Scancodes reference: https://github.com/qemu/qemu/blob/master/ui/vnc_keysym.h
	sMap := make(map[string]uint32)
	sMap["bs"] = 0xFF08
	sMap["del"] = 0xFFFF
	sMap["down"] = 0xFF54
	sMap["end"] = 0xFF57
	sMap["enter"] = 0xFF0D
	sMap["esc"] = 0xFF1B
	sMap["f1"] = 0xFFBE
	sMap["f2"] = 0xFFBF
	sMap["f3"] = 0xFFC0
	sMap["f4"] = 0xFFC1
	sMap["f5"] = 0xFFC2
	sMap["f6"] = 0xFFC3
	sMap["f7"] = 0xFFC4
	sMap["f8"] = 0xFFC5
	sMap["f9"] = 0xFFC6
	sMap["f10"] = 0xFFC7
	sMap["f11"] = 0xFFC8
	sMap["f12"] = 0xFFC9
	sMap["home"] = 0xFF50
	sMap["insert"] = 0xFF63
	sMap["left"] = 0xFF51
	sMap["leftalt"] = 0xFFE9
	sMap["leftctrl"] = 0xFFE3
	sMap["leftshift"] = 0xFFE1
	sMap["leftsuper"] = 0xFFEB
	sMap["menu"] = 0xFF67
	sMap["pagedown"] = 0xFF56
	sMap["pageup"] = 0xFF55
	sMap["return"] = 0xFF0D
	sMap["right"] = 0xFF53
	sMap["rightalt"] = 0xFFEA
	sMap["rightctrl"] = 0xFFE4
	sMap["rightshift"] = 0xFFE2
	sMap["rightsuper"] = 0xFFEC
	sMap["spacebar"] = 0x020
	sMap["tab"] = 0xFF09
	sMap["up"] = 0xFF52

	return &vncDriver{
		c:          c,
		interval:   keyInterval,
		specialMap: sMap,
	}
}

func (d *vncDriver) keyEvent(k uint32, down bool) error {
	if d.err != nil {
		return nil
	}
	if err := d.c.KeyEvent(k, down); err != nil {
		d.err = err
		return err
	}
	time.Sleep(d.interval)
	return nil
}

// Finalize does nothing here
func (d *vncDriver) Finalize() error {
	return nil
}

func (d *vncDriver) SendKey(key rune, action KeyAction) error {
	keyShift := unicode.IsUpper(key) || strings.ContainsRune(shiftedChars, key)
	keyCode := uint32(key)
	log.Printf("Sending char '%c', code 0x%X, shift %v", key, keyCode, keyShift)

	switch action {
	case KeyOn:
		if keyShift {
			d.keyEvent(KeyLeftShift, true)
		}
		d.keyEvent(keyCode, true)
	case KeyOff:
		if keyShift {
			d.keyEvent(KeyLeftShift, false)
		}
		d.keyEvent(keyCode, false)
	case KeyPress:
		if keyShift {
			d.keyEvent(KeyLeftShift, true)
		}
		d.keyEvent(keyCode, true)
		d.keyEvent(keyCode, false)
		if keyShift {
			d.keyEvent(KeyLeftShift, false)
		}
	}
	return d.err
}

func (d *vncDriver) SendSpecial(special string, action KeyAction) error {
	keyCode, ok := d.specialMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}
	log.Printf("Special code '<%s>' found, replacing with: 0x%X", special, keyCode)

	switch action {
	case KeyOn:
		d.keyEvent(keyCode, true)
	case KeyOff:
		d.keyEvent(keyCode, false)
	case KeyPress:
		d.keyEvent(keyCode, true)
		d.keyEvent(keyCode, false)
	}

	return d.err
}
