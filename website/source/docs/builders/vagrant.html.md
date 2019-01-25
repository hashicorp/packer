The Vagrant builder is intended for building new boxes from already-existing
boxes. Your source should be a URL or path to a .box file or a Vagrant Cloud
box name such as `hashicorp/precise64`.

Packer will not install vagrant, nor will it install the underlying
virtualization platforms or extra providers; We expect when you run this
builder that you have already installed what you need.

By default, this builder will initialize a new Vagrant workspace, launch your
box from that workspace, provision it, call `vagrant package` to package it
into a new box, and then destroy the original box. Please note that vagrant
will _not_ remove the box file from your system (we don't call
`vagrant box remove`).

You can change the behavior so that the builder doesn't destroy the box by
setting the `teardown_method` option. You can change the behavior so the builder
doesn't package it (not all provisioners support the `vagrant package` command)
by setting the `skip package` option. You can also change the behavior so that
rather than inititalizing a new Vagrant workspace, you use an already defined
one, by using `global_id` instead of `source_box`.

Required:

-    `source_box` (string) - URL of the vagrant box to use, or the name of the
    vagrant box. `hashicorp/precise64`, `./mylocalbox.box` and
    `https://example.com/my-box.box` are all valid source boxes. If your
    source is a .box file, whether locally or from a URL like the latter example
    above, you will also need to provide a `box_name`. This option is required,
    unless you set `global_id`. You may only set one or the other, not both.

    or

-  `global_id` (string) - the global id of a Vagrant box already added to Vagrant
   on your system. You can find the global id of your Vagrant boxes using the
   command `vagrant global-status`; your global_id will be a 7-digit number and
   letter comination that you'll find in the leftmost column of the
   global-status output.  If you choose to use `global_id` instead of
   `source_box`, Packer will skip the Vagrant initialize and add steps, and
   simply launch the box directly using the global id.

Optional:

-   `output_dir` (string) - The directory to create that will contain
    your output box. We always create this directory and run from inside of it to
    prevent Vagrant init collisions. If unset, it will be set to `packer-` plus
    your buildname.

-   `box_name` (string) - if your source\_box is a boxfile that we need to add
    to Vagrant, this is the name to give it. If left blank, will default to
    "packer_" plus your buildname.

-   `checksum` (string) - The checksum for the .box file. The type of the
    checksum is specified with `checksum_type`, documented below.

-   `checksum_type` (string) - The type of the checksum specified in `checksum`.
    Valid values are `none`, `md5`, `sha1`, `sha256`, or `sha512`. Although the
    checksum will not be verified when `checksum_type` is set to "none", this is
    not recommended since OVA files can be very large and corruption does happen
    from time to time.

-   `vagrantfile_template` (string) - a path to an ERB template to use for the
    vagrantfile when calling `vagrant init`. See the blog post
    [here](https://www.hashicorp.com/blog/hashicorp-vagrant-2-0-2#customized-vagrantfile-templates)
    for some more details on how this works. Available variables are `box_name`,
    `box_url`, and `box_version`.

-   `skip_add` (string) - Don't call "vagrant add" to add the box to your local
    environment; this is necesasry if you want to launch a box that is already
    added to your vagrant environment.

-   `teardown_method` (string) - Whether to halt, suspend, or destroy the box when
    the build has completed. Defaults to "halt"

-   `box_version` (string) - What box version to use when initializing Vagrant.

-   `init_minimal` (bool) - If true, will add the --minimal flag to the Vagrant
    init command, creating a minimal vagrantfile instead of one filled with helpful
    comments.

-   `add_cacert` (string) - Equivalent to setting the
    [`--cacert`](https://www.vagrantup.com/docs/cli/box.html#cacert-certfile)
    option in `vagrant add`; defaults to unset.

-   `add_capath` (string) - Equivalent to setting the
    [`--capath`](https://www.vagrantup.com/docs/cli/box.html#capath-certdir) option
    in `vagrant add`; defaults to unset.

-   `add_cert` (string) - Equivalent to setting the
    [`--cert`](https://www.vagrantup.com/docs/cli/box.html#cert-certfile) option in
    `vagrant add`; defaults to unset.

-   `add_clean` (bool) - Equivalent to setting the
    [`--clean`](https://www.vagrantup.com/docs/cli/box.html#clean) flag in
    `vagrant add`; defaults to unset.

-   `add_force` (bool) - Equivalent to setting the
    [`--force`](https://www.vagrantup.com/docs/cli/box.html#force) flag in
    `vagrant add`; defaults to unset.

-   `add_insecure` (bool) - Equivalent to setting the
    [`--force`](https://www.vagrantup.com/docs/cli/box.html#insecure) flag in
    `vagrant add`; defaults to unset.

-   `skip_package` (bool) - if true, Packer will not call `vagrant package` to
    package your base box into its own standalone .box file.
