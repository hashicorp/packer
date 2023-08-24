// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"fmt"
	"log"
	"strings"

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

// bundledStatus is used to know if one of the bundled components is loaded from
// an external plugin, or from the bundled plugins.
//
// We keep track of this to produce a warning if a user relies on one
// such plugin, as they will be removed in a later version of Packer.
var bundledStatus = map[string]bool{
	"packer-builder-amazon-ebs":               false,
	"packer-builder-amazon-chroot":            false,
	"packer-builder-amazon-ebssurrogate":      false,
	"packer-builder-amazon-ebsvolume":         false,
	"packer-builder-amazon-instance":          false,
	"packer-post-processor-amazon-import":     false,
	"packer-datasource-amazon-ami":            false,
	"packer-datasource-amazon-secretsmanager": false,

	"packer-provisioner-ansible":       false,
	"packer-provisioner-ansible-local": false,

	"packer-provisioner-azure-dtlartifact": false,
	"packer-builder-azure-arm":             false,
	"packer-builder-azure-chroot":          false,
	"packer-builder-azure-dtl":             false,

	"packer-builder-docker":               false,
	"packer-post-processor-docker-import": false,
	"packer-post-processor-docker-push":   false,
	"packer-post-processor-docker-save":   false,
	"packer-post-processor-docker-tag":    false,

	"packer-builder-googlecompute":               false,
	"packer-post-processor-googlecompute-export": false,
	"packer-post-processor-googlecompute-import": false,

	"packer-builder-qemu": false,

	"packer-builder-vagrant":              false,
	"packer-post-processor-vagrant":       false,
	"packer-post-processor-vagrant-cloud": false,

	"packer-builder-virtualbox-iso": false,
	"packer-builder-virtualbox-ovf": false,
	"packer-builder-virtualbox-vm":  false,

	"packer-builder-vmware-iso": false,
	"packer-builder-vmware-vmx": false,

	"packer-builder-vsphere-clone":           false,
	"packer-builder-vsphere-iso":             false,
	"packer-post-processor-vsphere-template": false,
	"packer-post-processor-vsphere":          false,
}

// TrackBundledPlugin marks a component as loaded from Packer's bundled plugins
// instead of from an externally loaded plugin.
//
// NOTE: `pluginName' must be in the format `packer-<type>-<component_name>'
func TrackBundledPlugin(pluginName string) {
	_, exists := bundledStatus[pluginName]
	if !exists {
		return
	}

	bundledStatus[pluginName] = true
}

var componentPluginMap = map[string]string{
	"packer-builder-amazon-ebs":               "github.com/hashicorp/amazon",
	"packer-builder-amazon-chroot":            "github.com/hashicorp/amazon",
	"packer-builder-amazon-ebssurrogate":      "github.com/hashicorp/amazon",
	"packer-builder-amazon-ebsvolume":         "github.com/hashicorp/amazon",
	"packer-builder-amazon-instance":          "github.com/hashicorp/amazon",
	"packer-post-processor-amazon-import":     "github.com/hashicorp/amazon",
	"packer-datasource-amazon-ami":            "github.com/hashicorp/amazon",
	"packer-datasource-amazon-secretsmanager": "github.com/hashicorp/amazon",

	"packer-provisioner-ansible":       "github.com/hashicorp/ansible",
	"packer-provisioner-ansible-local": "github.com/hashicorp/ansible",

	"packer-provisioner-azure-dtlartifact": "github.com/hashicorp/azure",
	"packer-builder-azure-arm":             "github.com/hashicorp/azure",
	"packer-builder-azure-chroot":          "github.com/hashicorp/azure",
	"packer-builder-azure-dtl":             "github.com/hashicorp/azure",

	"packer-builder-docker":               "github.com/hashicorp/docker",
	"packer-post-processor-docker-import": "github.com/hashicorp/docker",
	"packer-post-processor-docker-push":   "github.com/hashicorp/docker",
	"packer-post-processor-docker-save":   "github.com/hashicorp/docker",
	"packer-post-processor-docker-tag":    "github.com/hashicorp/docker",

	"packer-builder-googlecompute":               "github.com/hashicorp/googlecompute",
	"packer-post-processor-googlecompute-export": "github.com/hashicorp/googlecompute",
	"packer-post-processor-googlecompute-import": "github.com/hashicorp/googlecompute",

	"packer-builder-qemu": "github.com/hashicorp/qemu",

	"packer-builder-vagrant":              "github.com/hashicorp/vagrant",
	"packer-post-processor-vagrant":       "github.com/hashicorp/vagrant",
	"packer-post-processor-vagrant-cloud": "github.com/hashicorp/vagrant",

	"packer-builder-virtualbox-iso": "github.com/hashicorp/virtualbox",
	"packer-builder-virtualbox-ovf": "github.com/hashicorp/virtualbox",
	"packer-builder-virtualbox-vm":  "github.com/hashicorp/virtualbox",

	"packer-builder-vmware-iso": "github.com/hashicorp/vmware",
	"packer-builder-vmware-vmx": "github.com/hashicorp/vmware",

	"packer-builder-vsphere-clone":           "github.com/hashicorp/vsphere",
	"packer-builder-vsphere-iso":             "github.com/hashicorp/vsphere",
	"packer-post-processor-vsphere-template": "github.com/hashicorp/vsphere",
	"packer-post-processor-vsphere":          "github.com/hashicorp/vsphere",
}

// compileBundledPluginList returns a list of plugins to import in a config
//
// This only works on bundled plugins and serves as a way to inform users that
// they should not rely on a bundled plugin anymore, but give them recommendations
// on how to manage those plugins instead.
func compileBundledPluginList(componentMap map[string]struct{}) []string {
	plugins := map[string]struct{}{}
	for component := range componentMap {
		plugin, ok := componentPluginMap[component]
		if !ok {
			log.Printf("Unknown bundled plugin component: %q", component)
			continue
		}

		plugins[plugin] = struct{}{}
	}

	pluginList := make([]string, 0, len(plugins))
	for plugin := range plugins {
		pluginList = append(pluginList, plugin)
	}

	return pluginList
}

func generateRequiredPluginsBlock(plugins []string) string {
	if len(plugins) == 0 {
		return ""
	}

	buf := &strings.Builder{}
	buf.WriteString(`
packer {
  required_plugins {`)

	for _, plugin := range plugins {
		pluginName := strings.Replace(plugin, "github.com/hashicorp/", "", 1)
		fmt.Fprintf(buf, `
    %s = {
      source  = %q
      version = "~> 1"
    }`, pluginName, plugin)
	}

	buf.WriteString(`
  }
}
`)

	return buf.String()
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
