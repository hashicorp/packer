---
layout: "docs"
---

# File Provisioner

Type: `file`

The file provisioner uploads files to machines build by Packer.

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
  working directory when ?Packer is executed.

* `destination` (string) - The path where the file will be uploaded to in the 
  machine. This value must be a writable location and any parent directories 
  must already exist.
