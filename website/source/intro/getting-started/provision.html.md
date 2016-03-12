---
description: |
    In the previous page of this guide, you created your first image with Packer.
    The image you just built, however, was basically just a repackaging of a
    previously existing base AMI. The real utility of Packer comes from being able
    to install and configure software into the images as well. This stage is also
    known as the *provision* step. Packer fully supports automated provisioning in
    order to install software onto the machines prior to turning them into images.
layout: intro
next_title: Parallel Builds
next_url: '/intro/getting-started/parallel-builds.html'
page_title: Provision
prev_url: '/intro/getting-started/build-image.html'
...

# Provision

In the previous page of this guide, you created your first image with Packer.
The image you just built, however, was basically just a repackaging of a
previously existing base AMI. The real utility of Packer comes from being able
to install and configure software into the images as well. This stage is also
known as the *provision* step. Packer fully supports automated provisioning in
order to install software onto the machines prior to turning them into images.

In this section, we're going to complete our image by installing Redis on it.
This way, the image we end up building actually contains Redis pre-installed.
Although Redis is a small, simple example, this should give you an idea of what
it may be like to install many more packages into the image.

Historically, pre-baked images have been frowned upon because changing them has
been so tedious and slow. Because Packer is completely automated, including
provisioning, images can be changed quickly and integrated with modern
configuration management tools such as Chef or Puppet.

## Configuring Provisioners

Provisioners are configured as part of the template. We'll use the built-in
shell provisioner that comes with Packer to install Redis. Modify the
`example.json` template we made previously and add the following. We'll explain
the various parts of the new configuration following the code block below.

``` {.javascript}
{
  "variables": ["..."],
  "builders": ["..."],

  "provisioners": [{
    "type": "shell",
    "inline": [
      "sleep 30",
      "sudo apt-get update",
      "sudo apt-get install -y redis-server"
    ]
  }]
}
```

-&gt; **Note:** The `sleep 30` in the example above is very important. Because
Packer is able to detect and SSH into the instance as soon as SSH is available,
Ubuntu actually doesn't get proper amounts of time to initialize. The sleep
makes sure that the OS properly initializes.

Hopefully it is obvious, but the `builders` section shouldn't actually contain
"...", it should be the contents setup in the previous page of the getting
started guide. Also note the comma after the `"builders": [...]` section, which
was not present in the previous lesson.

To configure the provisioners, we add a new section `provisioners` to the
template, alongside the `builders` configuration. The provisioners section is an
array of provisioners to run. If multiple provisioners are specified, they are
run in the order given.

By default, each provisioner is run for every builder defined. So if we had two
builders defined in our template, such as both Amazon and DigitalOcean, then the
shell script would run as part of both builds. There are ways to restrict
provisioners to certain builds, but it is outside the scope of this getting
started guide. It is covered in more detail in the complete
[documentation](/docs).

The one provisioner we defined has a type of `shell`. This provisioner ships
with Packer and runs shell scripts on the running machine. In our case, we
specify two inline commands to run in order to install Redis.

## Build

With the provisioner configured, give it a pass once again through
`packer validate` to verify everything is okay, then build it using
`packer build example.json`. The output should look similar to when you built
your first image, except this time there will be a new step where the
provisioning is run.

The output from the provisioner is too verbose to include in this guide, since
it contains all the output from the shell scripts. But you should see Redis
successfully install. After that, Packer once again turns the machine into an
AMI.

If you were to launch this AMI, Redis would be pre-installed. Cool!

This is just a basic example. In a real world use case, you may be provisioning
an image with the entire stack necessary to run your application. Or maybe just
the web stack so that you can have an image for web servers pre-built. This
saves tons of time later as you launch these images since everything is
pre-installed. Additionally, since everything is pre-installed, you can test the
images as they're built and know that when they go into production, they'll be
functional.
