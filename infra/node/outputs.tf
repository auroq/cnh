output "ipv4" {
  value = digitalocean_droplet.node.ipv4_address
}

output "ssh" {
  value = "ssh root@${digitalocean_droplet.node.ipv4_address}"
}

output "hostname" {
  value = "${var.hostname}.${var.domain}"
}

output "dns_published" {
  value = var.publish_dns
}
