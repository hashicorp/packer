---
name: Plugin Integration Request
about: Open request to add your plugin as a Packer integration (https://developer.hashicorp.com/packer/integrations)
labels: integration-request
---

#### Description

A written description of your plugin along with a link to the plugin repository. 

#### Integration Tier
<!--- By default all integrations are registered as community integrations.
HashiCorp Technology partners https://www.hashicorp.com/partners/find-a-partner will be registered as a partner once verified. --->

#### Checklist
- [ ] Has valid [`metadata.hcl`](https://github.com/hashicorp/integration-template) file in plugin repository.
- [ ] Has added integration scripts [packer-plugin-scaffolding](https://github.com/hashicorp/packer-plugin-scoffolding) to plugin repository.
- [ ] Has added top-level integration README.md file to plugin `docs` directory.
- [ ] All plugins components have one README.md describing their usage.
- [ ] Has a fully synced `.web-docs` directory ready for publishing to the integrations portal.

