builders = {
  type = "test"

  value = "{{build_name}}"
}

sensitive-variables = ["foo"]

variables = {
  foo = "bar"
}