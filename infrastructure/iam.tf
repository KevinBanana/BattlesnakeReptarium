resource "aws_iam_instance_profile" "battlesnake-ec2-iam-instance-profile" {
  name = "battlesnake-ec2-iam-instance-profile"
  role = aws_iam_role.role.name
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "role" {
  name               = "battlesnake_ec2_role"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}