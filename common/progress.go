package common

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/hashicorp/packer/packer"
	"log"
	"time"
)

// Default progress bar appearance
func GetNewProgressBar(ui *packer.Ui) pb.ProgressBar {
	bar := pb.New64(0)
	bar.ShowPercent = true
	bar.ShowCounters = true
	bar.ShowSpeed = false
	bar.ShowBar = true
	bar.ShowTimeLeft = false
	bar.ShowFinalTime = false
	bar.SetUnits(pb.U_BYTES)
	bar.Format("[=>-]")
	bar.SetRefreshRate(1 * time.Second)
	bar.SetWidth(80)

	// If there's no UI set, then the progress bar doesn't need anything else
	if ui == nil {
		return *bar
	}
	UI := *ui

	// Now check the UI's width to adjust the progress bar
	uiWidth := UI.GetMinimumLength() + len("\n")

	// If the UI's width is signed, then this interface doesn't really
	// benefit from a progress bar
	if uiWidth < 0 {
		log.Println("Refusing to render progress-bar for unsupported UI.")
		return *bar
	}
	bar.Callback = UI.Message

	// Figure out the terminal width if possible
	width, _, err := GetTerminalDimensions()
	if err != nil {
		newerr := fmt.Errorf("Unable to determine terminal dimensions: %v", err)
		log.Printf("Using default width (%d) for progress-bar due to error: %s", bar.GetWidth(), newerr)
		return *bar
	}

	// Adjust the progress bar's width according to the terminal size
	// and whatever width is returned from the UI
	if width > uiWidth {
		width -= uiWidth
		bar.SetWidth(width)
	} else {
		newerr := fmt.Errorf("Terminal width (%d) is smaller than UI message width (%d).", width, uiWidth)
		log.Printf("Using default width (%d) for progress-bar due to error: %s", bar.GetWidth(), newerr)
	}

	return *bar
}
