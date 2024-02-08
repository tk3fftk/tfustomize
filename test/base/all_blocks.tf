
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "example" {
  ami           = "ami-0c94855ba95c574c8"
  instance_type = "t2.micro"
}

variable "image_id" {
  description = "The id of the machine image (AMI) to use for the server"
  default     = "ami-0c94855ba95c574c8"
}

output "instance_ip_addr" {
  value = aws_instance.example.public_ip
}

data "aws_ami" "example" {
  most_recent = true
  owners      = ["self"]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.77.0"
  name    = "my-vpc"
  cidr    = "10.0.0.0/16"
}

locals {
  a = 1
}

terraform {
  required_version = ">= 0.12"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}
