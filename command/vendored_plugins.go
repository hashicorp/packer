package command

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	// Previously core-bundled components, split into their own plugins but
	// still vendored with Packer for now. Importing as library instead of
	// forcing use of packer init, until packer v1.8.0

	alicloudecsbuilder "github.com/hashicorp/packer-plugin-alicloud/builder/ecs"
	alicloudimportpostprocessor "github.com/hashicorp/packer-plugin-alicloud/post-processor/alicloud-import"
	ansibleprovisioner "github.com/hashicorp/packer-plugin-ansible/provisioner/ansible"
	ansiblelocalprovisioner "github.com/hashicorp/packer-plugin-ansible/provisioner/ansible-local"
	chefclientprovisioner "github.com/hashicorp/packer-plugin-chef/provisioner/chef-client"
	chefsoloprovisioner "github.com/hashicorp/packer-plugin-chef/provisioner/chef-solo"
	cloudstackbuilder "github.com/hashicorp/packer-plugin-cloudstack/builder/cloudstack"
	convergeprovisioner "github.com/hashicorp/packer-plugin-converge/provisioner/converge"
	digitaloceanbuilder "github.com/hashicorp/packer-plugin-digitalocean/builder/digitalocean"
	digitaloceanimportpostprocessor "github.com/hashicorp/packer-plugin-digitalocean/post-processor/digitalocean-import"
	googlecomputebuilder "github.com/hashicorp/packer-plugin-googlecompute/builder/googlecompute"
	googlecomputeexportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-export"
	googlecomputeimportpostprocessor "github.com/hashicorp/packer-plugin-googlecompute/post-processor/googlecompute-import"
	hcloudbuilder "github.com/hashicorp/packer-plugin-hcloud/builder/hcloud"
	hyperonebuilder "github.com/hashicorp/packer-plugin-hyperone/builder/hyperone"
	hypervisobuilder "github.com/hashicorp/packer-plugin-hyperv/builder/hyperv/iso"
	hypervvmcxbuilder "github.com/hashicorp/packer-plugin-hyperv/builder/hyperv/vmcx"
	inspecprovisioner "github.com/hashicorp/packer-plugin-inspec/provisioner/inspec"
	jdcloudbuilder "github.com/hashicorp/packer-plugin-jdcloud/builder/jdcloud"
	linodebuilder "github.com/hashicorp/packer-plugin-linode/builder/linode"
	lxcbuilder "github.com/hashicorp/packer-plugin-lxc/builder/lxc"
	lxdbuilder "github.com/hashicorp/packer-plugin-lxd/builder/lxd"
	ncloudbuilder "github.com/hashicorp/packer-plugin-ncloud/builder/ncloud"
	oneandonebuilder "github.com/hashicorp/packer-plugin-oneandone/builder/oneandone"
	openstackbuilder "github.com/hashicorp/packer-plugin-openstack/builder/openstack"
	oracleclassicbuilder "github.com/hashicorp/packer-plugin-oracle/builder/classic"
	oracleocibuilder "github.com/hashicorp/packer-plugin-oracle/builder/oci"
	oscbsubuilder "github.com/hashicorp/packer-plugin-outscale/builder/osc/bsu"
	oscbsusurrogatebuilder "github.com/hashicorp/packer-plugin-outscale/builder/osc/bsusurrogate"
	oscbsuvolumebuilder "github.com/hashicorp/packer-plugin-outscale/builder/osc/bsuvolume"
	oscchrootbuilder "github.com/hashicorp/packer-plugin-outscale/builder/osc/chroot"
	profitbricksbuilder "github.com/hashicorp/packer-plugin-profitbricks/builder/profitbricks"
	proxmoxclone "github.com/hashicorp/packer-plugin-proxmox/builder/proxmox/clone"
	proxmoxiso "github.com/hashicorp/packer-plugin-proxmox/builder/proxmox/iso"
	puppetmasterlessprovisioner "github.com/hashicorp/packer-plugin-puppet/provisioner/puppet-masterless"
	puppetserverprovisioner "github.com/hashicorp/packer-plugin-puppet/provisioner/puppet-server"
	qemubuilder "github.com/hashicorp/packer-plugin-qemu/builder/qemu"
	saltmasterlessprovisioner "github.com/hashicorp/packer-plugin-salt/provisioner/salt-masterless"
	tencentcloudcvmbuilder "github.com/hashicorp/packer-plugin-tencentcloud/builder/tencentcloud/cvm"
	tritonbuilder "github.com/hashicorp/packer-plugin-triton/builder/triton"
	uclouduhostbuilder "github.com/hashicorp/packer-plugin-ucloud/builder/ucloud/uhost"
	ucloudimportpostprocessor "github.com/hashicorp/packer-plugin-ucloud/post-processor/ucloud-import"
	vagrantbuilder "github.com/hashicorp/packer-plugin-vagrant/builder/vagrant"
	vagrantpostprocessor "github.com/hashicorp/packer-plugin-vagrant/post-processor/vagrant"
	vagrantcloudpostprocessor "github.com/hashicorp/packer-plugin-vagrant/post-processor/vagrant-cloud"
	yandexbuilder "github.com/hashicorp/packer-plugin-yandex/builder/yandex"
	yandexexportpostprocessor "github.com/hashicorp/packer-plugin-yandex/post-processor/yandex-export"
	yandeximportpostprocessor "github.com/hashicorp/packer-plugin-yandex/post-processor/yandex-import"
	scalewaybuilder "github.com/scaleway/packer-plugin-scaleway/builder/scaleway"
)

