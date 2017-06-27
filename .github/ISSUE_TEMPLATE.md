_Read these instructions before submitting_

**DELETE THIS TEMPLATE BEFORE SUBMITTING**

_Only use Github issues to report bugs or feature requests, see
https://www.packer.io/community.html_

Specifically we _don't accept_ issues for:
- _Timeout waiting for SSH/WinRM_ - Unless there is evidence pointing towards a bug in packer. e.g it works with a previous version, or you can manually connect with the same credentials, etc. Ask on the mailing list if you are unsure.

If you are planning to open a pull-request just open the pull-request instead of making an issue first.

FOR FEATURES:

Describe the feature you want and your use case _clearly_.

FOR BUGS:

Describe the problem and include the following information:

- Packer version from `packer version`
- Host platform
- Debug log output from `PACKER_LOG=1 packer build template.json`.
  Please paste this in a gist https://gist.github.com
- The _simplest example template and scripts_ needed to reproduce the bug.
  Include these in your gist https://gist.github.com
