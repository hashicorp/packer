package classic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPVConfigEntry(t *testing.T) {
	entry := 1
	var entryTests = []struct {
		imageList      string
		imageListEntry *int
		expected       int
	}{
		{"x", nil, 0},
		{"x", &entry, 1},
		{imageListDefault, nil, 5},
		{imageListDefault, &entry, 1},
	}
	for _, tt := range entryTests {
		tc := &PVConfig{
			PersistentVolumeSize:  1,
			BuilderImageList:      tt.imageList,
			BuilderImageListEntry: tt.imageListEntry,
		}
		errs := tc.Prepare(nil)
		assert.Nil(t, errs, "Didn't expect any errors")
		assert.Equal(t, tt.expected, *tc.BuilderImageListEntry)
	}
}