// VendoredDatasources are datasource components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredDatasources = map[string]packersdk.Datasource{}

// VendoredBuilders are builder components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredBuilders = map[string]packersdk.Builder{
	"alicloud-ecs":   new(alicloudecsbuilder.Builder),
	"cloudstack":     new(cloudstackbuilder.Builder),
	"digitalocean":   new(digitaloceanbuilder.Builder),
	"googlecompute":  new(googlecomputebuilder.Builder),
	"hcloud":         new(hcloudbuilder.Builder),
	"hyperv-iso":     new(hypervisobuilder.Builder),
	"hyperv-vmcx":    new(hypervvmcxbuilder.Builder),
	"hyperone":       new(hyperonebuilder.Builder),
	"jdcloud":        new(jdcloudbuilder.Builder),
	"linode":         new(linodebuilder.Builder),
	"lxc":            new(lxcbuilder.Builder),
	"lxd":            new(lxdbuilder.Builder),
	"ncloud":         new(ncloudbuilder.Builder),
	"oneandone":      new(oneandonebuilder.Builder),
	"openstack":      new(openstackbuilder.Builder),
	"oracle-classic": new(oracleclassicbuilder.Builder),
	"oracle-oci":     new(oracleocibuilder.Builder),
	"profitbricks":   new(profitbricksbuilder.Builder),
	"proxmox":        new(proxmoxiso.Builder),
	"proxmox-iso":    new(proxmoxiso.Builder),
	"proxmox-clone":  new(proxmoxclone.Builder),

	"qemu":             new(qemubuilder.Builder),
	"scaleway":         new(scalewaybuilder.Builder),
	"tencentcloud-cvm": new(tencentcloudcvmbuilder.Builder),
	"triton":           new(tritonbuilder.Builder),
	"ucloud-uhost":     new(uclouduhostbuilder.Builder),
	"vagrant":          new(vagrantbuilder.Builder),
	"osc-bsu":          new(oscbsubuilder.Builder),
	"osc-bsusurrogate": new(oscbsusurrogatebuilder.Builder),
	"osc-bsuvolume":    new(oscbsuvolumebuilder.Builder),
	"osc-chroot":       new(oscchrootbuilder.Builder),
	"yandex":           new(yandexbuilder.Builder),
}

// VendoredProvisioners are provisioner components that were once bundled with the
// Packer core, but are now being imported from their counterpart plugin repos
var VendoredProvisioners = map[string]packersdk.Provisioner{
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
	"digitalocean-import":  new(digitaloceanimportpostprocessor.PostProcessor),
	"googlecompute-export": new(googlecomputeexportpostprocessor.PostProcessor),
	"googlecompute-import": new(googlecomputeimportpostprocessor.PostProcessor),
	"ucloud-import":        new(ucloudimportpostprocessor.PostProcessor),
	"vagrant":              new(vagrantpostprocessor.PostProcessor),
	"vagrant-cloud":        new(vagrantcloudpostprocessor.PostProcessor),
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
