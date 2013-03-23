# Packer

Packer is a tool for building identical machine images across multiple clouds.

Packer provides a framework and configuration format for creating identical
machine images to launch into any environment, such as VirtualBox, VMware,
Amazon EC2, etc. Because this build process is automated, you can develop in
VirtualBox, then deploy to EC2 with an identical image.

## Developing Packer

If you wish to work on Packer itself, you'll first need [Go](http://golang.org)
installed. Next, clone this repository then just type `make`.

```
$ make
...
$ bin/packer --version
```
