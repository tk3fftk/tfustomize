data "aws_ami" "ubuntu" {
  most_recent = false
}

resource "aws_instance" "web" {
  instance_type     = "t3.large"
  availability_zone = "ap-northeast-1a"
}
