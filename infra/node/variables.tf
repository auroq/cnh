variable "region" {
  type    = string
  default = "sfo3"
}

variable "name" {
  type    = string
  default = "wide-door-node"
}

variable "size" {
  type        = string
  default     = "s-4vcpu-16gb-amd"
  description = "Droplet size; 16 GB fits MicroK8s, Elasticsearch with ELSER, Kibana and Argo CD with room to spare"
}

variable "image" {
  type    = string
  default = "ubuntu-24-04-x64"
}

variable "microk8s_channel" {
  type        = string
  default     = "1.36/stable"
  description = "Snap channel for microk8s (see https://snapcraft.io/microk8s)"
}

variable "ssh_public_key_path" {
  type        = string
  default     = "~/.ssh/id_ed25519.pub"
  description = "Public key registered with DigitalOcean and installed for root on the droplet"
}

variable "vpc_name" {
  type    = string
  default = "cnh"
}

variable "domain" {
  type    = string
  default = "cnh.auroq.com"
}

variable "hostname" {
  type        = string
  default     = "widedoor"
  description = "Record under var.domain that points at the droplet once publish_dns is true"
}

variable "publish_dns" {
  type        = bool
  default     = false
  description = "Create the A record for var.hostname. Leave false until the node answers over http; the specific record overrides the cluster's wildcard at once."
}

variable "es_password_item_title" {
  type    = string
  default = "Wide Door node elastic"
}
