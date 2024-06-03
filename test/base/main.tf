locals {
  a = 1
  b = 2
  c = 3
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    # tfustomize:merge_block:name
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "web" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "HelloWorld"
  }
}

locals {
  c = 4
}

module "servers" {
  source = "./app-cluster"

  servers = 5
}
