package qemu

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"
)

func TestStepResizeDisk_Skips(t *testing.T) {
	testConfigs := []*Config{
		&Config{
			DiskImage:      false,
			SkipResizeDisk: false,
		},
		&Config{
			DiskImage:      false,
			SkipResizeDisk: true,
		},
	}
	for _, config := range testConfigs {
		state := testState(t)
		driver := state.Get("driver").(*DriverMock)

		state.Put("config", config)
		step := new(stepResizeDisk)

		// Test the run
		if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v", action)
		}
		if _, ok := state.GetOk("error"); ok {
			t.Fatal("should NOT have error")
		}
		if len(driver.QemuImgCalls) > 0 {
			t.Fatal("should NOT have called qemu-img")
		}
	}
}

func Test_buildResizeCommand(t *testing.T) {
	type testCase struct {
		Step     *stepResizeDisk
		Expected []string
		Reason   string
	}
	testcases := []testCase{
		{
			&stepResizeDisk{
				Format:   "qcow2",
				DiskSize: "1234M",
			},
			[]string{"resize", "-f", "qcow2", "source.qcow", "1234M"},
			"no extra args",
		},
		{
			&stepResizeDisk{
				Format:   "qcow2",
				DiskSize: "1234M",
				QemuImgArgs: QemuImgArgs{
					Resize: []string{"-foo", "bar"},
				},
			},
			[]string{"resize", "-f", "qcow2", "-foo", "bar", "source.qcow", "1234M"},
			"one set of extra args",
		},
	}

	for _, tc := range testcases {
		command := tc.Step.buildResizeCommand("source.qcow")

		assert.Equal(t, command, tc.Expected,
			fmt.Sprintf("%s. Expected %#v", tc.Reason, tc.Expected))
	}
}
