// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	// Previously core-bundled components, split into their own plugins but
	// still vendored with Packer for now. Importing as library instead of
	// forcing use of packer init, until packer v1.8.0

	alicloudecsbuilder "github.com/hashicorp/packer-plugin-alicloud/builder/ecs"
	alicloudimportpostprocessor "github.com/hashicorp/packer-plugin-alicloud/post-processor/alicloud-import"
	amazonchrootbuilder "github.com/hashicorp/packer-plugin-amazon/builder/chroot"
	amazonebsbuilder "github.com/hashicorp/packer-plugin-amazon/builder/ebs"
	amazonebssurrogatebuilder "github.com/hashicorp/packer-plugin-amazon/builder/ebssurrogate"
	amazonebsvolumebuilder "github.com/hashicorp/packer-plugin-amazon/builder/ebsvolume"
	amazoninstancebuilder "github.com/hashicorp/packer-plugin-amazon/builder/instance"
	amazonamidatasource "github.com/hashicorp/packer-plugin-amazon/datasource/ami"
	amazonsecretsmanagerdatasource "github.com/hashicorp/packer-plugin-amazon/datasource/secretsmanager"
	anazibimportpostprocessor "github.com/hashicorp/packer-plugin-amazon/post-processor/import"
	ansibleprovisioner "github.com/hashicorp/packer-plugin-ansible/provisioner/ansible"
	ansiblelocalprovisioner "github.com/hashicorp/packer-plugin-ansible/provisioner/ansible-local"
	azurearmbuilder "github.com/hashicorp/packer-plugin-azure/builder/azure/arm"
	azurechrootbuilder "github.com/hashicorp/packer-plugin-azure/builder/azure/chroot"
	azuredtlbuilder "github.com/hashicorp/packer-plugin-azure/builder/azure/dtl"
	azuredtlartifactprovisioner "github.com/hashicorp/packer-plugin-azure/provisioner/azure-dtlartifact"
	chefclientprovisioner "github.com/hashicorp/packer-plugin-chef/provisioner/chef-client"
	chefsoloprovisioner "github.com/hashicorp/packer-plugin-chef/provisioner/chef-solo"
	cloudstackbuilder "github.com/hashicorp/packer-plugin-cloudstack/builder/cloudstack"
	convergeprovisioner "github.com/hashicorp/packer-plugin-converge/provisioner/converge"
	dockerbuilder "github.com/hashicorp/packer-plugin-docker/builder/docker"
	dockerimportpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-import"
	dockerpushpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-push"
	dockersavepostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-save"
	dockertagpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-tag"
	googlecomputebuilder "github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	googlecomputeexportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-export"
	googlecomputeimportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-import"
	hcloudbuilder "github.com/hashicorp/packer-plugin-hcloud/builder/hcloud"
	hyperonebuilder "github.com/hashicorp/packer-plugin-hyperone/builder/hyperone"
	hypervisobuilder "github.com/hashicorp/packer-plugin-hyperv/builder/hyperv/iso"
	hypervvmcxbuilder "github.com/hashicorp/packer-plugin-hyperv/builder/hyperv/vmcx"
	inspecprovisioner "github.com/hashicorp/packer-plugin-inspec/provisioner/inspec"
	jdcloudbuilder "github.com/hashicorp/packer-plugin-jdcloud/builder/jdcloud"
	lxcbuilder "github.com/hashicorp/packer-plugin-lxc/builder/lxc"
	lxdbuilder "github.com/hashicorp/packer-plugin-lxd/builder/lxd"
	ncloudbuilder "github.com/hashicorp/packer-plugin-ncloud/builder/ncloud"
	oneandonebuilder "github.com/hashicorp/packer-plugin-oneandone/builder/oneandone"
	openstackbuilder "github.com/hashicorp/packer-plugin-openstack/builder/openstack"
	parallelsisobuilder "github.com/hashicorp/packer-plugin-parallels/builder/parallels/iso"
	parallelspvmbuilder "github.com/hashicorp/packer-plugin-parallels/builder/parallels/pvm"
	profitbricksbuilder "github.com/hashicorp/packer-plugin-profitbricks/builder/profitbricks"
	proxmoxclone "github.com/hashicorp/packer-plugin-proxmox/builder/proxmox/clone"
	proxmoxiso "github.com/hashicorp/packer-plugin-proxmox/builder/proxmox/iso"
	puppetmasterlessprovisioner "github.com/hashicorp/packer-plugin-puppet/provisioner/puppet-masterless"
	puppetserverprovisioner "github.com/hashicorp/packer-plugin-puppet/provisioner/puppet-server"
	qemubuilder "github.com/hashicorp/packer-plugin-qemu/builder/qemu"
	saltmasterlessprovisioner "github.com/hashicorp/packer-plugin-salt/provisioner/salt-masterless"
	tencentcloudcvmbuilder "github.com/hashicorp/packer-plugin-tencentcloud/builder/tencentcloud/cvm"
	tritonbuilder "github.com/hashicorp/packer-plugin-triton/builder/triton"
	vagrantbuilder "github.com/hashicorp/packer-plugin-vagrant/builder/vagrant"
	vagrantpostprocessor "github.com/hashicorp/packer-plugin-vagrant/post-processor/vagrant"
	vagrantcloudpostprocessor "github.com/hashicorp/packer-plugin-vagrant/post-processor/vagrant-cloud"
	virtualboxisobuilder "github.com/hashicorp/packer-plugin-virtualbox/builder/virtualbox/iso"
	virtualboxovfbuilder "github.com/hashicorp/packer-plugin-virtualbox/builder/virtualbox/ovf"
	virtualboxvmbuilder "github.com/hashicorp/packer-plugin-virtualbox/builder/virtualbox/vm"
	vmwareisobuilder "github.com/hashicorp/packer-plugin-vmware/builder/vmware/iso"
	vmwarevmxbuilder "github.com/hashicorp/packer-plugin-vmware/builder/vmware/vmx"
	vsphereclonebuilder "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/clone"
	vsphereisobuilder "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/iso"
	vspherepostprocessor "github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere"
	vspheretemplatepostprocessor "github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere-template"
	yandexbuilder "github.com/hashicorp/packer-plugin-yandex/builder/yandex"
	yandexexportpostprocessor "github.com/hashicorp/packer-plugin-yandex/post-processor/yandex-export"
	yandeximportpostprocessor "github.com/hashicorp/packer-plugin-yandex/post-processor/yandex-import"
)

