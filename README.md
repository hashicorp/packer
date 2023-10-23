# Packer
[![License: BUSL-1.1](https://img.shields.io/badge/License-BUSL--1.1-yellow.svg)](LICENSE)
[![Build Status](https://github.com/hashicorp/packer/actions/workflows/build.yml/badge.svg)](https://github.com/hashicorp/packer/actions/workflows/build.yml)
[![Discuss](https://img.shields.io/badge/discuss-packer-3d89ff?style=flat)](https://discuss.hashicorp.com/c/packer)
===

<p align="center" style="text-align:center;">
  <a href="https://www.packer.io">
    <img alt="HashiCorp Packer logo" src="website/public/img/logo-packer-padded.svg" width="500" />
  </a>
</p>

Packer is a tool for building identical machine images for multiple platforms
from a single source configuration.

Packer is lightweight, runs on every major operating system, and is highly
performant, creating machine images for multiple platforms in parallel. Packer
supports various platforms through external plugin integrations, the full list of which can
be found at https://developer.hashicorp.com/packer/integrations.

The images that Packer creates can easily be turned into [Vagrant](http://www.vagrantup.com) boxes.

## Quick Start

### Packer 

There is a great [introduction and getting started guide](https://learn.hashicorp.com/tutorials/packer/get-started-install-cli)
for building a Docker image on your local machine without using any paid cloud resources. 

Alternatively, you can refer to [getting started with AWS](https://developer.hashicorp.com/packer/tutorials/aws-get-started) to
learn how to build a machine image for an external cloud provider. 

### HCP Packer

HCP Packer registry stores Packer image metadata, enabling you to track your image lifecycle. 

To get started with building an AWS machine image to HCP Packer for referencing in Terraform refer
to the collection of [HCP Packer Tutorials](https://developer.hashicorp.com/packer/tutorials/hcp-get-started).

## Documentation

Comprehensive documentation is viewable on the Packer website at https://developer.hashicorp.com/packer/docs.

## Contributing to Packer

See
[CONTRIBUTING.md](https://github.com/hashicorp/packer/blob/master/.github/CONTRIBUTING.md)
for best practices and instructions on setting up your development environment
to work on Packer.

## Unmaintained Plugins
As contributors' circumstances change, development on a community maintained
plugin can slow. When this happens, HashiCorp may use GitHub's option to archive the 
plugin’s repository, to clearly signal the plugin's status to users.

What does **unmaintained** mean?

1. The code repository and all commit history will still be available.
1. Documentation will remain on the Packer website.
1. Issues and pull requests are monitored as a best effort.
1. No active development will be performed by HashiCorp.

If you are interested in maintaining an unmaintained or archived plugin, please reach out to us at packer@hashicorp.com.


