package common

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/rpc"
)

// This is the arrow from packer/ui.go -> TargetedUI.prefixLines
const targetedUIArrowText = "==>"

// The ProgressBar interface is used for abstracting cheggaaa's progress-
// bar, or any other progress bar. If a UI does not support a progress-
// bar, then it must return a null progress bar.
const (
	DefaultProgressBarWidth = 80
)

type ProgressBar = *pb.ProgressBar

// Figure out the terminal dimensions and use it to calculate the available rendering space
func calculateProgressBarWidth(length int) int {
	// If the UI's width is signed, then this is an interface that doesn't really benefit from a progress bar
	if length < 0 {
		log.Println("Refusing to render progress-bar for unsupported UI.")
		return length
	}

	// Figure out the terminal width if possible
	width, _, err := GetTerminalDimensions()
	if err != nil {
		newerr := fmt.Errorf("Unable to determine terminal dimensions: %v", err)
		log.Printf("Using default width (%d) for progress-bar due to error: %s", DefaultProgressBarWidth, newerr)
		return DefaultProgressBarWidth
	}

	// If the terminal width is smaller than the requested length, then complain
	if width < length {
		newerr := fmt.Errorf("Terminal width (%d) is smaller than UI message width (%d).", width, length)
		log.Printf("Using default width (%d) for progress-bar due to error: %s", DefaultProgressBarWidth, newerr)
		return DefaultProgressBarWidth
	}

	// Otherwise subtract the minimum length and return it
	return width - length
}

// Get a progress bar with the default appearance
func GetDefaultProgressBar() ProgressBar {
	bar := pb.New64(0)
	bar.ShowPercent = true
	bar.ShowCounters = true
	bar.ShowSpeed = false
	bar.ShowBar = true
	bar.ShowTimeLeft = false
	bar.ShowFinalTime = false
	bar.SetUnits(pb.U_BYTES)
	bar.Format("[=>-]")
	bar.SetRefreshRate(5 * time.Second)
	return bar
}

// Return a dummy progress bar that doesn't do anything
func GetDummyProgressBar() ProgressBar {
	bar := pb.New64(0)
	bar.ManualUpdate = true
	return bar
}

// Given a packer.Ui, calculate the number of characters that a packer.Ui will
// prefix a message with. Then we can use this to calculate the progress bar's width.
func calculateUiPrefixLength(ui packer.Ui) int {
	var recursiveCalculateUiPrefixLength func(packer.Ui, int) int

	// Define a recursive closure that traverses through all the known packer.Ui types
	// and aggregates the length of the message prefix from each particular type
	recursiveCalculateUiPrefixLength = func(ui packer.Ui, agg int) int {
		switch ui.(type) {

		case *packer.ColoredUi:
			// packer.ColoredUi is simply a wrapper around .Ui
			u := ui.(*packer.ColoredUi)
			return recursiveCalculateUiPrefixLength(u.Ui, agg)

		case *packer.TargetedUI:
			// A TargetedUI adds the .Target and an arrow by default
			u := ui.(*packer.TargetedUI)
			res := fmt.Sprintf("%s %s: ", targetedUIArrowText, u.Target)
			return recursiveCalculateUiPrefixLength(u.Ui, agg+len(res))

		case *packer.BasicUi:
			// The standard BasicUi appends only a newline
			return agg + len("\n")

		// packer.rpc.Ui returns 0 here to trigger the hack described later
		case *rpc.Ui:
			return 0

		case *packer.MachineReadableUi:
			// MachineReadableUi doesn't emit anything...like at all
			return 0
		}

		log.Printf("Calculating the message prefix length for packer.Ui type (%T) is not implemented. Using the current aggregated length of %d.", ui, agg)
		return agg
	}
	return recursiveCalculateUiPrefixLength(ui, 0)
}

func GetPackerConfigFromStateBag(state multistep.StateBag) *PackerConfig {
	config := state.Get("config")
	rConfig := reflect.Indirect(reflect.ValueOf(config))
	iPackerConfig := rConfig.FieldByName("PackerConfig").Interface()
	packerConfig := iPackerConfig.(PackerConfig)
	return &packerConfig
}

func GetProgressBar(ui packer.Ui, config *PackerConfig) ProgressBar {
	// Figure out the prefix length by quering the UI
	uiPrefixLength := calculateUiPrefixLength(ui)

	// hack to deal with packer.rpc.Ui courtesy of @Swampdragons
	if _, ok := ui.(*rpc.Ui); uiPrefixLength == 0 && config != nil && ok {
		res := fmt.Sprintf("%s %s: \n", targetedUIArrowText, config.PackerBuildName)
		uiPrefixLength = len(res)
	}

	// Now we can use the prefix length to calculate the progress bar width
	width := calculateProgressBarWidth(uiPrefixLength)

	log.Printf("ProgressBar: Using progress bar width: %d\n", width)

	// Get a default progress bar and set some output defaults
	bar := GetDefaultProgressBar()
	bar.SetWidth(width)
	bar.Callback = func(message string) {
		ui.Message(message)
	}
	return bar
}
