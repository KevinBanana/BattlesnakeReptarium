resource "aws_ecr_repository" "battlesnake" {
  name                 = "battlesnake"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = local.common_tags
}

output "battlesnake_app_repo_url" {
  value = aws_ecr_repository.battlesnake.repository_url
}

resource "aws_ecr_lifecycle_policy" "image_deletion_rules" {
  repository = aws_ecr_repository.battlesnake.name
  policy     = <<EOF
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Expire untagged images older than 3 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 3
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 2,
      "description": "Keep only the 3 most recent tagged images",
      "selection": {
        "tagStatus": "tagged",
        "tagPatternList": [".*"],
        "countType": "imageCountMoreThan",
        "countNumber": 3
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
EOF
}