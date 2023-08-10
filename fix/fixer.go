// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

// A Fixer is something that can perform a fix operation on a template.
type Fixer interface {
	// DeprecatedOptions returns the name(s) of the option(s) being replaced in
	// this fixer. It is used to generate a list of deprecated options that the
	// template parser checks against to warn users that they need to call
	// `packer fix` against their templates after upgrading.
	DeprecatedOptions() map[string][]string

	// Fix takes a raw map structure input, potentially transforms it
	// in some way, and returns the new, transformed structure. The
	// Fix method is allowed to mutate the input.
	Fix(input map[string]interface{}) (map[string]interface{}, error)

	// Synopsis returns a string description of what the fixer actually
	// does.
	Synopsis() string
}

// Fixers is the map of all available fixers, by name.
var Fixers map[string]Fixer

// FixerOrder is the default order the fixers should be run.
var FixerOrder []string

func init() {
	Fixers = map[string]Fixer{
		"iso-md5":                    new(FixerISOMD5),
		"createtime":                 new(FixerCreateTime),
		"pp-vagrant-override":        new(FixerVagrantPPOverride),
		"virtualbox-gaattach":        new(FixerVirtualBoxGAAttach),
		"virtualbox-rename":          new(FixerVirtualBoxRename),
		"vmware-rename":              new(FixerVMwareRename),
		"parallels-headless":         new(FixerParallelsHeadless),
		"parallels-deprecations":     new(FixerParallelsDeprecations),
		"sshkeypath":                 new(FixerSSHKeyPath),
		"sshdisableagent":            new(FixerSSHDisableAgent),
		"scaleway-access-key":        new(FixerScalewayAccessKey),
		"manifest-filename":          new(FixerManifestFilename),
		"amazon-shutdown_behavior":   new(FixerAmazonShutdownBehavior),
		"amazon-enhanced-networking": new(FixerAmazonEnhancedNetworking),
		"amazon-private-ip":          new(FixerAmazonPrivateIP),
		"amazon-temp-sec-cidrs":      new(FixerAmazonTemporarySecurityCIDRs),
		"docker-email":               new(FixerDockerEmail),
		"powershell-escapes":         new(FixerPowerShellEscapes),
		"hyperv-deprecations":        new(FixerHypervDeprecations),
		"hyperv-vmxc-typo":           new(FixerHypervVmxcTypo),
		"hyperv-cpu-and-ram":         new(FizerHypervCPUandRAM),
		"vmware-compaction":          new(FixerVMwareCompaction),
		"clean-image-name":           new(FixerCleanImageName),
		"spot-price-auto-product":    new(FixerAmazonSpotPriceProductDeprecation),
		"qemu-disk-size":             new(FixerQEMUDiskSize),
		"galaxy-command":             new(FixerGalaxyCommand),
		"comm-config":                new(FixerCommConfig),
		"ssh-wait-timeout":           new(FixerSSHTimout),
		"docker-tag-tags":            new(FixerDockerTagtoTags),
		"vsphere-iso-net-disk":       new(FixerVSphereNetworkDisk),
		"iso-checksum-type-and-url":  new(FixerISOChecksumTypeAndURL),
		"qemu-host-port":             new(FixerQEMUHostPort),
		"azure-exclude_from_latest":  new(FixerAzureExcludeFromLatest),
		"proxmox-type":               new(FixerProxmoxType),
	}

	FixerOrder = []string{
		"iso-md5",
		"createtime",
		"virtualbox-gaattach",
		"pp-vagrant-override",
		"virtualbox-rename",
		"vmware-rename",
		"parallels-headless",
		"parallels-deprecations",
		"sshkeypath",
		"sshdisableagent",
		"scaleway-access-key",
		"manifest-filename",
		"amazon-shutdown_behavior",
		"amazon-enhanced-networking",
		"amazon-private-ip",
		"amazon-temp-sec-cidrs",
		"docker-email",
		"docker-tag-tags",
		"powershell-escapes",
		"vmware-compaction",
		"hyperv-deprecations",
		"hyperv-vmxc-typo",
		"hyperv-cpu-and-ram",
		"clean-image-name",
		"spot-price-auto-product",
		"qemu-disk-size",
		"galaxy-command",
		"comm-config",
		"ssh-wait-timeout",
		"vsphere-iso-net-disk",
		"iso-checksum-type-and-url",
		"qemu-host-port",
		"azure-exclude_from_latest",
		"proxmox-type",
	}
}
