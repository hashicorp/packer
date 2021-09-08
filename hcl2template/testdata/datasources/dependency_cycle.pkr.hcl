data "null" "gummy" {
  input = "${data.null.bear.output}"
}
data "null" "bear" {
  input = "${data.null.gummy.output}"
}