package common

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

func TestHWConfigPrepare(t *testing.T) {
	c := new(HWConfig)

	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if c.CpuCount < 0 {
		t.Errorf("bad cpu count: %d", c.CpuCount)
	}

	if c.CoreCount < 0 {
		t.Errorf("bad core count: %d", c.CoreCount)
	}

	if c.MemorySize < 0 {
		t.Errorf("bad memory size: %d", c.MemorySize)
	}

	if c.Sound {
		t.Errorf("peripheral choice (sound) should be conservative: %t", c.Sound)
	}

	if c.USB {
		t.Errorf("peripheral choice (usb) should be conservative: %t", c.USB)
	}

	if strings.ToUpper(c.Parallel) != "NONE" {
		t.Errorf("parallel port should not be defined: %s", c.Parallel)
	}

	if strings.ToUpper(c.Serial) != "NONE" {
		t.Errorf("serial port should not be defined: %s", c.Serial)
	}
}

func TestHWConfigParallel_File(t *testing.T) {
	c := new(HWConfig)

	c.Parallel = "file:filename"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasParallel() {
		t.Errorf("parallel port should be defined")
	}

	parallel, err := c.ReadParallel()
	if err != nil {
		t.Fatalf("Unable to read parallel port definition: %s", err)
	}

	switch parallel.Union.(type) {
	case *ParallelPortFile:
		break
	default:
		t.Errorf("parallel port should be a file type")
	}

	if parallel.File.Filename != "filename" {
		t.Errorf("parallel port filename should be \"filename\": %s", parallel.File.Filename)
	}
}

func TestHWConfigParallel_Device(t *testing.T) {
	c := new(HWConfig)

	c.Parallel = "device:devicename,uni"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasParallel() {
		t.Errorf("parallel port should be defined")
	}

	parallel, err := c.ReadParallel()
	if err != nil {
		t.Fatalf("Unable to read parallel port definition: %s", err)
	}

	switch parallel.Union.(type) {
	case *ParallelPortDevice:
		break
	default:
		t.Errorf("parallel port should be a device type")
	}

	if strings.ToLower(parallel.Device.Bidirectional) != "false" {
		t.Errorf("parallel port device should not be bidirectional: %s", parallel.Device.Bidirectional)
	}

	if parallel.Device.Devicename != "devicename" {
		t.Errorf("parallel port device should be \"devicename\": %s", parallel.Device.Devicename)
	}
}

func TestHWConfigParallel_Auto(t *testing.T) {
	c := new(HWConfig)

	c.Parallel = "auto:bi"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasParallel() {
		t.Errorf("parallel port should be defined")
	}

	parallel, err := c.ReadParallel()
	if err != nil {
		t.Fatalf("Unable to read parallel port definition: %s", err)
	}

	switch parallel.Union.(type) {
	case *ParallelPortAuto:
		break
	default:
		t.Errorf("parallel port should be an auto type")
	}

	if strings.ToLower(parallel.Auto.Bidirectional) != "true" {
		t.Errorf("parallel port device should be bidirectional: %s", parallel.Auto.Bidirectional)
	}
}

func TestHWConfigParallel_None(t *testing.T) {
	c := new(HWConfig)

	c.Parallel = "none"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasParallel() {
		t.Errorf("parallel port should be defined")
	}

	parallel, err := c.ReadParallel()
	if err != nil {
		t.Fatalf("Unable to read parallel port definition: %s", err)
	}

	if parallel.Union != nil {
		t.Errorf("parallel port shouldn't exist")
	}
}

func TestHWConfigSerial_File(t *testing.T) {
	c := new(HWConfig)

	c.Serial = "file:filename,true"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasSerial() {
		t.Errorf("serial port should be defined")
	}

	serial, err := c.ReadSerial()
	if err != nil {
		t.Fatalf("Unable to read serial port definition: %s", err)
	}

	switch serial.Union.(type) {
	case *SerialConfigFile:
		break
	default:
		t.Errorf("serial port should be a file type")
	}

	if serial.File.Filename != "filename" {
		t.Errorf("serial port filename should be \"filename\": %s", serial.File.Filename)
	}

	if strings.ToLower(serial.File.Yield) != "true" {
		t.Errorf("serial port yield should be true: %s", serial.File.Yield)
	}
}

func TestHWConfigSerial_Device(t *testing.T) {
	c := new(HWConfig)

	c.Serial = "device:devicename,true"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasSerial() {
		t.Errorf("serial port should be defined")
	}

	serial, err := c.ReadSerial()
	if err != nil {
		t.Fatalf("Unable to read serial port definition: %s", err)
	}

	switch serial.Union.(type) {
	case *SerialConfigDevice:
		break
	default:
		t.Errorf("serial port should be a device type")
	}

	if serial.Device.Devicename != "devicename" {
		t.Errorf("serial port device should be \"devicename\": %s", serial.Device.Devicename)
	}

	if strings.ToLower(serial.Device.Yield) != "true" {
		t.Errorf("serial port device should yield: %s", serial.Device.Yield)
	}
}

func TestHWConfigSerial_Pipe(t *testing.T) {
	c := new(HWConfig)

	c.Serial = "pipe:mypath,client,app,true"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasSerial() {
		t.Errorf("serial port should be defined")
	}

	serial, err := c.ReadSerial()
	if err != nil {
		t.Fatalf("Unable to read serial port definition: %s", err)
	}

	switch serial.Union.(type) {
	case *SerialConfigPipe:
		break
	default:
		t.Errorf("serial port should be a pipe type")
	}

	if serial.Pipe.Filename != "mypath" {
		t.Errorf("serial port pipe name should be \"mypath\": %s", serial.Pipe.Filename)
	}

	if strings.ToLower(serial.Pipe.Endpoint) != "client" {
		t.Errorf("serial port endpoint should be \"client\": %s", serial.Pipe.Endpoint)
	}

	if strings.ToLower(serial.Pipe.Host) != "true" {
		t.Errorf("serial port host type for app should be true: %s", serial.Pipe.Host)
	}

	if strings.ToLower(serial.Pipe.Yield) != "true" {
		t.Errorf("serial port should yield: %s", serial.Pipe.Yield)
	}
}

func TestHWConfigSerial_Auto(t *testing.T) {
	c := new(HWConfig)

	c.Serial = "auto:true"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasSerial() {
		t.Errorf("serial port should be defined")
	}

	serial, err := c.ReadSerial()
	if err != nil {
		t.Fatalf("Unable to read serial port definition: %s", err)
	}

	switch serial.Union.(type) {
	case *SerialConfigAuto:
		break
	default:
		t.Errorf("serial port should be an auto type")
	}

	if strings.ToLower(serial.Auto.Yield) != "true" {
		t.Errorf("serial port should yield: %s", serial.Auto.Yield)
	}
}

func TestHWConfigSerial_None(t *testing.T) {
	c := new(HWConfig)

	c.Serial = "none"
	if errs := c.Prepare(interpolate.NewContext()); len(errs) > 0 {
		t.Fatalf("err: %#v", errs)
	}

	if !c.HasSerial() {
		t.Errorf("serial port should be defined")
	}

	serial, err := c.ReadSerial()
	if err != nil {
		t.Fatalf("Unable to read serial port definition: %s", err)
	}

	if serial.Union != nil {
		t.Errorf("serial port shouldn't exist")
	}
}
