package common

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step compacts the virtual disk for the VM unless the "skip_compaction"
// boolean is true.
//
// Uses:
//   driver Driver
//   disk_full_paths ([]string) - The full paths to all created disks
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type StepCompactDisk struct {
	Skip bool
}

func (s StepCompactDisk) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	diskFullPaths := state.Get("disk_full_paths").([]string)

	if s.Skip {
		log.Println("Skipping disk compaction step...")
		return multistep.ActionContinue
	}

	ui.Say("Compacting all attached virtual disks...")
	for i, diskFullPath := range diskFullPaths {
		ui.Message(fmt.Sprintf("Compacting virtual disk %d", i+1))
		// Get the file size of the virtual disk prior to compaction
		fi, err := os.Stat(diskFullPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error getting virtual disk file info pre compaction: %s", err))
			return multistep.ActionHalt
		}
		diskFileSizeStart := fi.Size()
		// Defragment and compact the disk
		if err := driver.CompactDisk(diskFullPath); err != nil {
			state.Put("error", fmt.Errorf("Error compacting disk: %s", err))
			return multistep.ActionHalt
		}
		// Get the file size of the virtual disk post compaction
		fi, err = os.Stat(diskFullPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error getting virtual disk file info post compaction: %s", err))
			return multistep.ActionHalt
		}
		diskFileSizeEnd := fi.Size()
		// Report compaction results
		log.Printf("Before compaction the disk file size was: %d", diskFileSizeStart)
		log.Printf("After compaction the disk file size was: %d", diskFileSizeEnd)
		if diskFileSizeStart > 0 {
			percentChange := ((float64(diskFileSizeEnd) / float64(diskFileSizeStart)) * 100.0) - 100.0
			switch {
			case percentChange < 0:
				ui.Message(fmt.Sprintf("Compacting reduced the disk file size by %.2f%%", math.Abs(percentChange)))
			case percentChange == 0:
				ui.Message(fmt.Sprintf("The compacting operation left the disk file size unchanged"))
			case percentChange > 0:
				ui.Message(fmt.Sprintf("WARNING: Compacting increased the disk file size by %.2f%%", percentChange))
			}
		}
	}

	return multistep.ActionContinue
}

func (StepCompactDisk) Cleanup(multistep.StateBag) {}
