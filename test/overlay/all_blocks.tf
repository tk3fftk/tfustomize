
provider "aws" {
  region = "ap-northeast-1"
}

resource "aws_instance" "example" {
  instance_type = "t2.medium"
}

variable "image_id" {
  description = "foo"
}

# there is no patch for output block

data "aws_ami" "example" {
  most_recent = false
}

module "vpc" {
  name = "staging-vpc"
}

locals {
  b = 2
}

terraform {
  required_version = ">= 1.0"
}
