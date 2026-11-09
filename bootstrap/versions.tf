terraform {
  required_version = ">= 1.10"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.107.0"
    }
    onepassword = {
      source  = "1Password/onepassword"
      version = "3.4.1"
    }
  }
}
