module github.com/jetbrains-infra/packer-builder-vsphere

require (
	github.com/hashicorp/packer v1.4.2
	github.com/vmware/govmomi v0.20.0
	go.uber.org/goleak v0.10.1-0.20190517053103-3b0196519f16
	golang.org/x/mobile v0.0.0-20190607214518-6fa95d984e88
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

