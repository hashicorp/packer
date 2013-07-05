---
layout: "docs"
---

# Templates

Templates are JSON files that configure the various components of Packer
in order to create one or more machine images. Templates are portable, static,
and readable and writable by both humans and computers. This has the added
benefit of being able to not only create and modify templates by hand, but
also write scripts to dynamically create or modify templates.

Templates are given to commands such as `packer build`, which will
take the template and actually run the builds within it, producing
any resulting machine images.

## Template Structure

A template is a JSON object that has a set of keys configuring various
components of Packer. The available keys within a template are listed below.
Along with each key, it is noted whether it is required or not.

* `builders` (_required_) is an array of one or more objects that defines
  the builders that will be used to create machine images for this template,
  and configures each of those builders. For more information on how to define
  and configure a builder, read the sub-section on
  [configuring builders in templates](/docs/templates/builders.html).

* `provisioners` (optional) is an array of one or more objects that defines
  the provisioners that will be used to install and configure software for
  the machines created by each of the builders. This is an optional field.
  If it is not specified, then no provisioners will be run. For more
  information on how to define and configure a provisioner, read the
  sub-section on [configuring provisioners in templates](/docs/templates/provisioners.html).

* `post-processors` (optional) is an array of that defines the various
  post-processing steps to take with the built images. This is an optional
  field. If not specified, then no post-processing will be done. For more
  information on what post-processors do and how they're defined, read the
  sub-section on [configuring post-processors in templates](/docs/templates/post-processors.html).

## Example Template

Below is an example of a basic template that is nearly fully functional. It is just
missing valid AWS access keys. Otherwise, it would work properly with
`packer build`.

<pre class="prettyprint">
{
  "builders": [
    {
      "type": "amazon-ebs",
      "access_key": "...",
      "secret_key": "...",
      "region": "us-east-1",
      "source_ami": "ami-de0d9eb7",
      "instance_type": "t1.micro",
      "ssh_username": "ubuntu",
      "ami_name": "packer {{.CreateTime}}"
    }
  ],

  "provisioners": [
    {
      "type": "shell",
      "script": "setup_things.sh"
    }
  ]
}
</pre>
