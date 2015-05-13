---
layout: "intro"
page_title: "Packer and the HashiCorp Ecosystem"
prev_url: "/intro/platforms.html"
next_url: "/intro/getting-started/setup.html"
next_title: "Getting Started: Install Packer"
description: |-
  Learn how Packer fits in with the rest of the HashiCorp ecosystem of tools
---

# Packer and the HashiCorp Ecosystem

HashiCorp is the creator of the open source projects Vagrant, Packer, Terraform, Serf, and Consul, and the commercial product Atlas. Packer is just one piece of the ecosystem HashiCorp has built to make application delivery a versioned, auditable, repeatable, and collaborative process. To learn more about our beliefs on the qualities of the modern datacenter and responsible application delivery, read [The Atlas Mindset: Version Control for Infrastructure](https://hashicorp.com/blog/atlas-mindset.html/?utm_source=packer&utm_campaign=HashicorpEcosystem).

If you are using Packer to build machine images and deployable artifacts, it’s likely that you need a solution for deploying those artifacts. Terraform is our tool for creating, combining, and modifying infrastructure.

Below are summaries of HashiCorp’s open source projects and a graphic showing how Atlas connects them to create a full application delivery workflow.

# HashiCorp Ecosystem
![Atlas Workflow](docs/atlas-workflow.png)

[Atlas](https://atlas.hashicorp.com/?utm_source=packer&utm_campaign=HashicorpEcosystem) is HashiCorp's only commercial product. It unites Packer, Terraform, and Consul to make application delivery a versioned, auditable, repeatable, and collaborative process.

[Packer](https://packer.io/?utm_source=packer&utm_campaign=HashicorpEcosystem) is a HashiCorp tool for creating machine images and deployable artifacts such as AMIs, OpenStack images, Docker containers, etc. 

[Terraform](https://terraform.io/?utm_source=packer&utm_campaign=HashicorpEcosystem) is a HashiCorp tool for creating, combining, and modifying infrastructure. In the Atlas workflow Terraform reads from the artifact registry and provisions infrastructure. 

[Consul](https://consul.io/?utm_source=packer&utm_campaign=HashicorpEcosystem) is a HashiCorp tool for service discovery, service registry, and health checks. In the Atlas workflow Consul is configured at the Packer build stage and identifies the service(s) contained in each artifact. Since Consul is configured at the build phase with Packer, when the artifact is deployed with Terraform, it is fully configured with dependencies and service discovery pre-baked. This greatly reduces the risk of an unhealthy node in production due to configuration failure at runtime.

[Serf](https://serfdom.io/?utm_source=packer&utm_campaign=HashicorpEcosystem) is a HashiCorp tool for cluster membership and failure detection. Consul uses Serf’s gossip protocol as the foundation for service discovery.

[Vagrant](https://www.vagrantup.com/?utm_source=packer&utm_campaign=HashicorpEcosystem) is a HashiCorp tool for managing development environments that mirror production. Vagrant environments reduce the friction of developing a project and reduce the risk of unexpected behavior appearing after deployment. Vagrant boxes can be built in parallel with production artifacts with Packer to maintain parity between development and production.
