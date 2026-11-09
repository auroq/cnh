output "project_id" {
  value = digitalocean_project.cnh.id
}

output "bucket" {
  value = digitalocean_spaces_bucket.tfstate.name
}
