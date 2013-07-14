package fix

const helpString = `
Usage: packer fix [options] TEMPLATE

  Reads the JSON template and attempts to fix known backwards
  incompatibilities. The fixed template will be outputted to standard out.

  If the template cannot be fixed due to an error, the command will exit
  with a non-zero exit status. Error messages will appear on standard error.

Options:

  -verbose        Output each fix to standard error
`
