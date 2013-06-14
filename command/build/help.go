package build

const helpText = `
Usage: packer build TEMPLATE

  Will execute multiple builds in parallel as defined in the template.
  The various artifacts created by the template will be outputted.

Options:

  -debug                     Debug mode enabled for builds
  -except=foo,bar,baz        Build all builds other than these
  -only=foo,bar,baz          Only build the given builds by name
`
