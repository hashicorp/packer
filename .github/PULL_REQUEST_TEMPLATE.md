**DELETE THIS PART BEFORE SUBMITTING**

In order to have a good experience with our community, we recommend that you
read the contributing guidelines for making a PR, and understand the lifecycle
of a Packer Plugin PR:
- https://github.com/hashicorp/$REPO_NAME/blob/main/.github/CONTRIBUTING.md#opening-an-pull-request

Please include tests. Check out these examples:

- https://github.com/hashicorp/packer/blob/master/builder/parallels/common/ssh_config_test.go#L34
- https://github.com/hashicorp/packer/blob/master/post-processor/compress/post-processor_test.go#L153-L182

----

### Description
What code changed, and why?

### Resolved Issues
If your PR resolves any open issue(s), please indicate them like this so they 
will be closed when your PR is merged:
Closes #xxx
Closes #xxx

<!-- heimdall_github_prtemplate:grc-pci_dss-2024-01-05 -->
### Rollback Plan
If a change needs to be reverted, we will roll out an update to the code within 
7 days.

### Changes to Security Controls
Are there any changes to security controls (access controls, encryption, logging) 
in this pull request? If so, explain.
