
variable "fruit" {
  type = string
  default = "banana"
}

variable "unknown_string" {
  type = string
}


variable "unknown_list_of_string" {
  type = list(string)
}

variable "unknown_unknown" {
}
