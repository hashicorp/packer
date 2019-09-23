package vminstance

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	Platform      = "platform"
	GuestOSType   = "guestOSType"
	BackupStorage = "backupStorage"
	Vm            = "vmInstance"
	Image         = "image"
	DataVolume    = "volume"
	ExportPath    = "exportPath"
	DeviceId      = "deviceId"
	MountPath     = "/builder"
)

func getSizeFromStr(size string) (uint64, error) {
	suffix := size[len(size)-1:]
	value, err := strconv.ParseUint(size[0:len(size)-1], 10, 64)
	if err != nil {
		return 0, err
	}
	switch suffix {
	case "k":
		return value * 1024, nil
	case "m":
		return value * 1024 * 1024, nil
	case "g":
		return value * 1024 * 1024 * 1024, nil
	case "t":
		return value * 1024 * 1024 * 1024 * 1024, nil
	case "p":
		return value * 1024 * 1024 * 1024 * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("only support 'k', 'm', 'g', 't', 'p' as the suffix for datavolume_size")
	}
}

func halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func toSlice(arr interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("toslice arr not slice")
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret, nil
}

func GetCommonFromState(state multistep.StateBag) (driver Driver, config Config, ui packer.Ui) {
	driver = state.Get("driver").(Driver)
	config = state.Get("config").(Config)
	ui = state.Get("ui").(packer.Ui)
	return
}
