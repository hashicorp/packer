---
description: |
    The Packer Docker push post-processor takes an artifact from the docker-import
    post-processor and pushes it to a Docker registry.
layout: docs
page_title: 'Docker Push Post-Processor'
...

# Docker Push Post-Processor

Type: `docker-push`

The Packer Docker push post-processor takes an artifact from the
[docker-import](/docs/post-processors/docker-import.html) post-processor and
pushes it to a Docker registry.

## Configuration

This post-processor has only optional configuration:

-   `login` (boolean) - Defaults to false. If true, the post-processor will
    login prior to pushing.

-   `login_email` (string) - The email to use to authenticate to login.

-   `login_username` (string) - The username to use to authenticate to login.

-   `login_password` (string) - The password to use to authenticate to login.

-   `login_server` (string) - The server address to login to.

-&gt; **Note:** If you login using the credentials above, the post-processor
will automatically log you out afterwards (just the server specified).

## Example

For an example of using docker-push, see the section on using generated
artifacts from the [docker builder](/docs/builders/docker.html).
