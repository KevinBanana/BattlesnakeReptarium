data "aws_ami" "latest_amazon_linux" {
  most_recent = true
  owners = ["amazon"]

  filter {
    name = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_instance" "battlesnake" {
  ami                         = data.aws_ami.latest_amazon_linux.id
  associate_public_ip_address = true
  instance_type               = "t3.micro"
  iam_instance_profile        = aws_iam_instance_profile.ec2_instance_profile.name
  subnet_id                   = aws_subnet.public.id
  vpc_security_group_ids = [
    aws_security_group.public_traffic.id
  ]
  root_block_device {
    delete_on_termination = true
    volume_size           = 10
    volume_type           = "gp2"
  }

  user_data = <<-EOF
    #!/bin/bash
    # Update packages and install Docker and AWS CLI
    apt-get update -y
    apt-get install -y docker.io awscli

    # Start Docker service if not already running
    systemctl start docker
    systemctl enable docker

    # Authenticate Docker to your ECR registry
    # Note: Replace region and account details accordingly.
    aws ecr get-login-password --region ${local.region} | sudo docker login --username AWS --password-stdin ${local.account_id}.dkr.ecr.${local.region}.amazonaws.com

    # Pull and run your Docker image from ECR
    docker pull ${local.ecr_image_uri}
    docker run -d -p 8080:8080 ${local.ecr_image_uri}
  EOF

  tags = merge(local.common_tags, {
    Name = "battlesnake-ec2"
  })
}

output "ec2_instance_public_ip" {
  value = aws_instance.battlesnake.public_ip
}