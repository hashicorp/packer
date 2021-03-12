package addrs

import (
	"reflect"
	"testing"
)

func TestParsePluginSourceString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		args      args
		want      *Plugin
		wantDiags bool
	}{
		{args{"potato"}, nil, true},
		{args{"hashicorp/azr"}, nil, true},
		{args{"github.com/hashicorp/azr"}, &Plugin{"github.com", "hashicorp", "azr"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.args.str, func(t *testing.T) {
			got, gotDiags := ParsePluginSourceString(tt.args.str)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePluginSourceString() got = %v, want %v", got, tt.want)
			}
			if tt.wantDiags == (len(gotDiags) == 0) {
				t.Errorf("Unexpected diags %s", gotDiags)
			}
		})
	}
}
