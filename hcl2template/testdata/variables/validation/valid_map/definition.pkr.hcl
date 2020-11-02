
variable "image_metadata" {
  default = {
    key: "value",
    something: { 
      foo: "bar",
    }
  }
  validation {
    condition     = length(var.image_metadata.key) > 4
    error_message = "The image_metadata.key field must be more than 4 runes."
  }
  validation {
    condition     = substr(var.image_metadata.something.foo, 0, 3) == "bar"
    error_message = "The image_metadata.something.foo field must start with \"bar\"."
  }
}
