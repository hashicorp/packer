
locals {
  timestamp = formatdate("YYYY-MM-DDX", timestamp())
  other_local = "test-${local.timestamp}"
}
