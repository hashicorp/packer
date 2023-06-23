// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	// Previously core-bundled components, split into their own plugins but
	// still vendored with Packer for now. Importing as library instead of
	// forcing use of packer init.

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
	dockerbuilder "github.com/hashicorp/packer-plugin-docker/builder/docker"
	dockerimportpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-import"
	dockerpushpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-push"
	dockersavepostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-save"
	dockertagpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-tag"
	googlecomputebuilder "github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	googlecomputeexportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-export"
	googlecomputeimportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-import"
	qemubuilder "github.com/hashicorp/packer-plugin-qemu/builder/qemu"
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
	"amazon-ebs":          new(amazonebsbuilder.Builder),
	"amazon-chroot":       new(amazonchrootbuilder.Builder),
	"amazon-ebssurrogate": new(amazonebssurrogatebuilder.Builder),
	"amazon-ebsvolume":    new(amazonebsvolumebuilder.Builder),
	"amazon-instance":     new(amazoninstancebuilder.Builder),
	"azure-arm":           new(azurearmbuilder.Builder),
	"azure-chroot":        new(azurechrootbuilder.Builder),
	"azure-dtl":           new(azuredtlbuilder.Builder),
	"docker":              new(dockerbuilder.Builder),
	"googlecompute":       new(googlecomputebuilder.Builder),
	"qemu":                new(qemubuilder.Builder),
	"vagrant":             new(vagrantbuilder.Builder),
	"vsphere-clone":       new(vsphereclonebuilder.Builder),
	"vsphere-iso":         new(vsphereisobuilder.Builder),
	"virtualbox-iso":      new(virtualboxisobuilder.Builder),
	"virtualbox-ovf":      new(virtualboxovfbuilder.Builder),
	"virtualbox-vm":       new(virtualboxvmbuilder.Builder),
	"vmware-iso":          new(vmwareisobuilder.Builder),
	"vmware-vmx":          new(vmwarevmxbuilder.Builder),
}

// VendoredProvisioners are provisioner components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredProvisioners = map[string]packersdk.Provisioner{
	"azure-dtlartifact": new(azuredtlartifactprovisioner.Provisioner),
	"ansible":           new(ansibleprovisioner.Provisioner),
	"ansible-local":     new(ansiblelocalprovisioner.Provisioner),
}

// VendoredPostProcessors are post-processor components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredPostProcessors = map[string]packersdk.PostProcessor{
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
