output "cluster_id" {
  value = digitalocean_kubernetes_cluster.cnh.id
}

output "cluster_name" {
  value = digitalocean_kubernetes_cluster.cnh.name
}

output "endpoint" {
  value = digitalocean_kubernetes_cluster.cnh.endpoint
}

output "cluster_ca_certificate" {
  value     = digitalocean_kubernetes_cluster.cnh.kube_config[0].cluster_ca_certificate
  sensitive = true
}

output "kubeconfig" {
  value     = digitalocean_kubernetes_cluster.cnh.kube_config[0].raw_config
  sensitive = true
}

output "domain" {
  value = digitalocean_domain.cnh.name
}

output "nameservers" {
  value = ["ns1.digitalocean.com", "ns2.digitalocean.com", "ns3.digitalocean.com"]
}
