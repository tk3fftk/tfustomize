data "aws_ami" "ubuntu" {

  filter {
    # tfustomize:merge_block:name
    name   = "name_is_updated"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-24.04-amd64-server-*"]
  }
}
