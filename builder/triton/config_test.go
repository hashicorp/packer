package triton

import (
	"testing"
)

func testConfig(t *testing.T) Config {
	return Config{
		AccessConfig:        testAccessConfig(t),
		SourceMachineConfig: testSourceMachineConfig(t),
		TargetImageConfig:   testTargetImageConfig(t),
	}
}
