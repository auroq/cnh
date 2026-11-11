provider "onepassword" {}

data "onepassword_vault" "cnh" {
  name = "CNH"
}

data "onepassword_item" "es_elastic" {
  vault = data.onepassword_vault.cnh.uuid
  title = var.es_password_item_title
}

data "onepassword_item" "dash0" {
  vault = data.onepassword_vault.cnh.uuid
  title = var.dash0_item_title
}

locals {
  kubeconfig_file = abspath("${path.module}/${var.kubeconfig_path}")
  kubeconfig      = yamldecode(file(local.kubeconfig_file))
  k8s = {
    host                   = local.kubeconfig.clusters[0].cluster.server
    cluster_ca_certificate = base64decode(local.kubeconfig.clusters[0].cluster["certificate-authority-data"])
  }
}

provider "kubernetes" {
  config_path = local.kubeconfig_file
}

provider "helm" {
  kubernetes = {
    config_path = local.kubeconfig_file
  }
}
