#!/bin/zsh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


## This script is to be run before a Packer release in order to update
## all vendored plugins to the latest available release.
## The SDK is included in the plugin list and will be upgraded as well if a
## newest version is available.
## This script should be run in packer's root.

declare -a plugins=(
	"hashicorp/packer-plugin-alicloud"
	"hashicorp/packer-plugin-amazon"
	"hashicorp/packer-plugin-ansible"
	"hashicorp/packer-plugin-azure"
	"hashicorp/packer-plugin-chef"
	"hashicorp/packer-plugin-cloudstack"
	"hashicorp/packer-plugin-converge"
	"digitalocean/packer-plugin-digitalocean"
	"hashicorp/packer-plugin-docker"
	"hashicorp/packer-plugin-googlecompute"
	"hashicorp/packer-plugin-hcloud"
	"hashicorp/packer-plugin-hyperone"
	"hashicorp/packer-plugin-hyperv"
	"hashicorp/packer-plugin-jdcloud"
	"hashicorp/packer-plugin-linode"
	"hashicorp/packer-plugin-lxc"
	"hashicorp/packer-plugin-lxd"
	"hashicorp/packer-plugin-ncloud"
	"hashicorp/packer-plugin-openstack"
	"hashicorp/packer-plugin-oneandone"
	"hashicorp/packer-plugin-parallels"
	"hashicorp/packer-plugin-profitbricks"
	"hashicorp/packer-plugin-proxmox"
	"hashicorp/packer-plugin-puppet"
	"hashicorp/packer-plugin-qemu"
	"hashicorp/packer-plugin-sdk"
	"hashicorp/packer-plugin-tencentcloud"
	"hashicorp/packer-plugin-triton"
	"hashicorp/packer-plugin-ucloud"
	"hashicorp/packer-plugin-vagrant"
	"hashicorp/packer-plugin-virtualbox"
	"hashicorp/packer-plugin-vmware"
	"hashicorp/packer-plugin-vsphere"
	"hashicorp/packer-plugin-yandex"
)

## now loop through the above plugin array
## update the plugins and the SDK to the latest available version
for i in "${plugins[@]}"
do
   happy=false
   while ! $happy
    do
      echo "upgrading $i"
      output=$(go get -d github.com/$i)
      happy=true
      if [[ $output == *"443: Connection refused"*  ]]; then
        echo "Try again after 5 seconds"
        sleep 5
        happy=false
      fi
    done
   sleep 1
done

go mod tidy -compat=1.18
