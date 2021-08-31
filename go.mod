module github.com/hashicorp/packer

require (
	cloud.google.com/go/storage v1.16.0 // indirect
	github.com/Azure/azure-sdk-for-go v57.0.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.20 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.15 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.8 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.3 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20200615164410-66371956d46c // indirect
	github.com/Telmate/proxmox-api-go v0.0.0-20210825163308-5e4c0d698a78 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1244 // indirect
	github.com/aliyun/aliyun-oss-go-sdk v2.1.10+incompatible // indirect
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/aws/aws-sdk-go v1.40.33 // indirect
	github.com/biogo/hts v1.4.3
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cheggaaa/pb v1.0.27
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/digitalocean/go-libvirt v0.0.0-20210723161134-761cfeeb5968 // indirect
	github.com/dsnet/compress v0.0.1
	github.com/exoscale/packer-plugin-exoscale v0.1.1
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-openapi/runtime v0.19.24
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v33 v33.0.1-0.20210113204525-9318e629ec69
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026
	github.com/hashicorp/consul/api v1.10.0 // indirect
	github.com/hashicorp/go-checkpoint v0.0.0-20171009173528-1545e56e46de
	github.com/hashicorp/go-cty-funcs v0.0.0-20200930094925-2721b1e36840
	github.com/hashicorp/go-getter/v2 v2.0.0
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl/v2 v2.10.1
	github.com/hashicorp/hcp-sdk-go v0.10.1-0.20210727200019-239ce8d80646
	github.com/hashicorp/packer-plugin-alicloud v1.0.0
	github.com/hashicorp/packer-plugin-amazon v1.0.1-dev
	github.com/hashicorp/packer-plugin-ansible v1.0.0
	github.com/hashicorp/packer-plugin-azure v1.0.2
	github.com/hashicorp/packer-plugin-chef v1.0.1
	github.com/hashicorp/packer-plugin-cloudstack v1.0.0
	github.com/hashicorp/packer-plugin-converge v1.0.0
	github.com/hashicorp/packer-plugin-digitalocean v1.0.1
	github.com/hashicorp/packer-plugin-docker v1.0.1
	github.com/hashicorp/packer-plugin-googlecompute v1.0.2
	github.com/hashicorp/packer-plugin-hcloud v1.0.0
	github.com/hashicorp/packer-plugin-hyperone v1.0.0
	github.com/hashicorp/packer-plugin-hyperv v1.0.0
	github.com/hashicorp/packer-plugin-inspec v0.0.1
	github.com/hashicorp/packer-plugin-jdcloud v1.0.0
	github.com/hashicorp/packer-plugin-linode v1.0.0
	github.com/hashicorp/packer-plugin-lxc v1.0.0
	github.com/hashicorp/packer-plugin-lxd v1.0.0
	github.com/hashicorp/packer-plugin-ncloud v1.0.1
	github.com/hashicorp/packer-plugin-oneandone v1.0.0
	github.com/hashicorp/packer-plugin-openstack v1.0.0
	github.com/hashicorp/packer-plugin-oracle v1.0.1
	github.com/hashicorp/packer-plugin-outscale v1.0.1
	github.com/hashicorp/packer-plugin-parallels v1.0.0
	github.com/hashicorp/packer-plugin-profitbricks v1.0.1
	github.com/hashicorp/packer-plugin-proxmox v1.0.1
	github.com/hashicorp/packer-plugin-puppet v1.0.0
	github.com/hashicorp/packer-plugin-qemu v1.0.0
	github.com/hashicorp/packer-plugin-salt v0.0.2
	github.com/hashicorp/packer-plugin-scaleway v1.0.1
	github.com/hashicorp/packer-plugin-sdk v0.2.4
	github.com/hashicorp/packer-plugin-tencentcloud v1.0.1
	github.com/hashicorp/packer-plugin-triton v1.0.0
	github.com/hashicorp/packer-plugin-ucloud v1.0.0
	github.com/hashicorp/packer-plugin-vagrant v1.0.0
	github.com/hashicorp/packer-plugin-virtualbox v1.0.0
	github.com/hashicorp/packer-plugin-vmware v1.0.1
	github.com/hashicorp/packer-plugin-vsphere v1.0.1
	github.com/hashicorp/packer-plugin-yandex v1.0.0
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/hetznercloud/hcloud-go v1.32.0 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869
	github.com/klauspost/compress v1.13.5 // indirect
	github.com/klauspost/pgzip v1.2.5
	github.com/masterzen/winrm v0.0.0-20210623064412-3b76017826b0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-tty v0.0.0-20191112051231-74040eebce08
	github.com/mitchellh/cli v1.1.0
	github.com/mitchellh/go-fs v0.0.0-20180402235330-b7b9ca407fff // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mitchellh/panicwrap v1.0.0
	github.com/mitchellh/prefixedio v0.0.0-20151214002211-6e6954073784
	github.com/packer-community/winrmcp v0.0.0-20180921211025-c76d91c1e7db // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/pkg/sftp v1.13.2 // indirect
	github.com/posener/complete v1.2.3
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/shirou/gopsutil v3.21.1+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/tencentcloud/tencentcloud-sdk-go v3.0.233+incompatible // indirect
	github.com/ugorji/go v1.2.6 // indirect
	github.com/ulikunitz/xz v0.5.10
	github.com/vmware/govmomi v0.26.1 // indirect
	github.com/yandex-cloud/go-sdk v0.0.0-20210824141121-182aedd44a25 // indirect
	github.com/zclconf/go-cty v1.9.1
	github.com/zclconf/go-cty-yaml v1.0.1
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/mobile v0.0.0-20210716004757-34ab1303b554 // indirect
	golang.org/x/mod v0.5.0
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.5
	google.golang.org/api v0.55.0 // indirect
	google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2 // indirect
	google.golang.org/grpc v1.40.0
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

go 1.16
