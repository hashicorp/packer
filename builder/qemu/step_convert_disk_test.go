package qemu

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildConvertCommand(t *testing.T) {
	type testCase struct {
		Step     *stepConvertDisk
		Expected []string
		Reason   string
	}
	testcases := []testCase{
		{
			&stepConvertDisk{
				Format:          "qcow2",
				DiskCompression: false,
			},
			[]string{"convert", "-O", "qcow2", "source.qcow", "target.qcow2"},
			"Basic, happy path, no compression, no extra args",
		},
		{
			&stepConvertDisk{
				Format:          "qcow2",
				DiskCompression: true,
			},
			[]string{"convert", "-c", "-O", "qcow2", "source.qcow", "target.qcow2"},
			"Basic, happy path, with compression, no extra args",
		},
		{
			&stepConvertDisk{
				Format:          "qcow2",
				DiskCompression: true,
				QemuImgArgs: QemuImgArgs{
					Convert: []string{"-o", "preallocation=full"},
				},
			},
			[]string{"convert", "-c", "-o", "preallocation=full", "-O", "qcow2", "source.qcow", "target.qcow2"},
			"Basic, happy path, with compression, one set of extra args",
		},
	}

	for _, tc := range testcases {
		command := tc.Step.buildConvertCommand("source.qcow", "target.qcow2")

		assert.Equal(t, command, tc.Expected,
			fmt.Sprintf("%s. Expected %#v", tc.Reason, tc.Expected))
	}
}
