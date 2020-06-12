package bootcommand

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/packer/common"
	"golang.org/x/mobile/event/key"
)

// SendUsbScanCodes will be called to send codes to the VM
type SendUsbScanCodes func(k key.Code, down bool) error

type usbDriver struct {
	sendImpl    SendUsbScanCodes
	interval    time.Duration
	specialMap  map[string]key.Code
	scancodeMap map[rune]key.Code
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
		"return":     key.CodeReturnEnter,
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
		"leftsuper":  key.CodeLeftGUI,
		"rightsuper": key.CodeRightGUI,
		"spacebar":   key.CodeSpacebar,
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

func (d *usbDriver) keyEvent(k key.Code, down bool) error {
	if err := d.sendImpl(k, down); err != nil {
		return err
	}
	time.Sleep(d.interval)
	return nil
}

func (d *usbDriver) Flush() error {
	return nil
}

func (d *usbDriver) SendKey(k rune, action KeyAction) error {
	keyShift := unicode.IsUpper(k) || strings.ContainsRune(shiftedChars, k)
	keyCode := d.scancodeMap[k]
	log.Printf("Sending char '%c', code %s, shift %v", k, keyCode, keyShift)
	return d.keyEvent(keyCode, keyShift)
}

func (d *usbDriver) SendSpecial(special string, action KeyAction) (err error) {
	keyCode, ok := d.specialMap[special]
	if !ok {
		return fmt.Errorf("special %s not found.", special)
	}
	log.Printf("Special code '<%s>' found, replacing with: %s", special, keyCode)

	switch action {
	case KeyOn:
		err = d.keyEvent(keyCode, true)
	case KeyOff, KeyPress:
		err = d.keyEvent(keyCode, false)
	}

	return err
}
