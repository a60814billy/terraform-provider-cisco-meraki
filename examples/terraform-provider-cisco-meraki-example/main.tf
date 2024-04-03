terraform {
  required_providers {
    ciscomeraki = {
      source = "hashicorp.com/a60814billy/cisco-meraki"
    }
  }
}

variable "api_key" {
  type        = string
  description = "The Meraki API key"
  default = "1d8b4f977530c14b268886a45096272c99ed727c"
}

provider "ciscomeraki" {
  api_key = var.api_key
}

data "ciscomeraki_orgs" "orgs" {}

data "ciscomeraki_org" "first_org" {
    id = data.ciscomeraki_orgs.orgs.ids.0
}

output "orgs" {
  value = data.ciscomeraki_orgs.orgs
}

output "org" {
  value = data.ciscomeraki_org.first_org
}