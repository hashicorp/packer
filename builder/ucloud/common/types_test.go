package common

import (
	"testing"
)

func Test_parseInstanceType(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name    string
		args    args
		want    *InstanceType
		wantErr bool
	}{
		{"ok_highcpu", args{"n-highcpu-1"}, &InstanceType{1, 1024, "n", "highcpu"}, false},
		{"ok_basic", args{"n-basic-1"}, &InstanceType{1, 2048, "n", "basic"}, false},
		{"ok_standard", args{"n-standard-1"}, &InstanceType{1, 4096, "n", "standard"}, false},
		{"ok_highmem", args{"n-highmem-1"}, &InstanceType{1, 8192, "n", "highmem"}, false},
		{"ok_customized", args{"n-customized-1-12"}, &InstanceType{1, 12288, "n", "customized"}, false},

		{"err_customized", args{"n-customized-1-5"}, nil, true},
		{"err_type", args{"nx-highcpu-1"}, nil, true},
		{"err_scale_type", args{"n-invalid-1"}, nil, true},
		{"err_cpu_too_much", args{"n-highcpu-33"}, nil, true},
		{"err_cpu_too_less", args{"n-highcpu-0"}, nil, true},
		{"err_cpu_is_invalid", args{"n-highcpu-x"}, nil, true},
		{"err_customized_format_len", args{"n-customized-1"}, nil, true},
		{"err_customized_format_number", args{"n-customized-x"}, nil, true},
		{"err_customized_should_be_standard", args{"n-customized-1-2"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInstanceType(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInstanceType() arg %s got %#v error = %v, wantErr %v", tt.args.s, got, err, tt.wantErr)
				return
			}

			if got == nil {
				return
			}

			if !(tt.want.CPU == got.CPU) ||
				!(tt.want.Memory == got.Memory) ||
				!(tt.want.HostType == got.HostType) ||
				!(tt.want.HostScaleType == got.HostScaleType) {
				t.Errorf("ParseInstanceType() = %v, want %v", got, tt.want)
			}
		})
	}
}
