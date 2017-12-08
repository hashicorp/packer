---
layout: guides
sidebar_current: guides-packer-on-cicd-upload-image-to-artifact-store
page_title: Uploading Images to Artifact Stores
description: |-
  ...
---

# Uploading Images to Artifact Stores

Once the image is generated it will be used by other parts of your operations workflow. For example, it is common to build VirtualBoxes with Packer to be used as base boxes in Vagrant.

On the agent machine install the [AWS Command Line Tool](https://aws.amazon.com/cli/). Since this is a one-time operation, this can be incorporated into the initial provisioning step when installing other dependencies.

```shell
pip install awscli
```

Add an additional **Build Step: Command Line** to the build and set the **Script content** field to the following:

```shell
awscli s3 cp . s3://bucket/ --exclude “*” --include “*.iso”
```

TeamCity provides a [Build Artifacts](https://confluence.jetbrains.com/display/TCD9/Build+Artifact) feature which can be used to store the newly generated image. Other CI/CD services also have similar build artifacts features built in, like [Circle CI Build Artifacts](https://circleci.com/docs/2.0/artifacts/). In addition to the built in artifact stores in CI/CD tools, there are also dedicated universal artifact storage services like [Artifactory](https://confluence.jetbrains.com/display/TCD9/Build+Artifact). All of these are great options for image artifact storage.
