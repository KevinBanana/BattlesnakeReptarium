resource "aws_iam_role" "beanstalk_role" {
  name = "BeanstalkAppRole"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticbeanstalk.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "beanstalk_policy" {
  name        = "BeanstalkAppPolicy"
  description = "Policy for Elastic Beanstalk application"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowCloudformationOperationsOnElasticBeanstalkStacks",
            "Effect": "Allow",
            "Action": [
                "cloudformation:*"
            ],
            "Resource": [
                "arn:aws:cloudformation:*:*:stack/awseb-*",
                "arn:aws:cloudformation:*:*:stack/eb-*"
            ]
        },
        {
            "Sid": "AllowDeleteCloudwatchLogGroups",
            "Effect": "Allow",
            "Action": [
                "logs:DeleteLogGroup"
            ],
            "Resource": [
                "arn:aws:logs:*:*:log-group:/aws/elasticbeanstalk*"
            ]
        },
        {
            "Sid": "AllowECSTagResource",
            "Effect": "Allow",
            "Action": [
                "ecs:TagResource"
            ],
            "Resource": "*",
            "Condition": {
                "StringEquals": {
                    "ecs:CreateAction": [
                        "CreateCluster",
                        "RegisterTaskDefinition"
                    ]
                }
            }
        },
        {
            "Sid": "AllowS3OperationsOnElasticBeanstalkBuckets",
            "Effect": "Allow",
            "Action": [
                "s3:*"
            ],
            "Resource": [
                "arn:aws:s3:::elasticbeanstalk-*",
                "arn:aws:s3:::elasticbeanstalk-*/*"
            ]
        },
        {
            "Sid": "AllowLaunchTemplateRunInstances",
            "Effect": "Allow",
            "Action": "ec2:RunInstances",
            "Resource": "*",
            "Condition": {
                "ArnLike": {
                    "ec2:LaunchTemplate": "arn:aws:ec2:*:*:launch-template/*"
                }
            }
        },
        {
            "Sid": "AllowELBAddTags",
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:AddTags"
            ],
            "Resource": "*",
            "Condition": {
                "StringEquals": {
                    "elasticloadbalancing:CreateAction": [
                        "CreateLoadBalancer"
                    ]
                }
            }
        },
        {
            "Sid": "AllowOperations",
            "Effect": "Allow",
            "Action": [
                "autoscaling:AttachInstances",
                "autoscaling:CreateAutoScalingGroup",
                "autoscaling:CreateLaunchConfiguration",
                "autoscaling:CreateOrUpdateTags",
                "autoscaling:DeleteLaunchConfiguration",
                "autoscaling:DeleteAutoScalingGroup",
                "autoscaling:DeleteScheduledAction",
                "autoscaling:DescribeAccountLimits",
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:DescribeLaunchConfigurations",
                "autoscaling:DescribeLoadBalancers",
                "autoscaling:DescribeNotificationConfigurations",
                "autoscaling:DescribeScalingActivities",
                "autoscaling:DescribeScheduledActions",
                "autoscaling:DetachInstances",
                "autoscaling:DeletePolicy",
                "autoscaling:PutScalingPolicy",
                "autoscaling:PutScheduledUpdateGroupAction",
                "autoscaling:PutNotificationConfiguration",
                "autoscaling:ResumeProcesses",
                "autoscaling:SetDesiredCapacity",
                "autoscaling:SuspendProcesses",
                "autoscaling:TerminateInstanceInAutoScalingGroup",
                "autoscaling:UpdateAutoScalingGroup",
                "cloudwatch:PutMetricAlarm",
                "ec2:*",
                "ecs:CreateCluster",
                "ecs:DeleteCluster",
                "ecs:DescribeClusters",
                "ecs:RegisterTaskDefinition",
                "elasticbeanstalk:*",
                "elasticloadbalancing:ApplySecurityGroupsToLoadBalancer",
                "elasticloadbalancing:ConfigureHealthCheck",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
                "elasticloadbalancing:DescribeInstanceHealth",
                "elasticloadbalancing:DescribeLoadBalancers",
                "elasticloadbalancing:DescribeTargetHealth",
                "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
                "elasticloadbalancing:DescribeTargetGroups",
                "elasticloadbalancing:RegisterTargets",
                "elasticloadbalancing:DeregisterTargets",
                "iam:ListRoles",
                "iam:PassRole",
                "logs:CreateLogGroup",
                "logs:PutRetentionPolicy",
                "logs:DescribeLogGroups",
                "rds:DescribeDBEngineVersions",
                "rds:DescribeDBInstances",
                "rds:DescribeOrderableDBInstanceOptions",
                "s3:GetObject",
                "s3:GetObjectAcl",
                "s3:ListBucket",
                "sns:CreateTopic",
                "sns:GetTopicAttributes",
                "sns:ListSubscriptionsByTopic",
                "sns:Subscribe",
                "sns:SetTopicAttributes",
                "sqs:GetQueueAttributes",
                "sqs:GetQueueUrl",
                "codebuild:CreateProject",
                "codebuild:DeleteProject",
                "codebuild:BatchGetBuilds",
                "codebuild:StartBuild"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "beanstalk_attachment" {
  role       = aws_iam_role.beanstalk_role.name
  policy_arn = aws_iam_policy.beanstalk_policy.arn
}

resource "aws_iam_instance_profile" "beanstalk_iam_instance_profile" {
  name        = "beanstalk_iam_instance_profile"
  role = aws_iam_role.beanstalk_role.name
}