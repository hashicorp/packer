---
layout: "docs"
---

# DigitalOcean Builder

Type: `digitalocean`

The `digitalocean` builder is able to create new images for use with
[DigitalOcean](http://www.digitalocean.com). The builder takes a source
image, runs any provisioning necessary on the image after launching it,
then snapshots it into a reusable image. This reusable image can then be
used as the foundation of new servers that are launched within DigitalOcean.

The builder does _not_ manage images. Once it creates an image, it is up to
you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

* `api_key` (string) - The API key to use to access your account. You can
  retrieve this on the "API" page visible after logging into your account
  on DigitalOcean. Alternatively, the builder looks for the environment
  variable `DIGITALOCEAN_API_KEY`.

* `client_id` (string) - The client ID to use to access your account. You can
  find this on the "API" page visible after logging into your account on
  DigitalOcean. Alternatively, the builder looks for the environment
  variable `DIGITALOCEAN_CLIENT_ID`.

Optional:

* `image_id` (int) - The ID of the base image to use. This is the image that
  will be used to launch a new droplet and provision it. Defaults to "1505447",
  which happens to be "Ubuntu 12.04.3 x64 Server."

* `region_id` (int) - The ID of the region to launch the droplet in. Consequently,
  this is the region where the snapshot will be available. This defaults to
  "1", which is "New York 1."

* `size_id` (int) - The ID of the droplet size to use. This defaults to "66",
  which is the 512MB droplet.

* `private_networking` (bool) - Set to `true` to enable private networking
  for the droplet being created. This defaults to `false`, or not enabled.

* `snapshot_name` (string) - The name of the resulting snapshot that will
  appear in your account. This must be unique.
  To help make this unique, use a function like `timestamp` (see
  [configuration templates](/docs/templates/configuration-templates.html) for more info)

* `droplet_name` (string) - The name assigned to the droplet. DigitalOcean
  sets the hostname of the machine to this value.

* `ssh_port` (int) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "5m". The default SSH timeout is "1m".

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running droplet. Default is "root".

* `state_timeout` (string) - The time to wait, as a duration string,
for a droplet to enter a desired state (such as "active") before
timing out. The default state timeout is "6m".

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your
own access tokens:

<pre class="prettyprint">
{
  "type": "digitalocean",
  "client_id": "YOUR CLIENT ID",
  "api_key": "YOUR API KEY"
}
</pre>

## Finding Image, Region, and Size IDs

Unfortunately, finding a list of available values for `image_id`, `region_id`,
and `size_id` is not easy at the moment. Basically, it has to be done through
the [DigitalOcean API](https://www.digitalocean.com/api_access) using the
`/images`, `/regions`, and `/sizes` endpoints. You can use `curl` for this
or request it in your browser.

If you're comfortable installing RubyGems, [Tugboat](https://github.com/pearkes/tugboat)
is a fantastic DigitalOcean command-line client that has commands to
find the available images, regions, and sizes. For example, to see all the
global images, you can run `tugboat images --global`.
