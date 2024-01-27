data "aws_ami" "ubuntu" {
  owners             = ["099720109477"] # Canonical
  include_deprecated = true
}

