---
layout: "docs"
page_title: "File Provisioner"
---

# File Provisioner

Type: `file`

The file provisioner uploads files to machines built by Packer. The
recommended usage of the file provisioner is to use it to upload files,
and then use [shell provisioner](/docs/provisioners/shell.html) to move
them to the proper place, set permissions, etc.

## Basic Example

<pre class="prettyprint">
{
  "type": "file",
  "source": "app.tar.gz",
  "destination": "/tmp/app.tar.gz"
}
</pre>

## Configuration Reference

The available configuration options are listed below. All elements are required.

* `source` (string) - The path to a local file to upload to the machine. The
  path can be absolute or relative. If it is relative, it is relative to the
  working directory when Packer is executed.

* `destination` (string) - The path where the file will be uploaded to in the
  machine. This value must be a writable location and any parent directories
  must already exist.
