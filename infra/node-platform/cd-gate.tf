data "onepassword_item" "cd_github_token" {
  count = var.enable_cd_gate ? 1 : 0
  vault = data.onepassword_vault.cnh.uuid
  title = "CNH GitHub CD"
}

provider "github" {
  owner = split("/", var.github_repository)[0]
  token = var.enable_cd_gate ? data.onepassword_item.cd_github_token[0].credential : null
}

resource "kubernetes_service_account_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = "cd-gate"
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }
  automount_service_account_token = false
}

resource "kubernetes_role_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = "cd-gate"
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }

  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "replicasets"]
    verbs      = ["get", "list", "watch"]
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
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role_v1.cd_gate[0].metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account_v1.cd_gate[0].metadata[0].name
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }
}

resource "kubernetes_token_request_v1" "cd_gate" {
  count = var.enable_cd_gate ? 1 : 0

  metadata {
    name      = kubernetes_service_account_v1.cd_gate[0].metadata[0].name
    namespace = kubernetes_namespace_v1.app.metadata[0].name
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

resource "onepassword_item" "cd_kubeconfig" {
  count = var.enable_cd_gate ? 1 : 0

  vault      = data.onepassword_vault.cnh.uuid
  title      = "CD Kubeconfig node"
  category   = "password"
  password   = local.cd_kubeconfig
  note_value = "Kubeconfig for the cd-gate ServiceAccount in ${var.app_namespace} on ${var.cluster_name}: read deployments, replicasets and pods only. Token expires ${var.cd_gate_token_ttl_seconds} seconds after the last apply. Managed by infra/node-platform (ADR 0014)."
}
