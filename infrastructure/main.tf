provider "aws" {
  region  = var.aws_region
  profile = "battlesnake-terraform"
  default_tags {
    tags = {
      Name = "battlesnake"
    }
  }
}

resource "aws_elastic_beanstalk_application" "battlesnake" {
  name        = "battlesnake"
  description = "All battlesnake environments"
}

resource "aws_elastic_beanstalk_environment" "preprod_env" {
  name                = "preprod-battlesnake"
  application         = aws_elastic_beanstalk_application.battlesnake.name
  solution_stack_name = "64bit Amazon Linux 2 v3.8.2 running Go 1"

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.beanstalk_iam_instance_profile.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENVIRONMENT"
    value     = "preprod"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "PORT"
    value     = "8080"
  }
}