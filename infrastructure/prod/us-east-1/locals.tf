locals {
  environment = "prod"
  region      = "us-east-1"
  project     = "battlesnake"
  common_tags = {
    env     = local.environment
    project = local.project
  }


}