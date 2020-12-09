package yandexexport

import (
	"testing"

	"github.com/hashicorp/packer/builder/yandex"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
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
					AccessConfig: yandex.AccessConfig{
						Token:                 "",
						ServiceAccountKeyFile: "",
					},
					ExchangeConfig: ExchangeConfig{
						ServiceAccountID: "some-srv-acc-id",
					},
					CommonConfig: yandex.CommonConfig{
						CloudConfig: yandex.CloudConfig{
							FolderID: "some-folder-id",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "both token and sa key file",
			fields: fields{
				config: Config{
					AccessConfig: yandex.AccessConfig{
						Token:                 "some-value",
						ServiceAccountKeyFile: "path/not-exist.file",
					},
					ExchangeConfig: ExchangeConfig{
						ServiceAccountID: "some-srv-acc-id",
					},
					CommonConfig: yandex.CommonConfig{
						CloudConfig: yandex.CloudConfig{
							FolderID: "some-folder-id",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "use sa key file",
			fields: fields{
				config: Config{
					AccessConfig: yandex.AccessConfig{
						Token:                 "",
						ServiceAccountKeyFile: "testdata/fake-sa-key.json",
					},
					ExchangeConfig: ExchangeConfig{
						ServiceAccountID: "some-srv-acc-id",
					},
					CommonConfig: yandex.CommonConfig{
						CloudConfig: yandex.CloudConfig{
							FolderID: "some-folder-id",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "service_account_id required",
			fields: fields{
				config: Config{
					AccessConfig: yandex.AccessConfig{
						Token: "some token",
					},
					ExchangeConfig: ExchangeConfig{
						ServiceAccountID: "",
					},
					CommonConfig: yandex.CommonConfig{
						CloudConfig: yandex.CloudConfig{
							FolderID: "some-folder-id",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "folderID required",
			fields: fields{
				config: Config{
					AccessConfig: yandex.AccessConfig{
						Token: "some token",
					},
					ExchangeConfig: ExchangeConfig{
						ServiceAccountID: "some-srv-acc-id",
					},
					CommonConfig: yandex.CommonConfig{
						CloudConfig: yandex.CloudConfig{
							FolderID: "",
						},
					},
				},
			},
			wantErr: true,
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

func Test_formUrls(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
	}{
		{
			name: "empty list",
			args: args{
				paths: []string{},
			},
			wantResult: []string{},
		},
		{
			name: "one element",
			args: args{
				paths: []string{"s3://bucket1/object1"},
			},
			wantResult: []string{"https://" + defaultStorageEndpoint + "/bucket1/object1"},
		},
		{
			name: "several elements",
			args: args{
				paths: []string{
					"s3://bucket1/object1",
					"s3://bucket-name/object-with/prefix/filename.blob",
					"s3://bucket-too/foo/bar.test",
				},
			},
			wantResult: []string{
				"https://" + defaultStorageEndpoint + "/bucket1/object1",
				"https://" + defaultStorageEndpoint + "/bucket-name/object-with/prefix/filename.blob",
				"https://" + defaultStorageEndpoint + "/bucket-too/foo/bar.test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantResult, formUrls(tt.args.paths))
		})
	}
}
