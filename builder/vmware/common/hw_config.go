package common

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/template/interpolate"
)

type HWConfig struct {

	// cpu information
	CpuCount   int `mapstructure:"cpus"`
	MemorySize int `mapstructure:"memory"`
	CoreCount  int `mapstructure:"cores"`

	// network type and adapter
	Network            string `mapstructure:"network"`
	NetworkAdapterType string `mapstructure:"network_adapter_type"`

	// device presence
	Sound bool `mapstructure:"sound"`
	USB   bool `mapstructure:"usb"`

	// communication ports
	Serial   string `mapstructure:"serial"`
	Parallel string `mapstructure:"parallel"`
}

func (c *HWConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	// Hardware and cpu options
	if c.CpuCount < 0 {
		errs = append(errs, fmt.Errorf("An invalid number of cpus was specified (cpus < 0): %d", c.CpuCount))
	}

	if c.MemorySize < 0 {
		errs = append(errs, fmt.Errorf("An invalid amount of memory was specified (memory < 0): %d", c.MemorySize))
	}

	// Hardware and cpu options
	if c.CoreCount < 0 {
		errs = append(errs, fmt.Errorf("An invalid number of cores was specified (cores < 0): %d", c.CoreCount))
	}

	// Peripherals
	if !c.Sound {
		c.Sound = false
	}

	if !c.USB {
		c.USB = false
	}

	if c.Parallel == "" {
		c.Parallel = "none"
	}

	if c.Serial == "" {
		c.Serial = "none"
	}

	return errs
}

/* parallel port */
type ParallelUnion struct {
	Union  interface{}
	File   *ParallelPortFile
	Device *ParallelPortDevice
	Auto   *ParallelPortAuto
}
type ParallelPortFile struct {
	Filename string
}
type ParallelPortDevice struct {
	Bidirectional string
	Devicename    string
}
type ParallelPortAuto struct {
	Bidirectional string
}

func (c *HWConfig) HasParallel() bool {
	return c.Parallel != ""
}

func (c *HWConfig) ReadParallel() (*ParallelUnion, error) {
	input := strings.SplitN(c.Parallel, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("Unexpected format for parallel port: %s", c.Parallel)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "FILE":
		res := &ParallelPortFile{Filename: filepath.FromSlash(formatOptions)}
		return &ParallelUnion{Union: res, File: res}, nil
	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 1 || len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for parallel port: %s", c.Parallel)
		}
		res := new(ParallelPortDevice)
		res.Bidirectional = "FALSE"
		res.Devicename = filepath.FromSlash(comp[0])
		if len(comp) > 1 {
			switch strings.ToUpper(comp[1]) {
			case "BI":
				res.Bidirectional = "TRUE"
			case "UNI":
				res.Bidirectional = "FALSE"
			default:
				return nil, fmt.Errorf("Unknown direction %s specified for parallel port: %s", strings.ToUpper(comp[1]), c.Parallel)
			}
		}
		return &ParallelUnion{Union: res, Device: res}, nil

	case "AUTO":
		res := new(ParallelPortAuto)
		switch strings.ToUpper(formatOptions) {
		case "":
			fallthrough
		case "UNI":
			res.Bidirectional = "FALSE"
		case "BI":
			res.Bidirectional = "TRUE"
		default:
			return nil, fmt.Errorf("Unknown direction %s specified for parallel port: %s", strings.ToUpper(formatOptions), c.Parallel)
		}
		return &ParallelUnion{Union: res, Auto: res}, nil

	case "NONE":
		return &ParallelUnion{Union: nil}, nil
	}

	return nil, fmt.Errorf("Unexpected format for parallel port: %s", c.Parallel)
}

/* serial conversions */
type SerialConfigPipe struct {
	Filename string
	Endpoint string
	Host     string
	Yield    string
}

type SerialConfigFile struct {
	Filename string
	Yield    string
}

type SerialConfigDevice struct {
	Devicename string
	Yield      string
}

type SerialConfigAuto struct {
	Devicename string
	Yield      string
}

