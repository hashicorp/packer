---
layout: "docs"
---

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` builder is able to create
[images](https://developers.google.com/compute/docs/images)
for use with [Google Compute Engine](https://cloud.google.com/products/compute-engine)
(GCE) based on existing images. Google Compute Engine doesn't allow the creation
of images from scratch.

## Setting Up API Access

There is a small setup step required in order to obtain the credentials
that Packer needs to use Google Compute Engine. This needs to be done only
once if you intend to share the credentials.

In order for Packer to talk to Google Compute Engine, it will need
a _client secrets_ JSON file and a _client private key_. Both of these are
obtained from the [Google Cloud Console](https://cloud.google.com/console).

Follow the steps below:

1. Log into the [Google Cloud Console](https://cloud.google.com/console)
2. Click on the project you want to use Packer with (or create one if you
   don't have one yet).
3. Click "APIs & auth" in the left sidebar
4. Click "Credentials" in the left sidebar
5. Click "Create New Client ID" and choose "Service Account"
6. A private key will be downloaded for you. Note the password for the private key! This private key is your _client private key_.
7. After creating the account, click "Download JSON". This is your _client secrets JSON_ file. Make sure you didn't download the JSON from the "OAuth 2.0" section! This is a common mistake and will cause the builder to not work.

Finally, one last step, you'll have to convert the `p12` file you
got from Google into the PEM format. You can do this with OpenSSL, which
is installed standard on most Unixes:

```
$ openssl pkcs12 -in <path to .p12> -nocerts -passin pass:notasecret \
    -nodes -out private_key.pem
```

The client secrets JSON you downloaded along with the new "private\_key.pem"
file are the two files you need to configure Packer with to talk to GCE.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful,
since no provisioners are defined, but it will effectively repackage an
existing GCE image. The client secrets file and private key file are the
files obtained in the previous section.

<pre class="prettyprint">
{
  "type": "googlecompute",
  "bucket_name": "my-project-packer-images",
  "client_secrets_file": "client_secret.json",
  "private_key_file": "XXXXXX-privatekey.p12",
  "project_id": "my-project",
  "source_image": "debian-7-wheezy-v20131014",
  "zone": "us-central1-a"
}
</pre>

## Configuration Reference

Configuration options are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `bucket_name` (string) - The Google Cloud Storage bucket to store the
  images that are created. The bucket must already exist in your project.

* `client_secrets_file` (string) - The client secrets JSON file that
  was set up in the section above.

* `private_key_file` (string) - The client private key file that was
  generated in the section above.

* `project_id` (string) - The project ID that will be used to launch instances
  and store images.

* `source_image` (string) - The source image to use to create the new image
  from. Example: "debian-7"

* `zone` (string) - The zone in which to launch the instance used to create
  the image. Example: "us-central1-a"

### Optional:

* `image_name` (string) - The unique name of the resulting image.
  Defaults to `packer-{{timestamp}}`.

* `image_description` (string) - The description of the resulting image.

* `instance_name` (string) - A name to give the launched instance. Beware
  that this must be unique. Defaults to "packer-{{uuid}}".

* `machine_type` (string) - The machine type. Defaults to `n1-standard-1`.

* `metadata` (object of key/value strings)
<!---
@todo document me
-->

* `network` (string) - The Google Compute network to use for the launched
  instance. Defaults to `default`.

* `passphrase` (string) - The passphrase to use if the `private_key_file`
  is encrypted.

* `ssh_port` (integer) - The SSH port. Defaults to 22.

* `ssh_timeout` (string) - The time to wait for SSH to become available.
  Defaults to "1m".

* `ssh_username` (string) - The SSH username. Defaults to "root".

* `state_timeout` (string) - The time to wait for instance state changes.
  Defaults to "5m".

* `tags` (array of strings)
<!---
@todo document me
-->

## Gotchas

Centos images have root ssh access disabled by default. Set `ssh_username` to any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an `f1-micro` or `g1-small` to build images.
