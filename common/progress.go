package common

import "github.com/cheggaaa/pb"

// Default progress bar appearance
func GetDefaultProgressBar() pb.ProgressBar {
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

	return *bar
}