// VendoredDatasources are datasource components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredDatasources = map[string]packersdk.Datasource{
	"amazon-ami":            new(amazonamidatasource.Datasource),
	"amazon-secretsmanager": new(amazonsecretsmanagerdatasource.Datasource),
}

// VendoredBuilders are builder components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredBuilders = map[string]packersdk.Builder{
	"alicloud-ecs":        new(alicloudecsbuilder.Builder),
	"amazon-ebs":          new(amazonebsbuilder.Builder),
	"amazon-chroot":       new(amazonchrootbuilder.Builder),
	"amazon-ebssurrogate": new(amazonebssurrogatebuilder.Builder),
	"amazon-ebsvolume":    new(amazonebsvolumebuilder.Builder),
	"amazon-instance":     new(amazoninstancebuilder.Builder),
	"azure-arm":           new(azurearmbuilder.Builder),
	"azure-chroot":        new(azurechrootbuilder.Builder),
	"azure-dtl":           new(azuredtlbuilder.Builder),
	"cloudstack":          new(cloudstackbuilder.Builder),
	"docker":              new(dockerbuilder.Builder),
	"googlecompute":       new(googlecomputebuilder.Builder),
	"hcloud":              new(hcloudbuilder.Builder),
	"hyperv-iso":          new(hypervisobuilder.Builder),
	"hyperv-vmcx":         new(hypervvmcxbuilder.Builder),
	"hyperone":            new(hyperonebuilder.Builder),
	"jdcloud":             new(jdcloudbuilder.Builder),
	"lxc":                 new(lxcbuilder.Builder),
	"lxd":                 new(lxdbuilder.Builder),
	"ncloud":              new(ncloudbuilder.Builder),
	"oneandone":           new(oneandonebuilder.Builder),
	"openstack":           new(openstackbuilder.Builder),
	"profitbricks":        new(profitbricksbuilder.Builder),
	"proxmox":             new(proxmoxiso.Builder),
	"proxmox-iso":         new(proxmoxiso.Builder),
	"proxmox-clone":       new(proxmoxclone.Builder),
	"parallels-iso":       new(parallelsisobuilder.Builder),
	"parallels-pvm":       new(parallelspvmbuilder.Builder),
	"qemu":                new(qemubuilder.Builder),
	"tencentcloud-cvm":    new(tencentcloudcvmbuilder.Builder),
	"triton":              new(tritonbuilder.Builder),
	"vagrant":             new(vagrantbuilder.Builder),
	"vsphere-clone":       new(vsphereclonebuilder.Builder),
	"vsphere-iso":         new(vsphereisobuilder.Builder),
	"virtualbox-iso":      new(virtualboxisobuilder.Builder),
	"virtualbox-ovf":      new(virtualboxovfbuilder.Builder),
	"virtualbox-vm":       new(virtualboxvmbuilder.Builder),
	"vmware-iso":          new(vmwareisobuilder.Builder),
	"vmware-vmx":          new(vmwarevmxbuilder.Builder),
	"yandex":              new(yandexbuilder.Builder),
}

