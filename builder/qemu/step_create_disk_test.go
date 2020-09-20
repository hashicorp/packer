package qemu

import (
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/stretchr/testify/assert"
)

func Test_buildCreateCommand(t *testing.T) {
	type testCase struct {
		Step     *stepCreateDisk
		I        int
		Expected []string
		Reason   string
	}
	testcases := []testCase{
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: false,
			},
			0,
			[]string{"create", "-f", "qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, no backing store, no extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
			},
			0,
			[]string{"create", "-f", "qcow2", "-b", "source.qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, backing store, no extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
			},
			1,
			[]string{"create", "-f", "qcow2", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set but not at first index, no extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
				QemuImgArgs: QemuImgArgs{
					Create: []string{"-foo", "bar"},
				},
			},
			0,
			[]string{"create", "-f", "qcow2", "-b", "source.qcow2", "-foo", "bar", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set, extra args",
		},
		{
			&stepCreateDisk{
				Format:         "qcow2",
				UseBackingFile: true,
				QemuImgArgs: QemuImgArgs{
					Create: []string{"-foo", "bar"},
				},
			},
			1,
			[]string{"create", "-f", "qcow2", "-foo", "bar", "target.qcow2", "1234M"},
			"Basic, happy path, backing store set but not at first index, extra args",
		},
	}

	for _, tc := range testcases {
		state := new(multistep.BasicStateBag)
		state.Put("iso_path", "source.qcow2")
		command := tc.Step.buildCreateCommand("target.qcow2", "1234M", tc.I, state)

		assert.Equal(t, command, tc.Expected,
			fmt.Sprintf("%s. Expected %#v", tc.Reason, tc.Expected))
	}
}
