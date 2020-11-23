package yandex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	data1 = `
#cloud-config
bootcmd:
  - cmd1
  - cmd2
`
	data2 = `
#cloud-config
runcmd:
  - touch "cmd3"
  - cmd4
`
)

func TestCloudInitMerge(t *testing.T) {
	merged, err := MergeCloudUserMetaData(
		data1,
		data2,
	)

	require.NoError(t, err)
	require.NotEmpty(t, merged)

	require.Contains(t, merged, "cmd1")
	require.Contains(t, merged, "cmd2")
	require.Contains(t, merged, "\"cmd3\"")
	require.Contains(t, merged, "cmd4")

}
