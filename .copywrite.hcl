project {
  license = "BUSL-1.1"
  copyright_year = 2024
  header_ignore = [
    "*.hcl2spec.go", # generated code specs, since they'll be wiped out until we support adding the headers at generation-time
    "hcl2template/testdata/**",
    "test/**",
    "**/test-fixtures/**",
    "examples/**",
    "hcl2template/fixtures/**",
    "command/plugin.go",
    "website/**" # candidates for copyright are coming from external sources, so we should not handle those in Packer
  ]
}
