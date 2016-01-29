---
description: |
    Packer must first be installed on the machine you want to run it on. To make
    installation easy, Packer is distributed as a binary package for all supported
    platforms and architectures. This page will not cover how to compile Packer from
    source, as that is covered in the README and is only recommended for advanced
    users.
layout: docs
page_title: Install Packer
...

# Install Packer

Packer must first be installed on the machine you want to run it on. To make
installation easy, Packer is distributed as a [binary package](/downloads.html)
for all supported platforms and architectures. This page will not cover how to
compile Packer from source, as that is covered in the
[README](https://github.com/mitchellh/packer/blob/master/README.md) and is only
recommended for advanced users.

## Installing Packer

To install packer, first find the [appropriate package](/downloads.html) for
your system and download it. Packer is packaged as a "zip" file.

Next, unzip the downloaded package into a directory where Packer will be
installed. On Unix systems, `~/packer` or `/usr/local/packer` is generally good,
depending on whether you want to restrict the install to just your user or
install it system-wide. On Windows systems, you can put it wherever you'd like.

After unzipping the package, the directory should contain a set of binary
programs, such as `packer`, `packer-build-amazon-ebs`, etc. The final step to
installation is to make sure the directory you installed Packer to is on the
PATH. See [this
page](https://stackoverflow.com/questions/14637979/how-to-permanently-set-path-on-linux)
for instructions on setting the PATH on Linux and Mac. [This
page](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows)
contains instructions for setting the PATH on Windows.

## Verifying the Installation

After installing Packer, verify the installation worked by opening a new command
prompt or console, and checking that `packer` is available:

``` {.text}
$ packer
usage: packer [--version] [--help] <command> [<args>]

Available commands are:
    build        build image(s) from template
    fix          fixes templates from old versions of packer
    inspect      see components of a template
    validate     check that a template is valid
```

If you get an error that `packer` could not be found, then your PATH
environmental variable was not setup properly. Please go back and ensure that
your PATH variable contains the directory which has Packer installed.

Otherwise, Packer is installed and you're ready to go!

## Alternative Installation Methods

Installation from binary packages is currently the only officially supported
installation method. The binary packages are guaranteed to be the latest
available version and match the proper checksums. However, in addition to the
official binaries, there are other unofficial 3rd party methods of installation
managed by the Packer community:

### Homebrew

If you're using OS X and [Homebrew](http://brew.sh), you can install Packer:

``` {.text}
$ brew install packer
```

### Chocolatey

If you're using Windows and [Chocolatey](http://chocolatey.org), you can install
Packer from Windows command line (cmd). Remember that this is updated by a 3rd
party, so it may not be the latest available version.

``` {.text}
$ choco install packer
```
