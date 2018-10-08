variables = {
  foo = ""
}

builders = {
  type = "test"
}

push = {
  name = "{{user `foo`}}"
}

