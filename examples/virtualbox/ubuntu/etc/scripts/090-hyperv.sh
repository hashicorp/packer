#!/bin/sh -eux
ubuntu_version="`lsb_release -r | awk '{print $2}'`";
major_version="`echo $ubuntu_version | awk -F. '{print $1}'`";

case "$PACKER_BUILDER_TYPE" in
hyperv-iso)
  if [ "$major_version" -eq "16" ]; then
    apt-get install -y linux-tools-virtual-lts-xenial linux-cloud-tools-virtual-lts-xenial;
  else
    apt-get -y install linux-image-virtual linux-tools-virtual linux-cloud-tools-virtual;
  fi
esac
