package packer

import (
	"os"

	"github.com/hashicorp/packer-plugin-sdk/rpc"
)

// MayUseProtobuf is meant to look into the environment to prohibit protobuf
//
// If the PACKER_USE_PB environment variable is unset or set to a non-empty
// string that is neither "0", "no", or "false", Packer will choose dynamically
// which protocol to use when communicating with plugins.
//
// If however it is explicitly set to one of the false values, Packer will not
// attempt to detect which protocol to use, and instead will forcibly use gob.
func MayUseProtobuf() bool {
	usePB := os.Getenv(rpc.PackerUsePBEnvVar)
	switch usePB {
	case "0", "no", "false":
		return false
	}

	return true
}
