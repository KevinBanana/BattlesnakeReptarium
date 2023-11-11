#resource "aws_iam_role" "beanstalk_role" {
#  name = "BeanstalkAppRole"
#
#  assume_role_policy = <<EOF
#{
#  "Version": "2012-10-17",
#  "Statement": [
#    {
#      "Effect": "Allow",
#      "Principal": {
#        "Service": "elasticbeanstalk.amazonaws.com"
#      },
#      "Action": "sts:AssumeRole"
#    }
#  ]
#}
#EOF
#}
#
#resource "aws_iam_policy" "beanstalk_policy" {
#  name        = "BeanstalkAppPolicy"
#  description = "Policy for Elastic Beanstalk application"
#
#  policy = <<EOF
#{
#    "Version": "2012-10-17",
#    "Statement": [
#        {
#            "Effect": "Allow",
#            "Action": [
#                "ecr:*",
#                "cloudtrail:LookupEvents"
#            ],
#            "Resource": "*"
#        }
#    ]
#}
#EOF
#}
#
#resource "aws_iam_role_policy_attachment" "beanstalk_attachment" {
#  role       = aws_iam_role.beanstalk_role.name
#  policy_arn = aws_iam_policy.beanstalk_policy.arn
#}
#
#resource "aws_iam_instance_profile" "beanstalk_iam_instance_profile" {
#  name        = "beanstalk_iam_instance_profile"
#  role = aws_iam_role.beanstalk_role.name
#}