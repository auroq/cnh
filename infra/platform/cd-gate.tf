# The post-deploy gate's identity (roadmap item 036, ADR 0008): a ServiceAccount that can only read
# rollouts, replicasets and pods in the app namespace, a time-bound token for it, and the two GitHub
# Actions secrets CD needs. Every secret value comes from 1Password or the cluster and goes to GitHub
# and the vault without touching a terminal.

data "onepassword_item" "cd_github_token" {
  count = var.enable_cd_gate ? 1 : 0
  vault = data.onepassword_vault.cnh.uuid
  title = "CD GitHub Token CNH"
}

provider "github" {
  owner = split("/", var.github_repository)[0]
  token = var.enable_cd_gate ? data.onepassword_item.cd_github_token[0].credential : null
}

resource "kubernetes_service_account_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = "cd-gate"
    namespace = var.app_namespace
  }
  automount_service_account_token = false
}

resource "kubernetes_role_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = "cd-gate"
    namespace = var.app_namespace
  }

  rule {
    api_groups = ["argoproj.io"]
    resources  = ["rollouts"]
    verbs      = ["get", "list", "watch"]
  }
  rule {
    api_groups = ["apps"]
    resources  = ["replicasets"]
    verbs      = ["get", "list"]
  }
  rule {
    api_groups = [""]
    resources  = ["pods"]
    verbs      = ["get", "list"]
  }
}

resource "kubernetes_role_binding_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = "cd-gate"
    namespace = var.app_namespace
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role_v1.cd_gate[0].metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account_v1.cd_gate[0].metadata[0].name
    namespace = var.app_namespace
  }
}

resource "kubernetes_token_request_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = kubernetes_service_account_v1.cd_gate[0].metadata[0].name
    namespace = var.app_namespace
  }
  spec {
    expiration_seconds = var.cd_gate_token_ttl_seconds
  }
}

locals {
  cd_kubeconfig = var.enable_cd_gate ? yamlencode({
    apiVersion      = "v1"
    kind            = "Config"
    current-context = "cd-gate"
    clusters = [{
      name = var.cluster_name
      cluster = {
        server                     = local.k8s.host
        certificate-authority-data = base64encode(local.k8s.cluster_ca_certificate)
      }
    }]
    users = [{
      name = "cd-gate"
      user = { token = kubernetes_token_request_v1.cd_gate[0].token }
    }]
    contexts = [{
      name = "cd-gate"
      context = {
        cluster   = var.cluster_name
        user      = "cd-gate"
        namespace = var.app_namespace
      }
    }]
  }) : ""
}

resource "github_actions_secret" "cd_kubeconfig" {
  count = var.enable_cd_gate ? 1 : 0

  repository      = split("/", var.github_repository)[1]
  secret_name     = "CD_KUBECONFIG"
  plaintext_value = local.cd_kubeconfig
}

resource "github_actions_secret" "cd_github_token" {
  count = var.enable_cd_gate ? 1 : 0

  repository      = split("/", var.github_repository)[1]
  secret_name     = "CD_GITHUB_TOKEN"
  plaintext_value = data.onepassword_item.cd_github_token[0].credential
}

resource "onepassword_item" "cd_kubeconfig" {
  count = var.enable_cd_gate ? 1 : 0

  vault      = data.onepassword_vault.cnh.uuid
  title      = "CD Kubeconfig CNH"
  category   = "password"
  password   = local.cd_kubeconfig
  note_value = "Kubeconfig for the cd-gate ServiceAccount in ${var.app_namespace}: read rollouts, replicasets and pods only. Token expires ${var.cd_gate_token_ttl_seconds} seconds after the last apply. Managed by infra/platform (ADR 0008)."
}
