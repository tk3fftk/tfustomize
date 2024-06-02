locals {
  a = 1
  b = 2
  c = 4
  d = 100
}

resource "aws_instance" "web" {
  instance_type     = "t3.medium"
  availability_zone = "ap-northeast-1a"
}

data "aws_ami" "ubuntu" {
  most_recent = false

  filter {
    name   = "arch"
    values = ["arm64"]
  }
  filter {
    # tfustomize:block_merge:name
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-24.04-amd64-server-*"]
  }
}

resource "aws_instance" "be" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.medium"

  tags = {
    Name = "HelloWorld_Backend"
  }
  availability_zone = "ap-northeast-1a"
}

module "servers" {
  servers = 1
}
