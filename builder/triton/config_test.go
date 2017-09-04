package triton

import (
	"testing"
)

func testConfig(t *testing.T) Config {
	return Config{
		AccessConfig:        testAccessConfig(),
		SourceMachineConfig: testSourceMachineConfig(t),
		TargetImageConfig:   testTargetImageConfig(t),
	}
}