// VendoredProvisioners are provisioner components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredProvisioners = map[string]packersdk.Provisioner{
	"azure-dtlartifact": new(azuredtlartifactprovisioner.Provisioner),
	"ansible":           new(ansibleprovisioner.Provisioner),
	"ansible-local":     new(ansiblelocalprovisioner.Provisioner),
	"chef-client":       new(chefclientprovisioner.Provisioner),
	"chef-solo":         new(chefsoloprovisioner.Provisioner),
	"converge":          new(convergeprovisioner.Provisioner),
	"inspec":            new(inspecprovisioner.Provisioner),
	"puppet-masterless": new(puppetmasterlessprovisioner.Provisioner),
	"puppet-server":     new(puppetserverprovisioner.Provisioner),
	"salt-masterless":   new(saltmasterlessprovisioner.Provisioner),
}

// VendoredPostProcessors are post-processor components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredPostProcessors = map[string]packersdk.PostProcessor{
	"alicloud-import":      new(alicloudimportpostprocessor.PostProcessor),
	"amazon-import":        new(anazibimportpostprocessor.PostProcessor),
	"docker-import":        new(dockerimportpostprocessor.PostProcessor),
	"docker-push":          new(dockerpushpostprocessor.PostProcessor),
	"docker-save":          new(dockersavepostprocessor.PostProcessor),
	"docker-tag":           new(dockertagpostprocessor.PostProcessor),
	"googlecompute-export": new(googlecomputeexportpostprocessor.PostProcessor),
	"googlecompute-import": new(googlecomputeimportpostprocessor.PostProcessor),
	"vagrant":              new(vagrantpostprocessor.PostProcessor),
	"vagrant-cloud":        new(vagrantcloudpostprocessor.PostProcessor),
	"vsphere-template":     new(vspheretemplatepostprocessor.PostProcessor),
	"vsphere":              new(vspherepostprocessor.PostProcessor),
	"yandex-export":        new(yandexexportpostprocessor.PostProcessor),
	"yandex-import":        new(yandeximportpostprocessor.PostProcessor),
}

// Upon init lets load up any plugins that were vendored manually into the default
// set of plugins.
func init() {
	for k, v := range VendoredDatasources {
		if _, ok := Datasources[k]; ok {
			continue
		}
		Datasources[k] = v
	}

	for k, v := range VendoredBuilders {
		if _, ok := Builders[k]; ok {
			continue
		}
		Builders[k] = v
	}

	for k, v := range VendoredProvisioners {
		if _, ok := Provisioners[k]; ok {
			continue
		}
		Provisioners[k] = v
	}

	for k, v := range VendoredPostProcessors {
		if _, ok := PostProcessors[k]; ok {
			continue
		}
		PostProcessors[k] = v
	}
}
