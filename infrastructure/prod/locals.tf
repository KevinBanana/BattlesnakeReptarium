locals {
  environment = "prod"
  region      = "us-east-1"
  project     = "battlesnake"
  account_id  = "265577504730"
  common_tags = {
    env     = local.environment
    project = local.project
  }
  ecr_image_uri = "${local.account_id}.dkr.ecr.${local.region}.amazonaws.com/battlesnake:${var.image_tag}"
}