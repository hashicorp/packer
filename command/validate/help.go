package validate

const helpString = `
Usage: packer validate [options] TEMPLATE

  Checks the template is valid by parsing the template and also
  checking the configuration with the various builders, provisioners, etc.

  If it is not valid, the errors will be shown and the command will exit
  with a non-zero exit status. If it is valid, it will exit with a zero
  exit status.

Options:

  -syntax-only           Only check syntax. Do not verify config of the template.
  -except=foo,bar,baz    Validate all builds other than these
  -only=foo,bar,baz      Validate only these builds
  -var 'key=value'       Variable for templates, can be used multiple times.
  -var-file=path         JSON file containing user variables.
`
