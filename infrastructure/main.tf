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

resource "aws_elastic_beanstalk_environment" "prod_env" {
  name                = "prod-battlesnake"
  application         = aws_elastic_beanstalk_application.battlesnake.name
  solution_stack_name = "64bit Amazon Linux 2023 v4.0.0 running Go 1"


  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.battlesnake-ec2-iam-instance-profile.name
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "DisableIMDSv1"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENVIRONMENT"
    value     = "prod"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "PORT"
    value     = "8080"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "EnvironmentType"
    value     = "SingleInstance"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = "AWSServiceRoleForElasticBeanstalk"
  }
}
