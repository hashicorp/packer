package hcl2template

import (
	"testing"

	awscommon "github.com/hashicorp/packer/builder/amazon/common"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/packer/helper/communicator"

	amazonebs "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/builder/virtualbox/iso"

	"github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"

	amazon_import "github.com/hashicorp/packer/post-processor/amazon-import"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func getBasicParser() *Parser {
	return &Parser{
		Parser: hclparse.NewParser(),
		ProvisionersSchemas: map[string]Decodable{
			"shell": &shell.Config{},
			"file":  &file.Config{},
		},
		PostProvisionersSchemas: map[string]Decodable{
			"amazon-import": &amazon_import.Config{},
		},
		CommunicatorSchemas: map[string]Decodable{
			"ssh":   &communicator.SSH{},
			"winrm": &communicator.WinRM{},
		},
		SourceSchemas: map[string]Decodable{
			"amazon-ebs":     &amazonebs.Config{},
			"virtualbox-iso": &iso.Config{},
		},
	}
}

func TestParser_ParseFile(t *testing.T) {
	defaultParser := getBasicParser()

	type fields struct {
		Parser *hclparse.Parser
	}
	type args struct {
		filename string
		cfg      *PackerConfig
	}
	tests := []struct {
		name             string
		parser           *Parser
		args             args
		wantPackerConfig *PackerConfig
		wantDiags        bool
	}{
		{
			"valid " + sourceLabel + " load",
			defaultParser,
			args{"testdata/sources/basic.pkr.hcl", new(PackerConfig)},
			&PackerConfig{
				Sources: map[SourceRef]*Source{
					SourceRef{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
						Cfg: &iso.FlatConfig{
							HTTPDir:         strPtr("xxx"),
							ISOChecksum:     strPtr("769474248a3897f4865817446f9a4a53"),
							ISOChecksumType: strPtr("md5"),
							RawSingleISOUrl: strPtr("http://releases.ubuntu.com/12.04/ubuntu-12.04.5-server-amd64.iso"),
							BootCommand:     []string{"..."},
							ShutdownCommand: strPtr("echo 'vagrant' | sudo -S shutdown -P now"),
							BootWait:        strPtr("10s"),
							VBoxManage:      [][]string{},
							VBoxManagePost:  [][]string{},
						},
					},
					SourceRef{
						Type: "amazon-ebs",
						Name: "ubuntu-1604",
					}: {
						Type: "amazon-ebs",
						Name: "ubuntu-1604",
						Cfg: &amazonebs.FlatConfig{
							RawRegion:            strPtr("eu-west-3"),
							AMIEncryptBootVolume: boolPtr(true),
							InstanceType:         strPtr("t2.micro"),
							SourceAmiFilter: &awscommon.FlatAmiFilterOptions{
								Filters: map[string]string{
									"name":                "ubuntu/images/*ubuntu-xenial-{16.04}-amd64-server-*",
									"root-device-type":    "ebs",
									"virtualization-type": "hvm",
								},
								Owners: []string{"099720109477"},
							},
							AMIMappings:    []awscommon.FlatBlockDevice{},
							LaunchMappings: []awscommon.FlatBlockDevice{},
						},
					},
					SourceRef{
						Type: "amazon-ebs",
						Name: "that-ubuntu-1.0",
					}: {
						Type: "amazon-ebs",
						Name: "that-ubuntu-1.0",
						Cfg: &amazonebs.FlatConfig{
							RawRegion:            strPtr("eu-west-3"),
							AMIEncryptBootVolume: boolPtr(true),
							InstanceType:         strPtr("t2.micro"),
							SourceAmiFilter: &awscommon.FlatAmiFilterOptions{
								MostRecent: boolPtr(true),
							},
							AMIMappings:    []awscommon.FlatBlockDevice{},
							LaunchMappings: []awscommon.FlatBlockDevice{},
						},
					},
				},
			},
			false,
		},

		{
			"valid " + communicatorLabel + " load",
			defaultParser,
			args{"testdata/communicator/basic.pkr.hcl", new(PackerConfig)},
			&PackerConfig{
				Communicators: map[CommunicatorRef]*Communicator{
					{Type: "ssh", Name: "vagrant"}: {
						Type: "ssh", Name: "vagrant",
						Cfg: &communicator.FlatSSH{
							SSHUsername:               strPtr("vagrant"),
							SSHPassword:               strPtr("s3cr4t"),
							SSHClearAuthorizedKeys:    boolPtr(true),
							SSHHost:                   strPtr("sssssh.hashicorp.io"),
							SSHHandshakeAttempts:      intPtr(32),
							SSHPort:                   intPtr(42),
							SSHFileTransferMethod:     strPtr("scp"),
							SSHPrivateKeyFile:         strPtr("file.pem"),
							SSHPty:                    boolPtr(false),
							SSHTimeout:                strPtr("5m"),
							SSHAgentAuth:              boolPtr(false),
							SSHDisableAgentForwarding: boolPtr(true),
							SSHBastionHost:            strPtr(""),
							SSHBastionPort:            intPtr(0),
							SSHBastionAgentAuth:       boolPtr(true),
							SSHBastionUsername:        strPtr(""),
							SSHBastionPassword:        strPtr(""),
							SSHBastionPrivateKeyFile:  strPtr(""),
							SSHProxyHost:              strPtr("ninja-potatoes.com"),
							SSHProxyPort:              intPtr(42),
							SSHProxyUsername:          strPtr("dark-father"),
							SSHProxyPassword:          strPtr("pickle-rick"),
							SSHKeepAliveInterval:      strPtr("10s"),
							SSHReadWriteTimeout:       strPtr("5m"),
						},
					},
				},
			},
			false,
		},

		{
			"duplicate " + sourceLabel, defaultParser,
			args{"testdata/sources/basic.pkr.hcl", &PackerConfig{
				Sources: map[SourceRef]*Source{
					SourceRef{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
						Cfg: &iso.FlatConfig{
							HTTPDir: strPtr("xxx"),
						},
					},
				},
			},
			},
			&PackerConfig{
				Sources: map[SourceRef]*Source{
					SourceRef{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
						Cfg: &iso.FlatConfig{
							HTTPDir: strPtr("xxx"),
						},
					},
					SourceRef{
						Type: "amazon-ebs",
						Name: "ubuntu-1604",
					}: {
						Type: "amazon-ebs",
						Name: "ubuntu-1604",
						Cfg: &amazonebs.FlatConfig{
							RawRegion:            strPtr("eu-west-3"),
							AMIEncryptBootVolume: boolPtr(true),
							InstanceType:         strPtr("t2.micro"),
							SourceAmiFilter: &awscommon.FlatAmiFilterOptions{
								Filters: map[string]string{
									"name":                "ubuntu/images/*ubuntu-xenial-{16.04}-amd64-server-*",
									"root-device-type":    "ebs",
									"virtualization-type": "hvm",
								},
								Owners: []string{"099720109477"},
							},
							AMIMappings:    []awscommon.FlatBlockDevice{},
							LaunchMappings: []awscommon.FlatBlockDevice{},
						},
					},
					SourceRef{
						Type: "amazon-ebs",
						Name: "that-ubuntu-1.0",
					}: {
						Type: "amazon-ebs",
						Name: "that-ubuntu-1.0",
						Cfg: &amazonebs.FlatConfig{
							RawRegion:            strPtr("eu-west-3"),
							AMIEncryptBootVolume: boolPtr(true),
							InstanceType:         strPtr("t2.micro"),
							SourceAmiFilter: &awscommon.FlatAmiFilterOptions{
								MostRecent: boolPtr(true),
							},
							AMIMappings:    []awscommon.FlatBlockDevice{},
							LaunchMappings: []awscommon.FlatBlockDevice{},
						},
					},
				},
			},
			true,
		},

		{"valid variables load", defaultParser,
			args{"testdata/variables/basic.pkr.hcl", new(PackerConfig)},
			&PackerConfig{
				Variables: PackerV1Variables{
					"image_name": "foo-image-{{user `my_secret`}}",
					"key":        "value",
					"my_secret":  "foo",
				},
			},
			false,
		},

		{"valid " + buildLabel + " load", defaultParser,
			args{"testdata/build/basic.pkr.hcl", new(PackerConfig)},
			&PackerConfig{
				Builds: Builds{
					{
						Froms: BuildFromList{
							{
								Src: SourceRef{"amazon-ebs", "ubuntu-1604"},
							},
							{
								Src: SourceRef{"virtualbox-iso", "ubuntu-1204"},
							},
						},
						ProvisionerGroups: ProvisionerGroups{
							&ProvisionerGroup{
								CommunicatorRef: CommunicatorRef{"ssh", "vagrant"},
								Provisioners: []Provisioner{
									{Cfg: &shell.FlatConfig{
										Inline: []string{"echo '{{user `my_secret`}}' :D"},
									}},
									{Cfg: &shell.FlatConfig{
										Scripts:        []string{"script-1.sh", "script-2.sh"},
										ValidExitCodes: []int{0, 42},
									}},
									{Cfg: &file.FlatConfig{
										Source:      strPtr("app.tar.gz"),
										Destination: strPtr("/tmp/app.tar.gz"),
									}},
								},
							},
						},
						PostProvisionerGroups: ProvisionerGroups{
							&ProvisionerGroup{
								Provisioners: []Provisioner{
									{Cfg: &amazon_import.FlatConfig{
										Name: strPtr("that-ubuntu-1.0"),
									}},
								},
							},
						},
					},
					&Build{
						Froms: BuildFromList{
							{
								Src: SourceRef{"amazon", "that-ubuntu-1"},
							},
						},
						ProvisionerGroups: ProvisionerGroups{
							&ProvisionerGroup{
								Provisioners: []Provisioner{
									{Cfg: &shell.FlatConfig{
										Inline: []string{"echo HOLY GUACAMOLE !"},
									}},
								},
							},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.parser
			f, moreDiags := p.ParseHCLFile(tt.args.filename)
			if moreDiags != nil {
				t.Fatalf("diags: %s", moreDiags)
			}
			diags := p.ParseFile(f, tt.args.cfg)
			if tt.wantDiags == (diags == nil) {
				for _, diag := range diags {
					t.Errorf("PackerConfig.Load() unexpected diagnostics. %v", diag)
				}
				t.Error("")
			}
			if diff := cmp.Diff(tt.wantPackerConfig, tt.args.cfg,
				cmpopts.IgnoreUnexported(cty.Value{}),
				cmpopts.IgnoreTypes(HCL2Ref{}),
				cmpopts.IgnoreTypes([]hcl.Range{}),
				cmpopts.IgnoreTypes(hcl.Range{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Expression }{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Body }{}),
			); diff != "" {
				t.Errorf("PackerConfig.Load() wrong packer config. %s", diff)
			}
			if t.Failed() {
				t.Fatal()
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }
