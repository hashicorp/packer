package common

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSetFirstBootDevice struct {
	Generation      uint
	FirstBootDevice string
}

func ParseBootDeviceIdentifier(deviceIdentifier string, generation uint) (string, uint, uint, error) {

	captureExpression := "^(FLOPPY|IDE|NET)|(CD|DVD)$"
	if generation > 1 {
		captureExpression = "^((IDE|SCSI):(\\d+):(\\d+))|(DVD|CD)|(NET)$"
	}

	r, err := regexp.Compile(captureExpression)
	if err != nil {
		return "", 0, 0, err
	}

	// match against the appropriate set of values.. we force to uppercase to ensure that
	// all devices are always in the same case

	identifierMatches := r.FindStringSubmatch(strings.ToUpper(deviceIdentifier))
	if identifierMatches == nil {
		return "", 0, 0, fmt.Errorf("The value %q is not a properly formatted device or device group identifier.", deviceIdentifier)
	}

	switch {

	// CD or DVD are always returned as "CD"
	case ((generation == 1) && (identifierMatches[2] != "")) || ((generation > 1) && (identifierMatches[5] != "")):
		return "CD", 0, 0, nil

	// generation 1 only has FLOPPY, IDE or NET remaining..
	case (generation == 1):
		return identifierMatches[0], 0, 0, nil

	// generation 2, check for IDE or SCSI and parse location and number
	case (identifierMatches[2] != ""):
		{

			var controllerLocation int64
			var controllerNumber int64

			// NOTE: controllerNumber and controllerLocation cannot be negative, the regex expression
			// would not have matched if either number was signed

			controllerNumber, err = strconv.ParseInt(identifierMatches[3], 10, 8)
			if err == nil {

				controllerLocation, err = strconv.ParseInt(identifierMatches[4], 10, 8)
				if err == nil {

					return identifierMatches[2], uint(controllerNumber), uint(controllerLocation), nil

				}

			}

			return "", 0, 0, err

		}

	// only "NET" left on generation 2
	default:
		return "NET", 0, 0, nil

	}

}

func (s *StepSetFirstBootDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
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
