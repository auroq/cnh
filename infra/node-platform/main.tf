locals {
  res_small = {
    requests = { cpu = "50m", memory = "64Mi" }
    limits   = { cpu = "500m", memory = "256Mi" }
  }
  res_medium = {
    requests = { cpu = "100m", memory = "128Mi" }
    limits   = { cpu = "1", memory = "512Mi" }
  }
  res_large = {
    requests = { cpu = "250m", memory = "256Mi" }
    limits   = { cpu = "2", memory = "1Gi" }
  }
}

resource "kubernetes_namespace_v1" "app" {
  metadata {
    name = var.app_namespace
  }
}

resource "kubernetes_namespace_v1" "argocd" {
  metadata {
    name = "argocd"
  }
}

resource "kubernetes_secret_v1" "es_elastic_user" {
  metadata {
    name      = "wide-door-es-elastic-user"
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }

  data = {
    elastic = data.onepassword_item.es_elastic.password
  }
}

resource "kubernetes_secret_v1" "dash0_otlp" {
  metadata {
    name      = "dash0-otlp"
    namespace = kubernetes_namespace_v1.app.metadata[0].name
  }

  data = {
    headers = "Authorization=Bearer ${data.onepassword_item.dash0.credential}"
  }
}

resource "helm_release" "argo_cd" {
  name       = "argo-cd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = var.argo_cd_version
  namespace  = kubernetes_namespace_v1.argocd.metadata[0].name
  wait       = true
  timeout    = 900

  values = [yamlencode({
    configs = {
      params = { "server.insecure" = true }
    }
    controller      = { replicas = 1, resources = local.res_large }
    server          = { replicas = 1, resources = local.res_small }
    repoServer      = { replicas = 1, resources = local.res_medium }
    redis           = { resources = local.res_small }
    redisSecretInit = { resources = local.res_small }
    applicationSet  = { enabled = false }
    dex             = { enabled = false }
    notifications   = { enabled = false }
  })]
}

resource "helm_release" "node_config" {
  name      = "node-config"
  chart     = "${path.module}/charts/node-config"
  namespace = kubernetes_namespace_v1.argocd.metadata[0].name

  values = [yamlencode({
    application = {
      name           = "wide-door"
      repoURL        = var.application.repo_url
      targetRevision = var.application.target_revision
      path           = var.application.path
      namespace      = var.app_namespace
    }
  })]

  depends_on = [helm_release.argo_cd, kubernetes_secret_v1.es_elastic_user, kubernetes_secret_v1.dash0_otlp]
}
