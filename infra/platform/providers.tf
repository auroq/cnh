provider "onepassword" {}

data "onepassword_vault" "cnh" {
  name = "CNH"
}

data "onepassword_item" "do_token" {
  vault = data.onepassword_vault.cnh.uuid
  title = "Digital Ocean"
}

data "onepassword_item" "dash0" {
  count = var.enable_dash0 ? 1 : 0
  vault = data.onepassword_vault.cnh.uuid
  title = "Dash0 API Token CNH"
}

data "onepassword_item" "radar" {
  count = var.enable_radar ? 1 : 0
  vault = data.onepassword_vault.cnh.uuid
  title = "Radar API CNH"
}

provider "digitalocean" {
  token = data.onepassword_item.do_token.credential
}

data "digitalocean_kubernetes_cluster" "cnh" {
  name = var.cluster_name
}

locals {
  k8s = {
    host                   = data.digitalocean_kubernetes_cluster.cnh.endpoint
    token                  = data.digitalocean_kubernetes_cluster.cnh.kube_config[0].token
    cluster_ca_certificate = base64decode(data.digitalocean_kubernetes_cluster.cnh.kube_config[0].cluster_ca_certificate)
  }
}

provider "kubernetes" {
  host                   = local.k8s.host
  token                  = local.k8s.token
  cluster_ca_certificate = local.k8s.cluster_ca_certificate
}

provider "helm" {
  kubernetes = {
    host                   = local.k8s.host
    token                  = local.k8s.token
    cluster_ca_certificate = local.k8s.cluster_ca_certificate
  }
}
