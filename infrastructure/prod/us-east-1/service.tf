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
  instance_type               = "t2.nano"
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
    private_key = file("./../../../.ssh/battlesnake-key.pem")
    host = self.public_ip
  }

  provisioner "remote-exec" {
    inline = [
      "rm -f /home/ec2-user/app",
      "mkdir -p /home/ec2-user/app",
    ]
  }

  provisioner "file" {
    source      = "../../../"
    destination = "/home/ec2-user/app"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo yum update -y",
      "sudo yum install -y docker",
      "sudo service docker start",
      "sudo usermod -aG docker ec2-user",
      "cd /home/ec2-user/app",
      "sudo docker build -t go-app .",
      "sudo docker run -d -p 8080:8080 go-app"
    ]
  }

  tags = merge(local.common_tags, {
    Name = "battlesnake-ec2"
  })
}

resource "aws_security_group" "public_traffic" {
  description = "Security group allowing traffic on port 443"
  name        = "battlesnake-public-traffic"
  vpc_id      = aws_vpc.vpc1.id

  tags = local.common_tags
}

resource "aws_vpc_security_group_ingress_rule" "public_traffic" {
  security_group_id = aws_security_group.public_traffic.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 8080
  to_port           = 8080
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "ssh_ingress" {
  security_group_id = aws_security_group.public_traffic.id
  cidr_ipv4         = "136.55.173.17/32"
  from_port         = 22
  to_port           = 22
  ip_protocol       = "tcp"
}

output "ec2_instance_public_ip" {
  value = aws_instance.battlesnake.public_ip
}