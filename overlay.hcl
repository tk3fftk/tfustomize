log_level = "warn"
type = {
  "a" = 1
  "b" = 2
  "c" = 3
}
CI=true
threads = 1

resource "aws_instance" "web" {
  instance_type = "t3.medium"
  availability_zone = "ap-northeast-1a"
}

data "aws_ami" "ubuntu" {
  most_recent = false

  filter {
    name   = "arch"
    values = ["arm64"]
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
