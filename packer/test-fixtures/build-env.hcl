builders = {
  type = "test"

  value = "{{user `var`}}"
}

variables = {
  var = "{{env `PACKER_TEST_ENV`}}"
}