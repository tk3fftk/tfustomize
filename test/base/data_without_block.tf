data "aws_ami" "ubuntu" {
  executable_users = ["self"]
  name_regex       = "^myami-\\d{3}"
  owners           = ["self"]
}
