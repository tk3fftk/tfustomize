data "aws_ami" "ubuntu" {
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
