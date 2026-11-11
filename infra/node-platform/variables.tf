variable "kubeconfig_path" {
  type        = string
  default     = "../../.kube/node.config"
  description = "Admin kubeconfig fetched by `make node-kubeconfig`, relative to this root"
}

variable "cluster_name" {
  type    = string
  default = "wide-door-node"
}

variable "app_namespace" {
  type    = string
  default = "wide-door"
}

variable "es_password_item_title" {
  type    = string
  default = "Wide Door node elastic"
}

variable "dash0_item_title" {
  type    = string
  default = "Dash0 API Token CNH"
}

variable "argo_cd_version" {
  type    = string
  default = "10.17.0"
}

variable "application" {
  type = object({
    repo_url        = string
    target_revision = string
    path            = string
  })
  default = {
    repo_url        = "https://github.com/auroq/cnh.git"
    target_revision = "main"
    path            = "deploy/node"
  }
}

variable "enable_cd_gate" {
  type        = bool
  default     = true
  description = "Create the cd-gate ServiceAccount, its token and the CD_KUBECONFIG repository secret (ADR 0008). Reads the GitHub token from the 1Password item \"CNH GitHub CD\"."
}

variable "github_repository" {
  type    = string
  default = "auroq/cnh"
}

variable "cd_gate_token_ttl_seconds" {
  type        = number
  default     = 1209600
  description = "Lifetime of the cd-gate token; 14 days covers the week after the event, re-apply to rotate."
}
