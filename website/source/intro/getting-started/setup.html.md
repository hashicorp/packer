---
description: |
    Packer must first be installed on the machine you want to run it on. To make
    installation easy, Packer is distributed as a binary package for all supported
    platforms and architectures. This page will not cover how to compile Packer from
    source, as that is covered in the README and is only recommended for advanced
    users.
layout: intro
next_title: Build an Image
next_url: '/intro/getting-started/build-image.html'
page_title: Install Packer
prev_url: '/intro/platforms.html'
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

After unzipping the package, the directory should contain a single binary
program called `packer`. The final step to
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
    build       build image(s) from template
    fix         fixes templates from old versions of packer
    inspect     see components of a template
    push        push template files to a Packer build service
    validate    check that a template is valid
    version     Prints the Packer version
```

If you get an error that `packer` could not be found, then your PATH environment
variable was not setup properly. Please go back and ensure that your PATH
variable contains the directory which has Packer installed.

Otherwise, Packer is installed and you're ready to go!

## Alternative Installation Methods

While the binary packages is the only official method of installation, there are
alternatives available.

### Homebrew

If you're using OS X and [Homebrew](http://brew.sh), you can install Packer:

    $ brew install packer

## Troubleshooting

On some RedHat-based Linux distributions there is another tool named `packer`
installed by default. You can check for this using `which -a packer`. If you get
an error like this it indicates there is a name conflict.

    $ packer
    /usr/share/cracklib/pw_dict.pwd: Permission denied
    /usr/share/cracklib/pw_dict: Permission denied

To fix this, you can create a symlink to packer that uses a different name like
`packer.io`, or invoke the `packer` binary you want using its absolute path,
e.g. `/usr/local/packer`.
