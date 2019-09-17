---
layout: intro
sidebar_current: intro-getting-started-install
page_title: Install Packer - Getting Started
description: |-
  Packer must first be installed on the machine you want to run it on. To make
  installation easier, Packer is distributed as a binary package for all supported
  platforms and architectures. This page will not cover how to compile Packer
  from source, as that is covered in the README and is only recommended for
  advanced users.
---

# Install Options

Packer may be installed in the following ways:

1. Using a [precompiled binary](#precompiled-binaries); We release binaries
  for all supported platforms and architectures. This method is recommended for
  most users.

2. Installing [from source](#compiling-from-source) This method is only
  recommended for advanced users.

3. An unoffical [alternative installation method](#alternative-installation-methods)

## Precompiled Binaries

To install the precompiled binary, [download](/downloads.html) the appropriate
package for your system. Packer is currently packaged as a zip file. We do not
have any near term plans to provide system packages.

Next, unzip the downloaded package into a directory where Packer will be
installed. On Unix systems, `~/packer` or `/usr/local/packer` is generally good,
depending on whether you want to restrict the install to just your user or
install it system-wide. If you intend to access it from the command-line, make
sure to place it somewhere on your `PATH` before `/usr/sbin`. On Windows
systems, you can put it wherever you'd like. The `packer` (or `packer.exe` for
Windows) binary inside is all that is necessary to run Packer. Any additional
files aren't required to run Packer.

After unzipping the package, the directory should contain a single binary
program called `packer`. The final step to
installation is to make sure the directory you installed Packer to is on the
PATH. See [this
page](https://stackoverflow.com/questions/14637979/how-to-permanently-set-path-on-linux)
for instructions on setting the PATH on Linux and Mac. [This
page](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows)
contains instructions for setting the PATH on Windows.

## Compiling from Source

To compile from source, you will need [Go](https://golang.org) installed and
configured properly as well as a copy of [`git`](https://www.git-scm.com/)
in your `PATH`.

1.  Clone the Packer repository from GitHub into your `GOPATH`:

    ``` shell
    $ mkdir -p $(go env GOPATH)/src/github.com/hashicorp && cd $_
    $ git clone https://github.com/hashicorp/packer.git
    $ cd packer
    ```

2.  Build Packer for your current system and put the
    binary in `./bin/` (relative to the git checkout). The `make dev` target is
    just a shortcut that builds `packer` for only your local build environment (no
    cross-compiled targets).

    ``` shell
    $ make dev
    ```

## Verifying the Installation

After installing Packer, verify the installation worked by opening a new command
prompt or console, and checking that `packer` is available:

```text
$ packer
usage: packer [--version] [--help] <command> [<args>]

Available commands are:
    build       build image(s) from template
    fix         fixes templates from old versions of packer
    inspect     see components of a template
    validate    check that a template is valid
    version     Prints the Packer version
```

If you get an error that `packer` could not be found, then your PATH environment
variable was not setup properly. Please go back and ensure that your PATH
variable contains the directory which has Packer installed.

Otherwise, Packer is installed and you're ready to go!

## Alternative Installation Methods

While the binary packages are the only official method of installation, there are
alternatives available.

### Homebrew

If you're using OS X and [Homebrew](http://brew.sh), you can install Packer by
running:

    $ brew install packer

### Chocolatey

If you're using Windows and [Chocolatey](http://chocolatey.org), you can
install Packer by running:

    choco install packer

## Troubleshooting

On some *RedHat*-based Linux distributions there is another tool named `packer`
installed by default. You can check for this using `which -a packer`. If you get
an error like this it indicates there is a name conflict.

    $ packer
    /usr/share/cracklib/pw_dict.pwd: Permission denied
    /usr/share/cracklib/pw_dict: Permission denied

To fix this, you can create a symlink to packer that uses a different name like
`packer.io`, or invoke the `packer` binary you want using its absolute path,
e.g. `/usr/local/packer`.

[Continue to building an image](./build-image.html)
