---
layout: "docs"
page_title: "Docker Push Post-Processor"
---

# Docker Push Post-Processor

Type: `docker-push`

The Docker push post-processor takes an artifact from the
[docker-import](/docs/post-processors/docker-import.html) post-processor
and pushes it to a Docker registry.

<div class="alert alert-info alert-block">
<strong>Before you use this,</strong> you must manually <code>docker login</code>
to the proper repository. A future version of Packer will automate this
for you, but for now you must manually do this.
</div>

## Configuration

This post-processor has no configuration! Simply add it to your chain
of post-processors and the image will be uploaded.

## Example

For an example of using docker-push, see the section on using
generated artifacts from the [docker builder](/docs/builders/docker.html).
