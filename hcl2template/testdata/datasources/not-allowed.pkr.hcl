data "amazon-ami" "test_0" {
  string = "string"
}

data "amazon-ami" "test_1" {
  string = data.amazon-ami.test_0.string
}