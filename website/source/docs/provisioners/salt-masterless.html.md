---
description: |
    The salt-masterless Packer provisioner provisions machines built by Packer
    using Salt states, without connecting to a Salt master.
layout: docs
page_title: 'Salt Masterless - Provisioners'
sidebar_current: 'docs-provisioners-salt-masterless'
---

# Salt Masterless Provisioner

Type: `salt-masterless`

The `salt-masterless` Packer provisioner provisions machines built by Packer
using [Salt](http://saltstack.com/) states, without connecting to a Salt master.

## Basic Example

The example below is fully functional.

``` json
{
  "type": "salt-masterless",
  "local_state_tree": "/Users/me/salt"
}
```

## Configuration Reference

The reference of available configuration options is listed below. The only
required element is "local_state_tree".

Required:

-   `local_state_tree` (string) - The path to your local [state
    tree](http://docs.saltstack.com/ref/states/highstate.html#the-salt-state-tree).
    This will be uploaded to the `remote_state_tree` on the remote.

Optional:

-   `bootstrap_args` (string) - Arguments to send to the bootstrap script. Usage
    is somewhat documented on
    [github](https://github.com/saltstack/salt-bootstrap), but the [script
    itself](https://github.com/saltstack/salt-bootstrap/blob/develop/bootstrap-salt.sh)
    has more detailed usage instructions. By default, no arguments are sent to
    the script.

-   `disable_sudo` (boolean) - By default, the bootstrap install command is prefixed with `sudo`. When using a
    Docker builder, you will likely want to pass `true` since `sudo` is often not pre-installed.

-   `remote_pillar_roots` (string) - The path to your remote [pillar
    roots](http://docs.saltstack.com/ref/configuration/master.html#pillar-configuration).
    default: `/srv/pillar`. This option cannot be used with `minion_config`.

-   `remote_state_tree` (string) - The path to your remote [state
    tree](http://docs.saltstack.com/ref/states/highstate.html#the-salt-state-tree).
    default: `/srv/salt`. This option cannot be used with `minion_config`.

-   `local_pillar_roots` (string) - The path to your local [pillar
    roots](http://docs.saltstack.com/ref/configuration/master.html#pillar-configuration).
    This will be uploaded to the `remote_pillar_roots` on the remote.

-   `custom_state` (string) - A state to be run instead of `state.highstate`.
    Defaults to `state.highstate` if unspecified.

-   `minion_config` (string) - The path to your local [minion config
    file](http://docs.saltstack.com/ref/configuration/minion.html). This will be
    uploaded to the `/etc/salt` on the remote. This option overrides the
    `remote_state_tree` or `remote_pillar_roots` options.

-   `grains_file` (string) - The path to your local [grains file](https://docs.saltstack.com/en/latest/topics/grains). This will be
    uploaded to `/etc/salt/grains` on the remote.

-   `skip_bootstrap` (boolean) - By default the salt provisioner runs [salt
    bootstrap](https://github.com/saltstack/salt-bootstrap) to install salt. Set
    this to true to skip this step.

-   `temp_config_dir` (string) - Where your local state tree will be copied
    before moving to the `/srv/salt` directory. Default is `/tmp/salt`.

-   `no_exit_on_failure` (boolean) - Packer will exit if the `salt-call` command
    fails. Set this option to true to ignore Salt failures.

-   `log_level` (string) - Set the logging level for the `salt-call` run.

-   `salt_call_args` (string) - Additional arguments to pass directly to `salt-call`. See
    [salt-call](https://docs.saltstack.com/ref/cli/salt-call.html) documentation for more
    information. By default no additional arguments (besides the ones Packer generates)
    are passed to `salt-call`.

-   `salt_bin_dir` (string) - Path to the `salt-call` executable. Useful if it is not
    on the PATH.
