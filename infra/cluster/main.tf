provider "onepassword" {}

data "onepassword_vault" "cnh" {
  name = "CNH"
}

data "onepassword_item" "do_token" {
  vault = data.onepassword_vault.cnh.uuid
  title = "Digital Ocean"
}

provider "digitalocean" {
  token = data.onepassword_item.do_token.credential
}

data "digitalocean_project" "cnh" {
  name = "cnh"
}

data "digitalocean_kubernetes_versions" "current" {}

resource "digitalocean_vpc" "cnh" {
  name     = "cnh"
  region   = var.region
  ip_range = "10.42.0.0/20"
}

resource "digitalocean_kubernetes_cluster" "cnh" {
  name          = "cnh"
  region        = var.region
  version       = data.digitalocean_kubernetes_versions.current.latest_version
  vpc_uuid      = digitalocean_vpc.cnh.id
  auto_upgrade  = false
  surge_upgrade = true

  node_pool {
    name       = "core"
    size       = var.node_size
    auto_scale = true
    min_nodes  = var.min_nodes
    max_nodes  = var.max_nodes
  }

  lifecycle {
    ignore_changes = [version]
  }
}

resource "digitalocean_domain" "cnh" {
  name = var.domain
}

resource "digitalocean_project_resources" "cnh" {
  project = data.digitalocean_project.cnh.id
  resources = [
    digitalocean_kubernetes_cluster.cnh.urn,
    digitalocean_domain.cnh.urn,
  ]
}
