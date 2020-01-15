package driver

import (
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/mobile/event/key"
	"strings"
	"unicode"
)

type KeyInput struct {
	Message  string
	Scancode key.Code
	Alt      bool
	Ctrl     bool
	Shift    bool
}

var scancodeMap = make(map[rune]key.Code)

func init() {
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

	for chars, start := range scancodeIndex {
		for i, r := range chars {
			scancodeMap[r] = start + key.Code(i)
		}
	}
}

const shiftedChars = "!@#$%^&*()_+{}|:\"~<>?"

func (vm *VirtualMachine) TypeOnKeyboard(input KeyInput) (int32, error) {
	var spec types.UsbScanCodeSpec

	for _, r := range input.Message {
		scancode := scancodeMap[r]
		shift := input.Shift || unicode.IsUpper(r) || strings.ContainsRune(shiftedChars, r)

		spec.KeyEvents = append(spec.KeyEvents, types.UsbScanCodeSpecKeyEvent{
			// https://github.com/lamw/vghetto-scripts/blob/f74bc8ba20064f46592bcce5a873b161a7fa3d72/powershell/VMKeystrokes.ps1#L130
			UsbHidCode: int32(scancode)<<16 | 7,
			Modifiers: &types.UsbScanCodeSpecModifierType{
				LeftControl: &input.Ctrl,
				LeftAlt:     &input.Alt,
				LeftShift:   &shift,
			},
		})
	}

	if input.Scancode != 0 {
		spec.KeyEvents = append(spec.KeyEvents, types.UsbScanCodeSpecKeyEvent{
			UsbHidCode: int32(input.Scancode)<<16 | 7,
			Modifiers: &types.UsbScanCodeSpecModifierType{
				LeftControl: &input.Ctrl,
				LeftAlt:     &input.Alt,
				LeftShift:   &input.Shift,
			},
		})
	}

	req := &types.PutUsbScanCodes{
		This: vm.vm.Reference(),
		Spec: spec,
	}

	resp, err := methods.PutUsbScanCodes(vm.driver.ctx, vm.driver.client.RoundTripper, req)
	if err != nil {
		return 0, err
	}

	return resp.Returnval, nil
}
