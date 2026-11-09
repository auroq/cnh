provider "onepassword" {}

data "onepassword_vault" "cnh" {
  name = "CNH"
}

data "onepassword_item" "do_token" {
  vault = data.onepassword_vault.cnh.uuid
  title = "Digital Ocean API CNH"
}

provider "digitalocean" {
  token = data.onepassword_item.do_token.credential
}

provider "digitalocean" {
  alias             = "spaces"
  token             = data.onepassword_item.do_token.credential
  spaces_access_id  = digitalocean_spaces_key.tfstate.access_key
  spaces_secret_key = digitalocean_spaces_key.tfstate.secret_key
}

resource "digitalocean_project" "cnh" {
  name        = "cnh"
  description = "Cloud Native Hackathon, KubeCon NA 2026 Salt Lake City"
  purpose     = "Web Application"
  environment = "Development"
}

resource "digitalocean_spaces_key" "tfstate" {
  name = "cnh-tfstate"

  grant {
    bucket     = ""
    permission = "fullaccess"
  }
}

resource "onepassword_item" "spaces_tfstate" {
  vault    = data.onepassword_vault.cnh.uuid
  title    = "cnh-tfstate"
  category = "login"
  username = digitalocean_spaces_key.tfstate.access_key
  password = digitalocean_spaces_key.tfstate.secret_key
  url      = "https://sfo3.digitaloceanspaces.com"
}

resource "digitalocean_spaces_bucket" "tfstate" {
  provider      = digitalocean.spaces
  name          = "cnh-tfstate"
  region        = "sfo3"
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }
}

resource "digitalocean_project_resources" "tfstate" {
  project   = digitalocean_project.cnh.id
  resources = [digitalocean_spaces_bucket.tfstate.urn]
}
