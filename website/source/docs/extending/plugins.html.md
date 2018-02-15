---
description: |
    Packer Plugins allow new functionality to be added to Packer without modifying
    the core source code. Packer plugins are able to add new builders,
    provisioners, hooks, and more.
layout: docs
page_title: 'Plugins - Extending'
sidebar_current: 'docs-extending-plugins'
---

# Plugins

Packer Plugins allow new functionality to be added to Packer without modifying
the core source code. Packer plugins are able to add new builders,
provisioners, hooks, and more. In fact, much of Packer itself is implemented by
writing plugins that are simply distributed with Packer. For example, all the
builders, provisioners, and more that ship with Packer are implemented
as Plugins that are simply hardcoded to load with Packer.

This section will cover how to install and use plugins. If you're interested in
developing plugins, the documentation for that is available below, in the [developing
plugins](#developing-plugins) section.

Because Packer is so young, there is no official listing of available Packer
plugins. Plugins are best found via Google. Typically, searching "packer plugin
*x*" will find what you're looking for if it exists. As Packer gets older, an
official plugin directory is planned.

## How Plugins Work

Packer plugins are completely separate, standalone applications that the core of
Packer starts and communicates with.

These plugin applications aren't meant to be run manually. Instead, Packer core
executes them as a sub-process, run as a sub-command (`packer plugin`) and
communicates with them. For example, the VMware builder is actually run as
`packer plugin packer-builder-vmware`. The next time you run a Packer build,
look at your process list and you should see a handful of `packer-` prefixed
applications running.

## Installing Plugins

The easiest way to install a plugin is to name it correctly, then place it in
the proper directory. To name a plugin correctly, make sure the binary is named
`packer-TYPE-NAME`. For example, `packer-builder-amazon-ebs` for a "builder"
type plugin named "amazon-ebs". Valid types for plugins are down this page more.

Once the plugin is named properly, Packer automatically discovers plugins in the
following directories in the given order. If a conflicting plugin is found
later, it will take precedence over one found earlier.

1.  The directory where `packer` is, or the executable directory.

2.  `~/.packer.d/plugins` on Unix systems or `%APPDATA%/packer.d/plugins`
    on Windows.

3.  The current working directory.

The valid types for plugins are:

-   `builder` - Plugins responsible for building images for a specific platform.

-   `post-processor` - A post-processor responsible for taking an artifact from
    a builder and turning it into something else.

-   `provisioner` - A provisioner to install software on images created by
    a builder.

## Developing Plugins

This page will document how you can develop your own Packer plugins. Prior to
reading this, it is assumed that you're comfortable with Packer and also know
the [basics of how Plugins work](/docs/extending/plugins.html), from a user
standpoint.

Packer plugins must be written in [Go](https://golang.org/), so it is also
assumed that you're familiar with the language. This page will not be a Go
language tutorial. Thankfully, if you are familiar with Go, the Go toolchain
provides many conveniences to help to develop Packer plugins.

~&gt; **Warning!** This is an advanced topic. If you're new to Packer, we
recommend getting a bit more comfortable before you dive into writing plugins.

### Plugin System Architecture

Packer has a fairly unique plugin architecture. Instead of loading plugins
directly into a running application, Packer runs each plugin as a *separate
application*. Inter-process communication and RPC is then used to communicate
between the many running Packer processes. Packer core itself is responsible for
orchestrating the processes and handles cleanup.

The beauty of this is that your plugin can have any dependencies it wants.
Dependencies don't need to line up with what Packer core or any other plugin
uses, because they're completely isolated into the process space of the plugin
itself.

And, thanks to Go's
[interfaces](https://golang.org/doc/effective_go.html#interfaces_and_types), it
doesn't even look like inter-process communication is occurring. You just use
the interfaces like normal, but in fact they're being executed in a remote
process. Pretty cool.

### Plugin Development Basics

Developing a plugin allows you to create additional functionality for Packer.
All the various kinds of plugins have a corresponding interface. The plugin needs
to implement this interface and expose it using the Packer plugin package
(covered here shortly), and that's it!

There are two packages that really matter that every plugin must use. Other than
the following two packages, you're encouraged to use whatever packages you want.
Because plugins are their own processes, there is no danger of colliding
dependencies.

-   `github.com/hashicorp/packer` - Contains all the interfaces that you have to
    implement for any given plugin.

-   `github.com/hashicorp/packer/packer/plugin` - Contains the code to serve
    the plugin. This handles all the inter-process communication stuff.

There are two steps involved in creating a plugin:

1.  Implement the desired interface. For example, if you're building a builder
    plugin, implement the `packer.Builder` interface.

2.  Serve the interface by calling the appropriate plugin serving method in your
    main method. In the case of a builder, this is `plugin.RegisterBuilder`.

A basic example is shown below. In this example, assume the `Builder` struct
implements the `packer.Builder` interface:

``` go
import (
  "github.com/hashicorp/packer/packer/plugin"
)

// Assume this implements packer.Builder
type Builder struct{}

func main() {
  plugin.RegisterBuilder(new(Builder))
}
```

**That's it!** `plugin.RegisterBuilder` handles all the nitty gritty of
communicating with Packer core and serving your builder over RPC. It can't get
much easier than that.

Next, just build your plugin like a normal Go application, using `go build` or
however you please. The resulting binary is the plugin that can be installed
using standard installation procedures.

The specifics of how to implement each type of interface are covered in the
relevant subsections available in the navigation to the left.

~&gt; **Lock your dependencies!** Using `govendor` is highly recommended since
the Packer codebase will continue to improve, potentially breaking APIs along
the way until there is a stable release. By locking your dependencies, your
plugins will continue to work with the version of Packer you lock to.

### Logging and Debugging

Plugins can use the standard Go `log` package to log. Anything logged using this
will be available in the Packer log files automatically. The Packer log is
visible on stderr when the `PACKER_LOG` environmental is set.

Packer will prefix any logs from plugins with the path to that plugin to make it
identifiable where the logs come from. Some example logs are shown below:

``` text
2013/06/10 21:44:43 Loading builder: custom
2013/06/10 21:44:43 packer-builder-custom: 2013/06/10 21:44:43 Plugin minimum port: 10000
2013/06/10 21:44:43 packer-builder-custom: 2013/06/10 21:44:43 Plugin maximum port: 25000
2013/06/10 21:44:43 packer-builder-custom: 2013/06/10 21:44:43 Plugin address: :10000
```

As you can see, the log messages from the custom builder plugin are prefixed
with "packer-builder-custom". Log output is *extremely* helpful in debugging
issues and you're encouraged to be as verbose as you need to be in order for the
logs to be helpful.

### Plugin Development Tips

Here are some tips for developing plugins, often answering common questions or
concerns.

#### Naming Conventions

It is standard practice to name the resulting plugin application in the format
of `packer-TYPE-NAME`. For example, if you're building a new builder for
CustomCloud, it would be standard practice to name the resulting plugin
`packer-builder-custom-cloud`. This naming convention helps users identify the
purpose of a plugin.

#### Testing Plugins

While developing plugins, you can configure your Packer configuration to point
directly to the compiled plugin in order to test it. For example, building the
CustomCloud plugin, I may configure packer like so:

``` json
{
  "builders": {
    "custom-cloud": "/an/absolute/path/to/packer-builder-custom-cloud"
  }
}
```

This would configure Packer to have the "custom-cloud" plugin, and execute the
binary that I am building during development. This is extremely useful during
development.

#### Distributing Plugins

It is recommended you use a tool like [goxc](https://github.com/laher/goxc) in
order to cross-compile your plugin for every platform that Packer supports,
since Go applications are platform-specific. goxc will allow you to build for
every platform from your own computer.
