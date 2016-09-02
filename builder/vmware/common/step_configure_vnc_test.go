package common

import (
	"fmt"
	"testing"
)

func TestStepConfigureVNC_implVNCAddressFinder(t *testing.T) {
	var _ VNCAddressFinder = new(StepConfigureVNC)
}

func TestStepConfigureVNC_UpdateVMX(t *testing.T) {
	var s StepConfigureVNC
	data := make(map[string]string)
	s.UpdateVMX("0.0.0.0", "", 5900, data)
	if ip := data["remotedisplay.vnc.ip"]; ip != "0.0.0.0" {
		t.Errorf("bad VMX data for key remotedisplay.vnc.ip: %v", ip)
	}
	if enabled := data["remotedisplay.vnc.enabled"]; enabled != "TRUE" {
		t.Errorf("bad VMX data for key remotedisplay.vnc.enabled: %v", enabled)
	}
	if port := data["remotedisplay.vnc.port"]; port != fmt.Sprint(port) {
		t.Errorf("bad VMX data for key remotedisplay.vnc.port: %v", port)
	}
}
