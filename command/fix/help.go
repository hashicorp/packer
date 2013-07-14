package fix

const helpString = `
Usage: packer fix [options] TEMPLATE

  Reads the JSON template and attempts to fix known backwards
  incompatibilities. The fixed template will be outputted to standard out.

  If the template cannot be fixed due to an error, the command will exit
  with a non-zero exit status. Error messages will appear on standard error.

Fixes that are run:

  iso-md5         Replaces "iso_md5" in builders with newer "iso_checksum"

`
