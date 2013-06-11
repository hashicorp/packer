---
layout: "docs"
---

# Custom Builder Development

Builders are the components of Packer responsible for creating a machine,
bringing it to a point where it can be provisioned, and then turning
that provisioned machine into some sort of machine image. Several builders
are officially distributed with Packer itself, such as the AMI builder, the
VMware builder, etc. However, it is possible to write custom builders using
the Packer plugin interface, and this page documents how to do that.

Prior to reading this page, it is assumed you have read the page on
[plugin development basics](/docs/extend/developing-plugins.html).

<div class="alert alert-block">
  <strong>Warning!</strong> This is an advanced topic. If you're new to Packer,
  we recommend getting a bit more comfortable before you dive into writing
  plugins.
</div>

## The Interface

The interface that must be implemented for a builder is the `packer.Builder`
interface. It is reproduced below for easy reference. The reference below
also contains some basic documentatin of what each of the methods are
supposed to do.

<pre class="prettyprint">
type Builder interface {
	// Prepare is responsible for reading in some configuration, in the raw form
	// of map[string]interface{}, and storing that state for use later. Any setup
	// should be done in this method. Note that NO side effects should really take
	// place in prepare. It is meant as a state setup step only.
	Prepare(config interface{}) error

	// Run is where the actual build should take place. It takes a Ui to
	// send messages to the user, Hook to execute hooks, and Cache in order
	// to save files across runs.
	Run(Ui, Hook, Cache) Artifact

	// Cancel cancels a possibly running Builder. This should block until
	// the builder actually cancels and cleans up after itself.
	Cancel()
}
</pre>

### The "Prepare" Method

The `Prepare` method for each builder is called prior to any runs with
the configuration that was given in the template. This is passed in as
an `interface{}` type, but is generally `map[string]interface{}`. The prepare
method is responsible for translating this configuration into an internal
structure, validating it, and returning any errors.

For decoding the `interface{}` into a meaningful structure, the
[mapstructure](https://github.com/mitchellh/mapstructure) library is recommended.
Mapstructure will take an `interface{}` and decode it into an arbitrarily
complex struct. If there are any errors, it generates very human friendly
errors that can be returned directly from the prepare method.

While it is not actively enforced, **no side effects** should occur from
running the `Prepare` method. Specifically, don't create files, don't launch
virtual machines, etc. Prepare's purpose is solely to configure the builder
and validate the configuration.

### The "Run" Method

`Run` is where all the interesting stuff happens. Run is executed, often
in parallel for multiple builders, to actually build the machine, provision
it, and create the resulting machine image, which is returned as an
implementation of the `packer.Artifact` interface.

The `Run` method takes three parameters. These are all very useful. The
`packer.Ui` object is used to send output to the console. `packer.Hook` is
used to execute hooks, which are covered in more detail in the hook section
below. And `packer.Cache` is used to store files between multiple Packer
runs, and is covered in more detail in the cache section below.

Because builder runs are typically a complex set of many steps, the
[multistep](https://github.com/mitchellh/multistep) library is recommended
to bring order to the complexity. Multistep is a library which allows you to
separate your logic into multiple distinct "steps" and string them together.
It fully supports cancellation mid-step and so on. Please check it out, it is
how the built-in builders are all implemented.

Finally, as a result of `Run`, an implementation of `packer.Artifact` should
be returned. More details on creating a `packer.Artifact` are covered in the
artifact section below.

### The "Cancel" Method

The `Run` method is often run in parallel. The `Cancel` method can be
called at any time and requests cancellation of any builder run in progress.
This method should block until the run actually stops.

Cancels are most commonly triggered by external interrupts, such as the
user pressing `Ctrl-C`. Packer will only exit once all the builders clean up,
so it is important that you architect your builder in a way that it is quick
to respond to these cancellations and clean up after itself.

## Creating an Artifact

TODO

## Hooks

TODO

## Provisioning

TODO

## Caching Files

TODO


