
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

variable "default_from_env" {
  default = env("DEFAULT_FROM_ENV")
}

variable "other_default_from_env" {
  default = env("OTHER_DEFAULT_FROM_ENV")
}
