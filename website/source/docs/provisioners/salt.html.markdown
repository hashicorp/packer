---
layout: "docs"
---

# Salt Provisioner

Type: `salt`

The salt provisioner provisions machines built by Packer using [Salt](http://saltstack.com/) states.

## Basic Example

The example below is fully functional.

<pre class="prettyprint">
{
    "type": "salt",
    "bootstrap_args": "git v0.14.0"
}
</pre>

## Configuration Reference

The reference of available configuration options is listed below:

* `skip_bootstrap` (boolean) - By default Packer runs [salt bootstrap](https://github.com/saltstack/salt-bootstrap) to install salt. Set this to true to skip this step.

* `boostrap_args` (string) -
  Arguments to send to the bootstrap script. Usage is somewhat documented on [github](https://github.com/saltstack/salt-bootstrap), but the [script itself](https://github.com/saltstack/salt-bootstrap/blob/develop/bootstrap-salt.sh) has more detailed usage instructions. Default is no arguments.
