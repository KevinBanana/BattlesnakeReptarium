provider "aws" {
  region = var.aws_region
  profile = "battlesnake-terraform"
  default_tags {
    tags = {
      Name = "battlesnake"
    }
  }
}

resource "aws_elastic_beanstalk_application" "battlesnake" {
  name = "battlesnake"
  description = "All battlesnake environments"


}