---
layout: "docs"
page_title: "Parallels Builder"
---

# Parallels Builder

The Parallels builder is able to create [Parallels Desktop for Mac](http://www.parallels.com/products/desktop/) virtual machines and export them in the PVM format.

Packer actually comes with multiple builders able to create Parallels
machines, depending on the strategy you want to use to build the image.
Packer supports the following Parallels builders:

* [parallels-iso](/docs/builders/parallels-iso.html) - Starts from
  an ISO file, creates a brand new Parallels VM, installs an OS,
  provisions software within the OS, then exports that machine to create
  an image. This is best for people who want to start from scratch.

* [parallels-pvm](/docs/builders/parallels-pvm.html) - This builder
  imports an existing PVM file, runs provisioners on top of that VM,
  and exports that machine to create an image. This is best if you have
  an existing Parallels VM export you want to use as the source. As an
  additional benefit, you can feed the artifact of this builder back into
  itself to iterate on a machine.


## Requirements

In addition to [Parallels Desktop for Mac](http://www.parallels.com/products/desktop/) this requires:

- [Parallels Virtualization SDK 9 for Mac](http://download.parallels.com//desktop/v9/pde.hf1/ParallelsVirtualizationSDK-9.0.24172.951362.dmg)
- [prl-utils](https://github.com/rickard-von-essen/prl-utils/)

The SDK can be installed by downloading and following the instructions in the dmg. The easiest way to install _prl-utils_ is using [Homebrew](http://brew.sh/)

  ```
  brew tap rickard-von-essen/homebrew-formulae
  brew install --HEAD prl-utils
  ```
