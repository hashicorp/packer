---
layout: "docs"
page_title: "Packer Plugins - Extend Packer"
---

# Packer Plugins

Plugins allow new functionality to be added to Packer without
modifying the core source code. Packer plugins are able to add new
commands, builders, provisioners, hooks, and more. In fact, much of Packer
itself is implemented by writing plugins that are simply distributed with
Packer. For example, all the commands, builders, provisioners, and more
that ship with Packer are implemented as Plugins that are simply hardcoded
to load with Packer.

This page will cover how to install and use plugins. If you're interested
in developing plugins, the documentation for that is available the
[developing plugins](/docs/extend/developing-plugins.html) page.

Because Packer is so young, there is no official listing of available
Packer plugins. Plugins are best found via Google. Typically, searching
"packer plugin _x_" will find what you're looking for if it exists. As
Packer gets older, an official plugin directory is planned.

## How Plugins Work

Packer plugins are completely separate, standalone applications that the
core of Packer starts and communicates with.

These plugin applications aren't meant to be run manually. Instead, Packer core executes
these plugin applications in a certain way and communicates with them.
For example, the VMware builder is actually a standalone binary named
`packer-builder-vmware`. The next time you run a Packer build, look at
your process list and you should see a handful of `packer-` prefixed
applications running.

## Installing Plugins

Plugins are installed by modifying the [core Packer configuration](/docs/other/core-configuration.html). Within
the core configuration, each component has a key/value mapping of the
plugin name to the actual plugin binary.

For example, if we're adding a new builder for CustomCloud, the core
Packer configuration may look like this:

<pre class="prettyprint">
{
  "builders": {
    "custom-cloud": "packer-builder-custom-cloud"
  }
}
</pre>

In this case, the "custom-cloud" type is the type that is actually used for the value
of the "type" configuration key for the builder definition.

The value, "packer-builder-custom-cloud", is the path to the plugin binary.
It can be an absolute or relative path. If it is not an absolute path, then
the binary is searched for on the PATH. In the example above, Packer will
search for `packer-builder-custom-cloud` on the PATH.

After adding the plugin to the core Packer configuration, it is immediately
available on the next run of Packer. To uninstall a plugin, just remove it
from the core Packer configuration.

In addition to builders, other types of plugins can be installed. The full
list is below:

* `builders` - A key/value pair of builder type to the builder plugin
  application.

* `commands` - A key/value pair of the command name to the command plugin
  application. The command name is what is executed on the command line, like
  `packer COMMAND`.

* `provisioners` - A key/value pair of the provisioner type to the
  provisioner plugin application. The provisioner type is the value of the
  "type" configuration used within templates.
