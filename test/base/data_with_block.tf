data "aws_ami" "ubuntu" {
  filter {
    # tfustomize:block_merge:name
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}
