package fix

const helpString = `
Usage: packer fix [options] TEMPLATE

  Reads the JSON template and attempts to fix known backwards
  incompatibilities. The fixed template will be outputted to standard out.

  If the template cannot be fixed due to an error, the command will exit
  with a non-zero exit status. Error messages will appear on standard error.

Fixes that are run:

  iso-md5             Replaces "iso_md5" in builders with newer "iso_checksum"
  createtime          Replaces ".CreateTime" in builder configs with "{{timestamp}}"
  virtualbox-gaattach Updates VirtualBox builders using "guest_additions_attach"
                      to use "guest_additions_mode"
  pp-vagrant-override Replaces old-style provider overrides for the Vagrant
                      post-processor to new-style as of Packer 0.5.0.
  virtualbox-rename   Updates "virtualbox" builders to "virtualbox-iso"

`
