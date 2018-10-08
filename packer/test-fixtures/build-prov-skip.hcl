builders = {
  type = "test"
}

builders = {
  name = "foo"

  type = "test"
}

provisioners = {
  only = ["foo"]

  type = "test"
}