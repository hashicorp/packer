---
description: |
    Packer strives to be stable and bug-free, but issues inevitably arise where
    certain things may not work entirely correctly, or may not appear to work
    correctly. In these cases, it is sometimes helpful to see more details about
    what Packer is actually doing.
layout: docs
page_title: Debugging Packer
...

# Debugging Packer Builds

For remote builds with cloud providers like Amazon Web Services AMIs, debugging
a Packer build can be eased greatly with `packer build -debug`. This disables
parallelization and enables debug mode.

Debug mode informs the builders that they should output debugging information.
The exact behavior of debug mode is left to the builder. In general, builders
usually will stop between each step, waiting for keyboard input before
continuing. This will allow you to inspect state and so on.

In debug mode once the remote instance is instantiated, Packer will emit to the
current directory an ephemeral private ssh key as a .pem file. Using that you
can `ssh -i <key.pem>` into the remote build instance and see what is going on
for debugging. The ephemeral key will be deleted at the end of the packer run
during cleanup.

### Windows

As of Packer 0.8.1 the default WinRM communicator will emit the password for a
Remote Desktop Connection into your instance. This happens following the several
minute pause as the instance is booted. Note a .pem key is still created for
securely transmitting the password. Packer automatically decrypts the password
for you in debug mode.

## Debugging Packer

Issues occasionally arise where certain things may not work entirely correctly,
or may not appear to work correctly. In these cases, it is sometimes helpful to
see more details about what Packer is actually doing.

Packer has detailed logs which can be enabled by setting the `PACKER_LOG`
environmental variable to any value like this
`PACKER_LOG=1 packer build <config.json>`. This will cause detailed logs to
appear on stderr. The logs contain log messages from Packer as well as any
plugins that are being used. Log messages from plugins are prefixed by their
application name.

Note that because Packer is highly parallelized, log messages sometimes appear
out of order, especially with respect to plugins. In this case, it is important
to pay attention to the timestamp of the log messages to determine order.

In addition to simply enabling the log, you can set `PACKER_LOG_PATH` in order
to force the log to always go to a specific file when logging is enabled. Note
that even when `PACKER_LOG_PATH` is set, `PACKER_LOG` must be set in order for
any logging to be enabled.

### Debugging Packer in Powershell/Windows

In Windows you can set the detailed logs environmental variable `PACKER_LOG` or
the log variable `PACKER_LOG_PATH` using powershell environment variables. For
example:

    $env:PACKER_LOG=1
    $env:PACKER_LOG_PATH="packerlog.txt"

If you find a bug with Packer, please include the detailed log by using a
service such as [gist](https://gist.github.com).

## Issues Installing Ubuntu Packages

Issues may arise using and building Ubuntu AMIs where common packages that
*should* be installed from Ubuntu's Main repository are not found during a
provisioner step:

    amazon-ebs: No candidate version found for build-essential
    amazon-ebs: No candidate version found for build-essential

This, obviously can cause problems where a build is unable to finish
successfully as the proper packages cannot be provisioned correctly. The problem
arises when cloud-init has not finished fully running on the source AMI by the
time that packer starts any provisioning steps.

Adding the following provisioner to the packer template, allows for the
cloud-init process to fully finish before packer starts provisioning the source
AMI.

    {
      "type": "shell",
      "inline": [
        "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do echo 'Waiting for cloud-init...'; sleep 1; done"
      ]
    }
