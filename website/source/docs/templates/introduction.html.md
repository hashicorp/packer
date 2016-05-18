---
description: |
    Templates are JSON files that configure the various components of Packer in
    order to create one or more machine images. Templates are portable, static, and
    readable and writable by both humans and computers. This has the added benefit
    of being able to not only create and modify templates by hand, but also write
    scripts to dynamically create or modify templates.
layout: docs
page_title: Templates
...

# Templates

Templates are JSON files that configure the various components of Packer in
order to create one or more machine images. Templates are portable, static, and
readable and writable by both humans and computers. This has the added benefit
of being able to not only create and modify templates by hand, but also write
scripts to dynamically create or modify templates.

Templates are given to commands such as `packer build`, which will take the
template and actually run the builds within it, producing any resulting machine
images.

## Template Structure

A template is a JSON object that has a set of keys configuring various
components of Packer. The available keys within a template are listed below.
Along with each key, it is noted whether it is required or not.

-   `builders` (*required*) is an array of one or more objects that defines the
    builders that will be used to create machine images for this template, and
    configures each of those builders. For more information on how to define and
    configure a builder, read the sub-section on [configuring builders in
    templates](/docs/templates/builders.html).

-   `description` (optional) is a string providing a description of what the
    template does. This output is used only in the [inspect
    command](/docs/command-line/inspect.html).

-   `min_packer_version` (optional) is a string that has a minimum Packer
    version that is required to parse the template. This can be used to ensure
    that proper versions of Packer are used with the template. A max version
    can't be specified because Packer retains backwards compatibility with
    `packer fix`.

-   `post-processors` (optional) is an array of one or more objects that defines
    the various post-processing steps to take with the built images. If not
    specified, then no post-processing will be done. For more information on
    what post-processors do and how they're defined, read the sub-section on
    [configuring post-processors in
    templates](/docs/templates/post-processors.html).

-   `provisioners` (optional) is an array of one or more objects that defines
    the provisioners that will be used to install and configure software for the
    machines created by each of the builders. If it is not specified, then no
    provisioners will be run. For more information on how to define and
    configure a provisioner, read the sub-section on [configuring provisioners
    in templates](/docs/templates/provisioners.html).

-   `variables` (optional) is an object of one or more key/value strings that
    defines user variables contained in the template. If it is not specified,
    then no variables are defined. For more information on how to define and use
    user variables, read the sub-section on [user variables in
    templates](/docs/templates/user-variables.html).

## Comments

JSON doesn't support comments and Packer reports unknown keys as validation
errors. If you'd like to comment your template, you can prefix a *root level*
key with an underscore. Example:

``` {.javascript}
{
  "_comment": "This is a comment",
  "builders": [{}]
}
```

**Important:** Only *root level* keys can be underscore prefixed. Keys within
builders, provisioners, etc. will still result in validation errors.

## Example Template

Below is an example of a basic template that could be invoked with `packer build`. It would create an instance in AWS, and once running copy a script to it and run that script using SSH.

-&gt; **Note:** This example requires an account with Amazon Web Services. There are a number of parameters which need to be provided for a functional build to take place. See the [Amazon builder](/docs/builders/amazon.html) documentation for more information.

``` {.javascript}
{
  "builders": [
    {
      "type": "amazon-ebs",
      "access_key": "...",
      "secret_key": "...",
      "region": "us-east-1",
      "source_ami": "ami-fce3c696",
      "instance_type": "t2.micro",
      "ssh_username": "ubuntu",
      "ami_name": "packer {{timestamp}}"
    }
  ],

  "provisioners": [
    {
      "type": "shell",
      "script": "setup_things.sh"
    }
  ]
}
```
