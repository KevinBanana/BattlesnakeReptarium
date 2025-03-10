locals {
  environment = "prod"
  region      = "us-east-1"
  project     = "battlesnake"
  account_id  = "265577504730"
  common_tags = {
    env     = local.environment
    project = local.project
  }

  # Availability Zones
  azs_count = 2
  azs_names = data.aws_availability_zones.available.names
}