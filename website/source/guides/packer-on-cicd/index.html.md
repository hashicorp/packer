---
layout: guides
sidebar_current: guides-packer-on-cicd-index
page_title: Building Immutable Infrastructure with Packer in CI/CD
---

# Building Immutable Infrastructure with Packer in CI/CD

This guide focuses on the following workflow for building immutable infrastructure. This workflow can be manual or automated and it can be implemented with a variety of technologies. The goal of this guide is to show how this workflow can be fully automated using Packer for building images from a CI/CD pipeline.

1. [Building Images using Packer in CI/CD](./building-image-in-cicd.html)
2. [Uploading the new image to S3](./uploading-images-to-artifact.html) for future deployment or use during development
3. Provision new instances with the images using Terraform Enterprise by [creating a new Terraform Enterprise runs](./triggering-tfe.html).