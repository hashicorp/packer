#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


export DEBIAN_FRONTEND=noninteractive

# Update and ensure we have apt-add-repository
apt-get update
apt-get install -y software-properties-common

apt-get install -y bzr \
	curl \
	gcc \
	git \
	make \
	mercurial \
	zip

# Ensure we cd into the working directory on login
if ! grep "cd /opt/gopath/src/github.com/hashicorp/packer" /home/vagrant/.profile ; then
	echo 'cd /opt/gopath/src/github.com/hashicorp/packer' >> /home/vagrant/.profile
fi
