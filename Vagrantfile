# -*- mode: ruby -*-
# vi: set ft=ruby :

LINUX_BASE_BOX = "bento/ubuntu-16.04"
FREEBSD_BASE_BOX = "jen20/FreeBSD-12.0-CURRENT"

Vagrant.configure(2) do |config|
	# Compilation and development boxes
	config.vm.define "linux", autostart: true, primary: true do |vmCfg|
		vmCfg.vm.box = LINUX_BASE_BOX
		vmCfg.vm.hostname = "linux"
		vmCfg = configureProviders vmCfg,
			cpus: suggestedCPUCores()

		vmCfg.vm.synced_folder ".", "/vagrant", disabled: true
		vmCfg.vm.synced_folder '.',
			'/opt/gopath/src/github.com/hashicorp/packer'

		vmCfg.vm.provision "shell",
			privileged: true,
			inline: 'rm -f /home/vagrant/linux.iso'

		vmCfg.vm.provision "shell",
			privileged: true,
			path: './scripts/vagrant-linux-priv-go.sh'

		vmCfg.vm.provision "shell",
			privileged: true,
			path: './scripts/vagrant-linux-priv-config.sh'

		vmCfg.vm.provision "shell",
			privileged: false,
			path: './scripts/vagrant-linux-unpriv-bootstrap.sh'
	end

	config.vm.define "freebsd", autostart: false, primary: false do |vmCfg|
		vmCfg.vm.box = FREEBSD_BASE_BOX
		vmCfg.vm.hostname = "freebsd"
		vmCfg = configureProviders vmCfg,
			cpus: suggestedCPUCores()

		vmCfg.vm.synced_folder ".", "/vagrant", disabled: true
		vmCfg.vm.synced_folder '.',
			'/opt/gopath/src/github.com/hashicorp/packer',
			type: "nfs",
			bsd__nfs_options: ['noatime']

		vmCfg.vm.provision "shell",
			privileged: true,
			path: './scripts/vagrant-freebsd-priv-config.sh'

		vmCfg.vm.provision "shell",
			privileged: false,
			path: './scripts/vagrant-freebsd-unpriv-bootstrap.sh'
	end
end

def configureProviders(vmCfg, cpus: "2", memory: "2048")
	vmCfg.vm.provider "virtualbox" do |v|
		v.memory = memory
		v.cpus = cpus
	end

	["vmware_fusion", "vmware_workstation"].each do |p|
		vmCfg.vm.provider p do |v|
			v.enable_vmrun_ip_lookup = false
			v.vmx["memsize"] = memory
			v.vmx["numvcpus"] = cpus
		end
	end

	return vmCfg
end

def suggestedCPUCores()
	case RbConfig::CONFIG['host_os']
	when /darwin/
		Integer(`sysctl -n hw.ncpu`) / 2
	when /linux/
		Integer(`cat /proc/cpuinfo | grep processor | wc -l`) / 2
	else
		2
	end
end
