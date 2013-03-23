# Packer

Packer is a tool for building identical machine images across multiple clouds.

Packer provides a framework and configuration format for creating identical
machine images to launch into any environment, such as VirtualBox, VMware,
Amazon EC2, etc. Because this build process is automated, you can develop in
VirtualBox, then deploy to EC2 with an identical image.

## Developing Packer

If you wish to work on Packer itself, you'll first need [Go](http://golang.org)
installed. Next, clone this repository and source "setup.sh" in your shell. This
will set up the environmental variables properly to work on Packer. After
that, just run `make`. Commands:

```
$ source setup.sh
...
$ make
...
$ bin/packer --version
```
