terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

# Set the variable value in *.tfvars file
# or using -var="do_token=..." CLI option
variable "do_token" {}

# Configure the DigitalOcean Provider
provider "digitalocean" {
  token = var.do_token
}

# Create a web server
resource "digitalocean_droplet" "web" {
  image   = "ubuntu-20-04-x64"
  name    = "web-1"
  region  = "nyc2"
  size    = "s-1vcpu-1gb"
  backups = false
  backup_policy {
    plan    = "weekly"
    weekday = "TUE"
    hour    = 8
  }
}