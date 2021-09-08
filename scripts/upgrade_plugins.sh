#!/bin/zsh

## This script is to be run before a Packer release in order to update
## all vendored plugins to the latest available release.
## The SDK is included in the plugin list and will be upgraded as well if a
## newest version is available.
## This script should be run in packer's root.

declare -a plugins=(
	"alicloud"
	"amazon"
	"ansible"
	"azure"
	"chef"
	"cloudstack"
	"converge"
	"digitalocean"
	"docker"
	"googlecompute"
	"hcloud"
	"hyperone"
	"hyperv"
	"jdcloud"
	"linode"
	"lxc"
	"lxd"
	"ncloud"
	"openstack"
	"oracle"
	"outscale"
	"oneandone"
	"parallels"
	"profitbricks"
	"proxmox"
	"puppet"
	"qemu"
	"scaleway"
	"sdk"
	"tencentcloud"
	"triton"
	"ucloud"
	"vagrant"
	"virtualbox"
	"vmware"
	"vsphere"
	"yandex"
)

## now loop through the above plugin array
## update the plugins and the SDK to the latest available version
for i in "${plugins[@]}"
do
   happy=false
   while ! $happy
    do
      echo "upgrading $i"
      output=$(go get -d github.com/hashicorp/packer-plugin-$i)
      happy=true
      if [[ $output == *"443: Connection refused"*  ]]; then
        echo "Try again after 5 seconds"
        sleep 5
        happy=false
      fi
    done
   sleep 1
done

go mod tidy