type SerialUnion struct {
	Union  interface{}
	Pipe   *SerialConfigPipe
	File   *SerialConfigFile
	Device *SerialConfigDevice
	Auto   *SerialConfigAuto
}

func (c *HWConfig) HasSerial() bool {
	return c.Serial != ""
}

func (c *HWConfig) ReadSerial() (*SerialUnion, error) {
	var defaultSerialPort string
	if runtime.GOOS == "windows" {
		defaultSerialPort = "COM1"
	} else {
		defaultSerialPort = "/dev/ttyS0"
	}

	input := strings.SplitN(c.Serial, ":", 2)
	if len(input) < 1 {
		return nil, fmt.Errorf("Unexpected format for serial port: %s", c.Serial)
	}

	var formatType, formatOptions string
	formatType = input[0]
	if len(input) == 2 {
		formatOptions = input[1]
	} else {
		formatOptions = ""
	}

	switch strings.ToUpper(formatType) {
	case "PIPE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) < 3 || len(comp) > 4 {
			return nil, fmt.Errorf("Unexpected format for serial port pipe: %s", c.Serial)
		}
		if res := strings.ToLower(comp[1]); res != "client" && res != "server" {
			return nil, fmt.Errorf("Unexpected format for endpoint in serial port pipe: %s -> %s", c.Serial, res)
		}
		if res := strings.ToLower(comp[2]); res != "app" && res != "vm" {
			return nil, fmt.Errorf("Unexpected format for host in serial port pipe: %s -> %s", c.Serial, res)
		}
		res := &SerialConfigPipe{
			Filename: comp[0],
			Endpoint: comp[1],
			Host:     map[string]string{"app": "TRUE", "vm": "FALSE"}[strings.ToLower(comp[2])],
			Yield:    "FALSE",
		}
		if len(comp) == 4 {
			res.Yield = strings.ToUpper(comp[3])
		}
		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for yield in serial port pipe: %s -> %s", c.Serial, res.Yield)
		}
		return &SerialUnion{Union: res, Pipe: res}, nil

	case "FILE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for serial port file: %s", c.Serial)
		}

		res := &SerialConfigFile{Yield: "FALSE"}

		res.Filename = filepath.FromSlash(comp[0])

		res.Yield = map[bool]string{true: strings.ToUpper(comp[1]), false: "FALSE"}[len(comp) > 1]
		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for yield in serial port file: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, File: res}, nil

	case "DEVICE":
		comp := strings.Split(formatOptions, ",")
		if len(comp) > 2 {
			return nil, fmt.Errorf("Unexpected format for serial port device: %s", c.Serial)
		}

		res := new(SerialConfigDevice)

		if len(comp) == 2 {
			res.Devicename = map[bool]string{true: filepath.FromSlash(comp[0]), false: defaultSerialPort}[len(comp[0]) > 0]
			res.Yield = strings.ToUpper(comp[1])
		} else if len(comp) == 1 {
			res.Devicename = map[bool]string{true: filepath.FromSlash(comp[0]), false: defaultSerialPort}[len(comp[0]) > 0]
			res.Yield = "FALSE"
		} else if len(comp) == 0 {
			res.Devicename = defaultSerialPort
			res.Yield = "FALSE"
		}

		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for yield in serial port device: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, Device: res}, nil

	case "AUTO":
		res := new(SerialConfigAuto)
		res.Devicename = defaultSerialPort

		if len(formatOptions) > 0 {
			res.Yield = strings.ToUpper(formatOptions)
		} else {
			res.Yield = "FALSE"
		}

		if res.Yield != "TRUE" && res.Yield != "FALSE" {
			return nil, fmt.Errorf("Unexpected format for yield in serial port auto: %s -> %s", c.Serial, res.Yield)
		}

		return &SerialUnion{Union: res, Auto: res}, nil

	case "NONE":
		return &SerialUnion{Union: nil}, nil

	default:
		return nil, fmt.Errorf("Unknown serial type %s: %s", strings.ToUpper(formatType), c.Serial)
	}
}
