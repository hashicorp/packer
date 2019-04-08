package proxmox

import (
	"fmt"
	"time"
	"unicode"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/common/bootcommand"
)

type proxmoxDriver struct {
	client     commandTyper
	vmRef      *proxmox.VmRef
	specialMap map[string]string
	runeMap    map[rune]string
	interval   time.Duration
}

func NewProxmoxDriver(c commandTyper, vmRef *proxmox.VmRef, interval time.Duration) *proxmoxDriver {
	// Mappings for packer shorthand to qemu qkeycodes
	sMap := map[string]string{
		"spacebar": "spc",
		"bs":       "backspace",
		"del":      "delete",
		"return":   "ret",
		"enter":    "ret",
		"pageUp":   "pgup",
		"pageDown": "pgdn",
	}
	// Mappings for runes that need to be translated to special qkeycodes
	// Taken from https://github.com/qemu/qemu/blob/master/pc-bios/keymaps/en-us
	rMap := map[rune]string{
		// Clean mappings
		' ':  "spc",
		'.':  "dot",
		',':  "comma",
		';':  "semicolon",
		'*':  "asterisk",
		'-':  "minus",
		'[':  "bracket_left",
		']':  "bracket_right",
		'=':  "equal",
		'\'': "apostrophe",
		'`':  "grave_accent",
		'/':  "slash",
		'\\': "backslash",

		'!': "shift-1",             // "exclam"
		'@': "shift-2",             // "at"
		'#': "shift-3",             // "numbersign"
		'$': "shift-4",             // "dollar"
		'%': "shift-5",             // "percent"
		'^': "shift-6",             // "asciicircum"
		'&': "shift-7",             // "ampersand"
		'(': "shift-9",             // "parenleft"
		')': "shift-0",             // "parenright"
		'{': "shift-bracket_left",  // "braceleft"
		'}': "shift-bracket_right", // "braceright"
		'"': "shift-apostrophe",    // "quotedbl"
		'+': "shift-equal",         // "plus"
		'_': "shift-minus",         // "underscore"
		':': "shift-semicolon",     // "colon"
		'<': "shift-comma",         // "less" is recognized, but seem to map to '/'?
		'>': "shift-dot",           // "greater"
		'~': "shift-grave_accent",  // "asciitilde"
		'?': "shift-slash",         // "question"
		'|': "shift-backslash",     // "bar"
	}

	return &proxmoxDriver{
		client:     c,
		vmRef:      vmRef,
		specialMap: sMap,
		runeMap:    rMap,
		interval:   interval,
	}
}

func (p *proxmoxDriver) SendKey(key rune, action bootcommand.KeyAction) error {
	if special, ok := p.runeMap[key]; ok {
		return p.send(special)
	}

	var keys string
	if unicode.IsUpper(key) {
		keys = fmt.Sprintf("shift-%c", unicode.ToLower(key))
	} else {
		keys = fmt.Sprintf("%c", key)
	}

	return p.send(keys)
}

func (p *proxmoxDriver) SendSpecial(special string, action bootcommand.KeyAction) error {
	keys := special
	if replacement, ok := p.specialMap[special]; ok {
		keys = replacement
	}

	return p.send(keys)
}

func (p *proxmoxDriver) send(keys string) error {
	res, err := p.client.MonitorCmd(p.vmRef, "sendkey "+keys)
	if err != nil {
		return err
	}
	if data, ok := res["data"].(string); ok && len(data) > 0 {
		return fmt.Errorf("failed to send keys: %s", data)
	}

	time.Sleep(p.interval)
	return nil
}

func (p *proxmoxDriver) Flush() error { return nil }
