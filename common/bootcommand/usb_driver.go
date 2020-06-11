package bootcommand

import (
	"fmt"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/common"
	"golang.org/x/mobile/event/key"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
)

// SendUsbScanCodes will be called to send codes to the VM
type SendUsbScanCodes func([]key.Code, []bool) error

type usbDriver struct {
	vm *driver.VirtualMachine

	sendImpl    SendUsbScanCodes
	interval    time.Duration
	specialMap  map[string]key.Code
	scancodeMap map[rune]key.Code

	codeBuffer []key.Code
	downBuffer []bool

	// keyEvent can set this error which will prevent it from continuing
	err error
}

func NewUSBDriver(send SendUsbScanCodes, interval time.Duration) *usbDriver {
	// We delay (default 100ms) between each key event to allow for CPU or
	// network latency. See PackerKeyEnv for tuning.
	keyInterval := common.PackerKeyDefault
	if delay, err := time.ParseDuration(os.Getenv(common.PackerKeyEnv)); err == nil {
		keyInterval = delay
	}
	// override interval based on builder-specific override.
	if interval > time.Duration(0) {
		keyInterval = interval
	}

	special := map[string]key.Code{
		"enter":      key.CodeReturnEnter,
		"esc":        key.CodeEscape,
		"bs":         key.CodeDeleteBackspace,
		"del":        key.CodeDeleteForward,
		"tab":        key.CodeTab,
		"f1":         key.CodeF1,
		"f2":         key.CodeF2,
		"f3":         key.CodeF3,
		"f4":         key.CodeF4,
		"f5":         key.CodeF5,
		"f6":         key.CodeF6,
		"f7":         key.CodeF7,
		"f8":         key.CodeF8,
		"f9":         key.CodeF9,
		"f10":        key.CodeF10,
		"f11":        key.CodeF11,
		"f12":        key.CodeF12,
		"insert":     key.CodeInsert,
		"home":       key.CodeHome,
		"end":        key.CodeEnd,
		"pageUp":     key.CodePageUp,
		"pageDown":   key.CodePageDown,
		"left":       key.CodeLeftArrow,
		"right":      key.CodeRightArrow,
		"up":         key.CodeUpArrow,
		"down":       key.CodeDownArrow,
		"leftalt":    key.CodeLeftAlt,
		"leftctrl":   key.CodeLeftControl,
		"leftshift":  key.CodeLeftShift,
		"rightalt":   key.CodeRightAlt,
		"rightctrl":  key.CodeRightControl,
		"rightshift": key.CodeRightShift,
	}

	scancodeIndex := make(map[string]key.Code)
	scancodeIndex["abcdefghijklmnopqrstuvwxyz"] = key.CodeA
	scancodeIndex["ABCDEFGHIJKLMNOPQRSTUVWXYZ"] = key.CodeA
	scancodeIndex["1234567890"] = key.Code1
	scancodeIndex["!@#$%^&*()"] = key.Code1
	scancodeIndex[" "] = key.CodeSpacebar
	scancodeIndex["-=[]\\"] = key.CodeHyphenMinus
	scancodeIndex["_+{}|"] = key.CodeHyphenMinus
	scancodeIndex[";'`,./"] = key.CodeSemicolon
	scancodeIndex[":\"~<>?"] = key.CodeSemicolon

	var scancodeMap = make(map[rune]key.Code)
	for chars, start := range scancodeIndex {
		for i, r := range chars {
			scancodeMap[r] = start + key.Code(i)
		}
	}

	return &usbDriver{
		sendImpl:    send,
		specialMap:  special,
		interval:    keyInterval,
		scancodeMap: scancodeMap,
	}
}

//func (d *usbDriver) keyEvent(k key.Code, down bool) error {
//	if d.err != nil {
//		return nil
//	}
//	if err := d.sendImpl(k, down); err != nil {
//		d.err = err
//		return err
//	}
//	//time.Sleep(d.interval)
//	return nil
//}

// Flush does nothing here
func (d *usbDriver) Flush() error {
	defer func() {
		d.codeBuffer = nil
	}()

	if err := d.sendImpl(d.codeBuffer, d.downBuffer); err != nil {
		return err
	}
	time.Sleep(d.interval)
	return nil
}

func (d *usbDriver) SendKey(k rune, action KeyAction) error {
	keyShift := unicode.IsUpper(k) || strings.ContainsRune(shiftedChars, k)
	keyCode := d.scancodeMap[k]
	log.Printf("Sending char '%c', code %s, shift %v", k, keyCode, keyShift)

	switch action {
	case KeyOn:
		if keyShift {
			d.send(key.CodeLeftShift, true)
		}
		d.send(keyCode, true)
	case KeyOff:
		if keyShift {
			d.send(key.CodeLeftShift, false)
		}
		d.send(keyCode, false)
	case KeyPress:
		if keyShift {
			d.send(key.CodeLeftShift, true)
		}
		d.send(keyCode, true)
		d.send(keyCode, false)
		if keyShift {
			d.send(key.CodeLeftShift, false)
		}
	}
	return d.err
}

func (d *usbDriver) SendSpecial(special string, action KeyAction) error {
	keyCode, ok := d.specialMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}
	log.Printf("Special code '<%s>' found, replacing with: %s", special, keyCode)

	switch action {
	case KeyOn:
		d.send(keyCode, true)
	case KeyOff:
		d.send(keyCode, false)
	case KeyPress:
		d.send(keyCode, true)
		d.send(keyCode, false)
	}

	return d.err
}

// send stores the codes in an internal buffer. Use Flush to send them.
func (d *usbDriver) send(code key.Code, down bool) {
	// slices to keep the input order
	d.codeBuffer = append(d.codeBuffer, code)
	d.downBuffer = append(d.downBuffer, down)

}
