package hcl2template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	amazonebs "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/builder/virtualbox/iso"

	"github.com/hashicorp/packer/helper/communicator"

	amazon_import "github.com/hashicorp/packer/post-processor/amazon-import"

	"github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
)

func TestParser_Parse(t *testing.T) {
	defaultParser := getBasicParser()

	type args struct {
		filename string
	}
	tests := []struct {
		name      string
		parser    *Parser
		args      args
		wantCfg   *PackerConfig
		wantDiags bool
	}{
		{"complete",
			defaultParser,
			args{"testdata/complete"},
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
				Variables: PackerV1Variables{
					"image_name": "foo-image-{{user `my_secret`}}",
					"key":        "value",
					"my_secret":  "foo",
				},
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
			}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCfg, gotDiags := tt.parser.Parse(tt.args.filename)
			if tt.wantDiags == (gotDiags == nil) {
				t.Errorf("Parser.Parse() unexpected diagnostics. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.wantCfg, gotCfg,
				cmpopts.IgnoreUnexported(cty.Value{}),
				cmpopts.IgnoreTypes(HCL2Ref{}),
				cmpopts.IgnoreTypes([]hcl.Range{}),
				cmpopts.IgnoreTypes(hcl.Range{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Expression }{}),
				cmpopts.IgnoreInterfaces(struct{ hcl.Body }{}),
			); diff != "" {
				t.Errorf("Parser.Parse() wrong packer config. %s", diff)
			}

		})
	}
}
