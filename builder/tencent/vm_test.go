package tencent

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateImage_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepCreateImage
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateImage.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepCreateImage_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepCreateImage
		args args
	}{
		// No cleanup, so what test???
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepStopImage_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepStopImage
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepStopImage.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateCustomImage_Run(t *testing.T) {
	state := &multistep.BasicStateBag{}

	requiredEnvVars := map[string]string{
		CRegion: "ap-singapore",
	}
	GetRequiredEnvVars(requiredEnvVars)
	config := &Config{
		ImageName:       "ccw-2", // what if an existing ID is specified???
		Region:          requiredEnvVars[CRegion],
		SecretID:        requiredEnvVars[CSecretId],
		SecretKey:       requiredEnvVars[CSecretKey],
		ImageIdLocation: `D:\Development\Packer\PackerBuilderTencent\ImageId.txt`,
		Timeout:         300000,
		Url:             CCVMUrlSiliconValley,
	}

	config.PackerDebug = true
	state.Put(CConfig, config)

	driver := new(TencentDriver)
	driver.Ui = NewPackerUi()
	state.Put(CDriver, driver)

	state.Put("ui", driver.Ui)
	step := &StepCreateCustomImage{}
	step.Run(nil, state)
}

func TestStepStopImage_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepStopImage
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepRunImage_Run(t *testing.T) {
}

func TestStepRunImage_Cleanup(t *testing.T) {
}

func TestStepWaitRunning_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepWaitRunning
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepWaitRunning.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepWaitRunning_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepWaitRunning
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepWaitStopped_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepWaitStopped
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepWaitStopped.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepWaitStopped_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepWaitStopped
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepCreateKeyPair_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepCreateKeyPair
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateKeyPair.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepCreateKeyPair_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepCreateKeyPair
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepGetImageStatus_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepGetImageStatus
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepGetImageStatus.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepGetImageStatus_Cleanup(t *testing.T) {
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepGetImageStatus
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Cleanup(tt.args.state)
		})
	}
}

func TestStepGetKeyPairStatus_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepGetKeyPairStatus
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepGetKeyPairStatus.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepGetKeyPairStatus_Cleanup(t *testing.T) {
}

func TestStepClear_Run(t *testing.T) {
	type args struct {
		in0   context.Context
		state multistep.StateBag
	}
	tests := []struct {
		name string
		s    *StepClear
		args args
		want multistep.StepAction
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Run(tt.args.in0, tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepClear.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStepCreateCustomImage_Run(t *testing.T) {
}

func TestStepGetInstanceIP_Run(t *testing.T) {
	state := &multistep.BasicStateBag{}

	requiredEnvVars := map[string]string{
		CRegion: "ap-singapore",
	}
	GetRequiredEnvVars(requiredEnvVars)
	config := &Config{
		ImageName:       "ccw-3",
		Region:          requiredEnvVars[CRegion],
		SecretID:        requiredEnvVars[CSecretId],
		SecretKey:       requiredEnvVars[CSecretKey],
		ImageIdLocation: `D:\Development\Packer\ImageToStop.txt`,
		Timeout:         300000,
		Url:             CCVMUrlSiliconValley,
	}
	state.Put(CInstanceId, "ins-0594sts4")

	config.PackerDebug = true
	state.Put(CConfig, config)

	driver := new(TencentDriver)
	state.Put(CDriver, driver)

	ui := NewPackerUi()
	state.Put("ui", ui)
	step := &StepGetInstanceIP{}
	step.Run(nil, state)
}

func TestStepHalt_Run(t *testing.T) {
}

func TestStepDisplayMessage_Run(t *testing.T) {
}
