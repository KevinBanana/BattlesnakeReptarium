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
  instance_type               = "t2.micro"
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

  key_name = "battlesnake-key"
  connection {
    type = "ssh"
    user = "ec2-user"
    private_key = file("./../../.ssh/battlesnake-key.pem")
    host = self.public_ip
  }

  provisioner "remote-exec" {
    inline = [
      "aws ecr get-login-password --region ${local.region} | sudo docker login --username AWS --password-stdin ${local.account_id}.dkr.ecr.${local.region}.amazonaws.com",
      "sudo docker pull ${local.ecr_image_uri}",
      "sudo docker run -d -p 8080:8080 ${local.ecr_image_uri}"
    ]
  }

  tags = merge(local.common_tags, {
    Name = "battlesnake-ec2"
  })
}

output "ec2_instance_public_ip" {
  value = aws_instance.battlesnake.public_ip
}