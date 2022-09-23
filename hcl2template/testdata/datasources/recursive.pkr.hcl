data "null" "foo" {
  input = "chocolate"
}

data "null" "yummy" {
  input = "${data.null.bang.output}-and-sprinkles"
}

data "null" "bar" {
  input = "vanilla"
}

data "null" "baz" {
  input = "${data.null.foo.output}-${data.null.bar.output}-swirl"
}

data "null" "bang" {
  input = "${data.null.baz.output}-with-marshmallows"
}

source "null" "test" {
  communicator = "none"
}

build {
  sources = ["null.test"]
}
