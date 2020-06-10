package yandexexport

import (
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestPostProcessor_Configure(t *testing.T) {
	type fields struct {
		config Config
		runner multistep.Runner
	}
	type args struct {
		raws []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "no one creds",
			fields: fields{
				config: Config{
					Token:                 "",
					ServiceAccountKeyFile: "",
				},
			},
			wantErr: true,
		},
		{
			name: "both token and sa key file",
			fields: fields{
				config: Config{
					Token:                 "some-value",
					ServiceAccountKeyFile: "path/not-exist.file",
				},
			},
			wantErr: true,
		},
		{
			name: "use sa key file",
			fields: fields{
				config: Config{
					Token:                 "",
					ServiceAccountKeyFile: "testdata/fake-sa-key.json",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.config.Paths = []string{"some-path"} // make Paths not empty
			p := &PostProcessor{
				config: tt.fields.config,
				runner: tt.fields.runner,
			}
			if err := p.Configure(tt.args.raws...); (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
