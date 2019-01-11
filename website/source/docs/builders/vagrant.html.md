The Vagrant builder is intended for building new boxes from already-existing
boxes. Your source should be a URL or path to a .box file or a Vagrant Cloud
box name such as `hashicorp/precise64`. This builder is not currently intended
to work with an already-set-up Vagrant environment.

Packer will not install vagrant, nor will it install the underlying
virtualization platforms or extra providers; We expect when you run this
builder that you have already installed what you need.

Required:

`source_box` (string) - URL of the vagrant box to use, or the name of the
vagrant box. For example, `hashicorp/precise64` or
`https://boxes.company.com/my-company.box` are valid source boxes. If using a URL like the latter example above, you will also need to provide a `box_name`.

Optional:

`output_dir` (string) - The directory to create that will contain
your output box. We always create this directory and run from inside of it to
prevent Vagrant init collisions. If unset, it will be set to `packer-` plus
your buildname.

`box_name` (string) - if your source\_box is a boxfile that we need to add to
Vagrant, this is the name to give it.

`vagrantfile_template` (string) - a path to an ERB template to use for the
vagrantfile when calling `vagrant init`. See the blog post
[here](https://www.hashicorp.com/blog/hashicorp-vagrant-2-0-2#customized-vagrantfile-templates)
for some more details on how this works. Available variables are `box_name`,
`box_url`, and `box_version`.

`teardown_method` (string) - Whether to halt, suspend, or destroy the box when
the build has completed. Defaults to "halt"

`box_version` (string) - What box version to use when initializing Vagrant.

`init_minimal` (bool) - If true, will add the --minimal flag to the Vagrant
init command, creating a minimal vagrantfile instead of one filled with helpful
comments.

`add_cacert` (string) - Equivalent to setting the
[`--cacert`](https://www.vagrantup.com/docs/cli/box.html#cacert-certfile)
option in `vagrant add`; defaults to unset.

`add_capath` (string) - Equivalent to setting the
[`--capath`](https://www.vagrantup.com/docs/cli/box.html#capath-certdir) option
in `vagrant add`; defaults to unset.

`add_cert` (string) - Equivalent to setting the
[`--cert`](https://www.vagrantup.com/docs/cli/box.html#cert-certfile) option in
`vagrant add`; defaults to unset.

`add_clean` (bool) - Equivalent to setting the
[`--clean`](https://www.vagrantup.com/docs/cli/box.html#clean) flag in
`vagrant add`; defaults to unset.

`add_force` (bool) - Equivalent to setting the
[`--force`](https://www.vagrantup.com/docs/cli/box.html#force) flag in
`vagrant add`; defaults to unset.

`add_insecure` (bool) - Equivalent to setting the
[`--force`](https://www.vagrantup.com/docs/cli/box.html#insecure) flag in
`vagrant add`; defaults to unset.

`skip_package` (bool) - if true, Packer will not call `vagrant package` to
package your base box into its own standalone .box file.
