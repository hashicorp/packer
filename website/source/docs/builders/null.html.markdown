---
layout: "docs"
page_title: "Null Builder"
description: |-
  The `null` Packer builder is not really a builder, it just sets up an SSH connection and runs the provisioners. It can be used to debug provisioners without incurring high wait times. It does not create any kind of image or artifact.
---

# Null Builder

Type: `null`

The `null` Packer builder is not really a builder, it just sets up an SSH connection
and runs the provisioners. It can be used to debug provisioners without
incurring high wait times. It does not create any kind of image or artifact.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful, since
no provisioners are defined, but it will connect to the specified host via ssh.

```javascript
{
  "type":     "null",
  "host":     "127.0.0.1",
  "ssh_username": "foo",
  "ssh_password": "bar"
}
```

## Configuration Reference

Configuration options are organized into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

### Required:

* `host` (string) - The hostname or IP address to connect to.

* `ssh_password` (string) - The password to be used for the ssh connection.
  Cannot be combined with ssh_private_key_file.

* `ssh_private_key_file` (string) - The filename of the ssh private key to be
  used for the ssh connection. E.g. /home/user/.ssh/identity_rsa.

* `ssh_username` (string) - The username to be used for the ssh connection.

### Optional:

* `port` (integer) - ssh port to connect to, defaults to 22.

