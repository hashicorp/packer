---
description: |
    The file Packer provisioner uploads files to machines built by Packer. The
    recommended usage of the file provisioner is to use it to upload files, and then
    use shell provisioner to move them to the proper place, set permissions, etc.
layout: docs
page_title: File Provisioner
...

# File Provisioner

Type: `file`

The file Packer provisioner uploads files to machines built by Packer. The
recommended usage of the file provisioner is to use it to upload files, and then
use [shell provisioner](/docs/provisioners/shell.html) to move them to the
proper place, set permissions, etc.

The file provisioner can upload both single files and complete directories.

## Basic Example

``` {.javascript}
{
  "type": "file",
  "source": "app.tar.gz",
  "destination": "/tmp/app.tar.gz"
}
```

## Configuration Reference

The available configuration options are listed below. All elements are required.

-   `source` (string) - The path to a local file or directory to upload to
    the machine. The path can be absolute or relative. If it is relative, it is
    relative to the working directory when Packer is executed. If this is a
    directory, the existence of a trailing slash is important. Read below on
    uploading directories.

-   `destination` (string) - The path where the file will be uploaded to in
    the machine. This value must be a writable location and any parent
    directories must already exist.

-   `direction` (string) - The direction of the file transfer. This defaults to
    "upload." If it is set to "download" then the file "source" in the machine
    will be downloaded locally to "destination"

## Directory Uploads

The file provisioner is also able to upload a complete directory to the remote
machine. When uploading a directory, there are a few important things you should
know.

First, the destination directory must already exist. If you need to create it,
use a shell provisioner just prior to the file provisioner in order to create
the directory.

Next, the existence of a trailing slash on the source path will determine
whether the directory name will be embedded within the destination, or whether
the destination will be created. An example explains this best:

If the source is `/foo` (no trailing slash), and the destination is `/tmp`, then
the contents of `/foo` on the local machine will be uploaded to `/tmp/foo` on
the remote machine. The `foo` directory on the remote machine will be created by
Packer.

If the source, however, is `/foo/` (a trailing slash is present), and the
destination is `/tmp`, then the contents of `/foo` will be uploaded into `/tmp`
directly.

This behavior was adopted from the standard behavior of rsync. Note that under
the covers, rsync may or may not be used.
