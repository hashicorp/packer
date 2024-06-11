locals {
  test_local = can(local.test_data) ? local.test_data : []

  test_data = [
    { key = "value" }
  ]
}
