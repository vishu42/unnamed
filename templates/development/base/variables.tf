variable "test" {
  default = "val"
  description = "some test var"
  type = string
  sensitive = true
}

variable "test_without_default" {
  description = "another test var"
  type = string
  sensitive = true
}