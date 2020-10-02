
variable "not_sensitive" {
  default   = "I am soooo not sensitive"
}

variable "not_sensitive_unknown" {
}

variable "sensitive" {
  default   = "I am soooo sensitive"
  sensitive = true
}

variable "sensitive_array" {
  default   = ["Im supersensitive", "me too !!!!"]
  sensitive = true
}

variable "sensitive_tags" {
  default   = {
      first_key  = "this-is-mega-sensitive"
      second_key = "this-is-also-sensitive"
  }
  sensitive = true
}

variable "sensitive_unknown" {
  sensitive = true
}
