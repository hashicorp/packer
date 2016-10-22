---
description: |
    The Packer Docker push post-processor takes an artifact from the docker-import
    post-processor and pushes it to a Docker registry.
layout: docs
page_title: 'Docker Push Post-Processor'
---

# Docker Push Post-Processor

Type: `docker-push`

The Packer Docker push post-processor takes an artifact from the
[docker-import](/docs/post-processors/docker-import.html) post-processor and
pushes it to a Docker registry.

## Configuration

This post-processor has only optional configuration:

-   `aws_access_key` (string) - The AWS access key used to communicate with AWS.
    [Learn how to
    set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `aws_secret_key` (string) - The AWS secret key used to communicate with AWS.
    [Learn how to
    set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `aws_token` (string) - The AWS access token to use. This is different from
    the access key and secret key. If you're not sure what this is, then you
    probably don't need it. This will also be read from the `AWS_SESSION_TOKEN`
    environmental variable.

-   `ecr_login` (boolean) - Defaults to false. If true, the post-processor will
    login in order to push the image to [Amazon EC2 Container
    Registry (ECR)](https://aws.amazon.com/ecr/). The post-processor only logs
    in for the duration of the push. If true `login_server` is required and
    `login`, `login_username`, and `login_password` will be ignored.

-   `login` (boolean) - Defaults to false. If true, the post-processor will
    login prior to pushing. For log into ECR see `ecr_login`.

-   `login_email` (string) - The email to use to authenticate to login.

-   `login_username` (string) - The username to use to authenticate to login.

-   `login_password` (string) - The password to use to authenticate to login.

-   `login_server` (string) - The server address to login to.

Note: When using *Docker Hub* or *Quay* registry servers, `login` must to be set
to `true` and `login_email`, `login_username`, **and** `login_password` must to
be set to your registry credentials. When using Docker Hub, `login_server` can
be omitted.

-&gt; **Note:** If you login using the credentials above, the post-processor
will automatically log you out afterwards (just the server specified).

## Example

For an example of using docker-push, see the section on using generated
artifacts from the [docker builder](/docs/builders/docker.html).
