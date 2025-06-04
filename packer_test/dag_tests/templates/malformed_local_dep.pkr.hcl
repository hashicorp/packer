data "http" "trusted_ca_certificates" {
  method = "GET"
  url    = local.no_dep
}
