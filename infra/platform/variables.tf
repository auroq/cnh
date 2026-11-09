variable "cluster_name" {
  type    = string
  default = "cnh"
}

variable "domain" {
  type    = string
  default = "cnh.auroq.com"
}

variable "acme_email" {
  type    = string
  default = "dev@ariavesta.com"
}

variable "enable_dash0" {
  type    = bool
  default = false
}

variable "dash0_otlp_endpoint" {
  type        = string
  default     = ""
  description = "OTLP/gRPC ingress endpoint, e.g. ingress.us-west-2.aws.dash0.com:4317"
}

variable "dash0_api_endpoint" {
  type        = string
  default     = ""
  description = "e.g. https://api.us-west-2.aws.dash0.com"
}

variable "enable_radar" {
  type    = bool
  default = false
}

variable "radar_cluster_id" {
  type        = string
  default     = ""
  description = "Cluster ID from Radar Cloud (infra/scripts/radar-create-cluster.sh prints it)"
}

variable "versions" {
  type = map(string)
  default = {
    envoy_gateway = "1.10.0"
    cert_manager  = "v1.21.3"
    external_dns  = "1.23.0"
    eck_operator  = "3.5.0"
    argo_cd       = "10.17.0"
    dash0         = "0.159.0"
  }
}

variable "enable_sandbox" {
  type        = bool
  default     = false
  description = "Learning sandbox: otel-demo namespace with a Dash0 export secret"
}
