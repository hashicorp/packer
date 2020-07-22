package yandeximport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArtifactState_StateData(t *testing.T) {
	expectedData := "this is the data"
	artifact := &Artifact{
		StateData: map[string]interface{}{"state_data": expectedData},
	}

	// Valid state
	result := artifact.State("state_data")
	require.Equal(t, expectedData, result)

	// Invalid state
	result = artifact.State("invalid_key")
	require.Equal(t, nil, result)

	// Nil StateData should not fail and should return nil
	artifact = &Artifact{}
	result = artifact.State("key")
	require.Equal(t, nil, result)
}
