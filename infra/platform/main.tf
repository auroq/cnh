locals {
  namespaces = toset(["envoy-gateway-system", "gateway", "cert-manager", "elastic-system", "argocd", "dash0-system", "radar"])
}

resource "kubernetes_namespace_v1" "ns" {
  for_each = local.namespaces

  metadata {
    name = each.key
  }
}

resource "kubernetes_secret_v1" "do_token" {
  for_each = toset(["cert-manager"])

  metadata {
    name      = "digitalocean-token"
    namespace = kubernetes_namespace_v1.ns[each.key].metadata[0].name
  }

  data = {
    token = data.onepassword_item.do_token.credential
  }
}

resource "helm_release" "envoy_gateway" {
  name       = "envoy-gateway"
  repository = "oci://docker.io/envoyproxy"
  chart      = "gateway-helm"
  version    = var.versions.envoy_gateway
  namespace  = kubernetes_namespace_v1.ns["envoy-gateway-system"].metadata[0].name
  wait       = true
}

resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = var.versions.cert_manager
  namespace  = kubernetes_namespace_v1.ns["cert-manager"].metadata[0].name
  wait       = true

  values = [yamlencode({
    crds = { enabled = true }
    config = {
      apiVersion       = "controller.config.cert-manager.io/v1alpha1"
      kind             = "ControllerConfiguration"
      enableGatewayAPI = true
    }
  })]

  depends_on = [helm_release.envoy_gateway]
}

resource "helm_release" "platform_config" {
  name      = "platform-config"
  chart     = "${path.module}/charts/platform-config"
  namespace = kubernetes_namespace_v1.ns["gateway"].metadata[0].name

  values = [yamlencode({
    domain    = var.domain
    acmeEmail = var.acme_email
    doTokenSecret = {
      name      = kubernetes_secret_v1.do_token["cert-manager"].metadata[0].name
      namespace = "cert-manager"
    }
  })]

  depends_on = [helm_release.cert_manager, helm_release.envoy_gateway, helm_release.argo_cd]
}

resource "helm_release" "eck_operator" {
  name       = "elastic-operator"
  repository = "https://helm.elastic.co"
  chart      = "eck-operator"
  version    = var.versions.eck_operator
  namespace  = kubernetes_namespace_v1.ns["elastic-system"].metadata[0].name
}

resource "helm_release" "argo_cd" {
  name       = "argo-cd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = var.versions.argo_cd
  namespace  = kubernetes_namespace_v1.ns["argocd"].metadata[0].name

  values = [yamlencode({
    global = { domain = "argocd.${var.domain}" }
    configs = {
      params = { "server.insecure" = true }
    }
  })]
}

resource "kubernetes_secret_v1" "dash0" {
  count = var.enable_dash0 ? 1 : 0

  metadata {
    name      = "dash0-authorization-secret"
    namespace = kubernetes_namespace_v1.ns["dash0-system"].metadata[0].name
  }

  data = {
    token = data.onepassword_item.dash0[0].credential
  }
}

resource "helm_release" "dash0_operator" {
  count = var.enable_dash0 ? 1 : 0

  name       = "dash0-operator"
  repository = "https://dash0hq.github.io/dash0-operator"
  chart      = "dash0-operator"
  version    = var.versions.dash0
  namespace  = kubernetes_namespace_v1.ns["dash0-system"].metadata[0].name

  values = [yamlencode({
    operator = {
      clusterName = var.cluster_name
      dash0Export = {
        enabled     = true
        endpoint    = var.dash0_otlp_endpoint
        apiEndpoint = var.dash0_api_endpoint
        secretRef = {
          name = kubernetes_secret_v1.dash0[0].metadata[0].name
          key  = "token"
        }
      }
    }
  })]
}

resource "kubernetes_secret_v1" "radar" {
  count = var.enable_radar ? 1 : 0

  metadata {
    name      = "radar-cloud-config"
    namespace = kubernetes_namespace_v1.ns["radar"].metadata[0].name
  }

  data = {
    token = data.onepassword_item.radar[0].credential
  }
}

resource "helm_release" "radar" {
  count = var.enable_radar ? 1 : 0

  name       = "radar"
  repository = "https://skyhook-io.github.io/helm-charts"
  chart      = "radar"
  namespace  = kubernetes_namespace_v1.ns["radar"].metadata[0].name

  values = [yamlencode({
    cloud = {
      enabled            = true
      url                = "wss://api.radarhq.io/agent"
      insecureSkipVerify = false
      clusterName        = var.radar_cluster_id
      existingSecret     = kubernetes_secret_v1.radar[0].metadata[0].name
    }
    auth = { mode = "proxy" }
    rbac = {
      helm        = true
      secrets     = true
      podExec     = true
      portForward = true
      metrics     = true
      selfUpgrade = true
    }
  })]
}

data "kubernetes_resource" "public_gateway" {
  api_version = "gateway.networking.k8s.io/v1"
  kind        = "Gateway"

  metadata {
    name      = "public"
    namespace = kubernetes_namespace_v1.ns["gateway"].metadata[0].name
  }

  depends_on = [helm_release.platform_config]
}

resource "digitalocean_record" "wildcard" {
  domain = var.domain
  type   = "A"
  name   = "*"
  value  = data.kubernetes_resource.public_gateway.object.status.addresses[0].value
  ttl    = 300
}

resource "kubernetes_namespace_v1" "otel_demo" {
  count = var.enable_sandbox ? 1 : 0

  metadata {
    name = "otel-demo"
  }
}

resource "kubernetes_secret_v1" "otel_demo_dash0" {
  count = var.enable_sandbox && var.enable_dash0 ? 1 : 0

  metadata {
    name      = "dash0-export"
    namespace = kubernetes_namespace_v1.otel_demo[0].metadata[0].name
  }

  data = {
    token = data.onepassword_item.dash0[0].credential
  }
}
