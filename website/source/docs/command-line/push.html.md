---
description: |
    The `packer push` command uploads a template and other required files to the
    Atlas build service, which will run your packer build for you.
layout: docs
page_title: 'Push - Command-Line'
...

# Command-Line: Push

The `packer push` command uploads a template and other required files to the
Atlas service, which will run your packer build for you. [Learn more about
Packer in Atlas.](https://atlas.hashicorp.com/help/packer/features)

Running builds remotely makes it easier to iterate on packer builds that are not
supported on your operating system, for example, building docker or QEMU while
developing on Mac or Windows. Also, the hard work of building VMs is offloaded
to dedicated servers with more CPU, memory, and network resources.

When you use push to run a build in Atlas, you may also want to store your build
artifacts in Atlas. In order to do that you will also need to configure the
[Atlas post-processor](/docs/post-processors/atlas.html). This is optional, and
both the post-processor and push commands can be used independently.

!&gt; The push command uploads your template and other files, like provisioning
scripts, to Atlas. Take care not to upload files that you don't intend to, like
secrets or large binaries. **If you have secrets in your Packer template, you
should [move them into environment
variables](https://www.packer.io/docs/templates/user-variables.html).**

Most push behavior is [configured in your packer
template](/docs/templates/push.html). You can override or supplement your
configuration using the options below.

## Options

-   `-token` - Your access token for the Atlas API.

-&gt; Login to Atlas to [generate an Atlas
Token](https://atlas.hashicorp.com/settings/tokens). The most convenient way to
configure your token is to set it to the `ATLAS_TOKEN` environment variable, but
you can also use `-token` on the command line.

-   `-name` - The name of the build in the service. This typically looks like
    `hashicorp/precise64`, which follows the form `<username>/<buildname>`. This
    must be specified here or in your template.

-   `-var` - Set a variable in your packer template. This option can be used
    multiple times. This is useful for setting version numbers for your build.

-   `-var-file` - Set template variables from a file.

## Environment Variables

-   `ATLAS_CAFILE` (path) - This should be a path to an X.509 PEM-encoded public key. If specified, this will be used to validate the certificate authority that signed certificates used by an Atlas installation.

-   `ATLAS_CAPATH` - This should be a path which contains an X.509 PEM-encoded public key file. If specified, this will be used to validate the certificate authority that signed certificates used by an Atlas installation.

## Examples

Push a Packer template:

``` {.shell}
$ packer push template.json
```

Push a Packer template with a custom token:

``` {.shell}
$ packer push -token ABCD1234 template.json
```

## Limits

`push` is limited to 5gb upload when pushing to Atlas. To be clear, packer *can*
build artifacts larger than 5gb, and Atlas *can* store artifacts larger than
5gb. However, the initial payload you push to *start* the build cannot exceed
5gb. If your boot ISO is larger than 5gb (for example if you are building OSX
images), you will need to put your boot ISO in an external web service and
download it during the packer run.

## Building Private `.iso` and `.dmg` Files

If you want to build a private `.iso` file you can upload the `.iso` to a secure
file hosting service like [Amazon
S3](https://docs.aws.amazon.com/AmazonS3/latest/dev/ShareObjectPreSignedURL.html),
[Google Cloud
Storage](https://cloud.google.com/storage/docs/gsutil/commands/signurl), or
[Azure File
Service](https://msdn.microsoft.com/en-us/library/azure/dn194274.aspx) and
download it at build time using a signed URL. You should convert `.dmg` files to
`.iso` and follow a similar procedure.

Once you have added [variables in your packer
template](/docs/templates/user-variables.html) you can specify credentials or
signed URLs using Atlas environment variables, or via the `-var` flag when you
run `push`.

![Configure your signed URL in the Atlas build variables
menu](/assets/images/packer-signed-urls.png)
