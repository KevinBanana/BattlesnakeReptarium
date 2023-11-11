provider "aws" {
  region = var.aws_region
  profile = "battlesnake-terraform"
  default_tags {
    tags = {
      Name = "battlesnake"
    }
  }
}
locals {
  container_name = "battlesnake-container"
  container_port = 8080 # Must be the same as the port exposed in the Dockerfile
}

data "aws_caller_identity" "this" {}
data "aws_ecr_authorization_token" "this" {

}

locals {
  ecr_address = format("%v.dkr.ecr.%v.amazonaws.com", data.aws_caller_identity.this.account_id, var.aws_region)
}
provider "docker" {
  registry_auth {
    address  = local.ecr_address
    password = data.aws_ecr_authorization_token.this.password
    username = data.aws_ecr_authorization_token.this.user_name
  }
}

module "ecr" {
  source = "terraform-aws-modules/ecr/aws"
  version = "~> 1.6.0"

  repository_force_delete = true
  repository_name         = "battlesnake"
  repository_lifecycle_policy = jsonencode({
    rules = [{
      action = {
        type = "expire"
      }
      description = "Delete all images except a handful of the newest images"
      rulePriority = 1
      selection = {
        countType = "imageCountMoreThan"
        countNumber = 3
        tagStatus = "any"
      }
    }]
  })
}

resource "docker_image" "this" {
  name = format("%v:%v", module.ecr.repository_url, formatdate("YYYY-MM-DD'T'hh-mm-ss", timestamp()))
  build {
    dockerfile = "../Dockerfile"
    context = "../."
  }
}

resource "docker_registry_image" "this" {
  keep_remotely = true
  name = resource.docker_image.this.name
}

