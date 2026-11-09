variable "region" {
  type    = string
  default = "sfo3"
}

variable "node_size" {
  type    = string
  default = "g-8vcpu-32gb"
}

variable "min_nodes" {
  type    = number
  default = 3
}

variable "max_nodes" {
  type    = number
  default = 5
}

variable "domain" {
  type    = string
  default = "cnh.auroq.com"
}
