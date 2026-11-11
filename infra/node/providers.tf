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
