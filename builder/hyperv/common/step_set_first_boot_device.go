package common

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSetFirstBootDevice struct {
	Generation      uint
	FirstBootDevice string
}

func ParseBootDeviceIdentifier(deviceIdentifier string, generation uint) (string, uint, uint, error) {

	// all input strings are forced to upperCase for comparison, I believe this is
	// safe as all of our values are 7bit ASCII clean.

	lookupDeviceIdentifier := strings.ToUpper(deviceIdentifier)

	if generation == 1 {

		// Gen1 values are a simple set of if/then/else values, which we coalesce into a map
		// here for simplicity

		lookupTable := map[string]string{
			"FLOPPY": "FLOPPY",
			"IDE":    "IDE",
			"NET":    "NET",
			"CD":     "CD",
			"DVD":    "CD",
		}

		controllerType, isDefined := lookupTable[lookupDeviceIdentifier]
		if !isDefined {

			return "", 0, 0, fmt.Errorf("The value %q is not a properly formatted device group identifier.", deviceIdentifier)

		}

		// success
		return controllerType, 0, 0, nil
	}

	// everything else is treated as generation 2... the first set of lookups covers
	// the simple options..

	lookupTable := map[string]string{
		"CD":  "CD",
		"DVD": "CD",
		"NET": "NET",
	}

	controllerType, isDefined := lookupTable[lookupDeviceIdentifier]
	if isDefined {

		// these types do not require controllerNumber or controllerLocation
		return controllerType, 0, 0, nil

	}

	// not a simple option, check for a controllerType:controllerNumber:controllerLocation formatted
	// device..

	r, err := regexp.Compile(`^(IDE|SCSI):(\d+):(\d+)$`)
	if err != nil {
		return "", 0, 0, err
	}

	controllerMatch := r.FindStringSubmatch(lookupDeviceIdentifier)
	if controllerMatch != nil {

		var controllerLocation int64
		var controllerNumber int64

		// NOTE: controllerNumber and controllerLocation cannot be negative, the regex expression
		// would not have matched if either number was signed

		controllerNumber, err = strconv.ParseInt(controllerMatch[2], 10, 8)
		if err == nil {

			controllerLocation, err = strconv.ParseInt(controllerMatch[3], 10, 8)
			if err == nil {

				return controllerMatch[1], uint(controllerNumber), uint(controllerLocation), nil

			}

		}

		return "", 0, 0, err

	}

	return "", 0, 0, fmt.Errorf("The value %q is not a properly formatted device identifier.", deviceIdentifier)
}

func (s *StepSetFirstBootDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	if s.FirstBootDevice != "" {

		controllerType, controllerNumber, controllerLocation, err := ParseBootDeviceIdentifier(s.FirstBootDevice, s.Generation)
		if err == nil {

			switch {

			case controllerType == "CD":
				{
					// the "DVD" controller is special, we only apply the setting if we actually mounted
					// an ISO and only if that was mounted as the "IsoUrl" not a secondary ISO.

					dvdControllerState := state.Get("os.dvd.properties")
					if dvdControllerState == nil {

						ui.Say("First Boot Device is DVD, but no primary ISO mounted. Ignoring.")
						return multistep.ActionContinue

					}

					ui.Say(fmt.Sprintf("Setting boot device to %q", s.FirstBootDevice))
					dvdController := dvdControllerState.(DvdControllerProperties)
					err = driver.SetFirstBootDevice(vmName, controllerType, dvdController.ControllerNumber, dvdController.ControllerLocation, s.Generation)

				}

			default:
				{
					// anything else, we just pass as is..
					ui.Say(fmt.Sprintf("Setting boot device to %q", s.FirstBootDevice))
					err = driver.SetFirstBootDevice(vmName, controllerType, controllerNumber, controllerLocation, s.Generation)
				}
			}

		}

		if err != nil {
			err := fmt.Errorf("Error setting first boot device: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt

		}

	}

	return multistep.ActionContinue
}

func (s *StepSetFirstBootDevice) Cleanup(state multistep.StateBag) {
	// do nothing
}
