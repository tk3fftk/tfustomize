package api_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/tk3fftk/tfustomize/api"
)

var regexpFormatNewLines = regexp.MustCompile(`\n{2,}`)

func TestCollectHCLFilePaths(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		paths   []string
		expect  []string
		wantErr bool
	}{
		{
			name:    "directory",
			baseDir: "../test",
			paths:   []string{"./collect_hcl_file_paths"},
			expect:  []string{"../test/collect_hcl_file_paths/1.tf", "../test/collect_hcl_file_paths/2.tf"},
			wantErr: false,
		},
		{
			name:    "file",
			baseDir: "../test",
			paths:   []string{"./collect_hcl_file_paths/1.tf"},
			expect:  []string{"../test/collect_hcl_file_paths/1.tf"},
			wantErr: false,
		},
		{
			name:    "not found",
			baseDir: "../test",
			paths:   []string{"./collect_hcl_file_paths/not_found.tf"},
			expect:  nil,
			wantErr: true,
		},
	}

	parser := api.HCLParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.CollectHCLFilePaths(tt.baseDir, tt.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectHCLFilePaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expect, got)
		})
	}

}

func TestConcatFiles(t *testing.T) {
	tests := []struct {
		name     string
		contents []string
		expect   string
		wantErr  bool
	}{
		{
			name:     "attribute in a file",
			contents: []string{`bucket = "foo"`},
			expect:   ``, // attribute is not a block so it results empty
			wantErr:  false,
		},
		{
			name: "a file",
			contents: []string{`resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
`,
			},
			expect: `resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
`,
			wantErr: false,
		},
		{
			name: "multiple files",
			contents: []string{`resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
`,
				`resource "aws_s3_bucket" "bar" {
  bucket = "bar"
}
`,
			},
			expect: `resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
resource "aws_s3_bucket" "bar" {
  bucket = "bar"
}
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "test_concat_file")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			var fileNames []string
			parser := api.HCLParser{}

			for i, content := range tt.contents {
				fileName := strconv.Itoa(i) + ".tf"
				filePath := filepath.Join(dir, fileName)
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				fileNames = append(fileNames, fileName)
			}

			filePaths, err := parser.CollectHCLFilePaths(dir, fileNames)
			if err != nil {
				t.Fatal(err)
			}

			hclFile, err := parser.ConcatFiles(filePaths)
			if err != nil && !tt.wantErr {
				t.Errorf("%q. ConcatFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				assert.Fail(t, "unexpected error")
			} else {
				assert.Equal(t, tt.expect, string(hclwrite.Format(hclFile.Bytes())))
			}
		})
	}
}

func TestMergeFileBlocks(t *testing.T) {
	tests := []struct {
		name    string
		base    []string
		overlay []string
		expect  string
		wantErr bool
	}{
		{
			name:    "locals merge test",
			base:    []string{"base/only_locals.tf"},
			overlay: []string{"overlay/only_locals.tf"},
			expect: `locals {
  a = 1
  b = 2
  c = 3
  d = 4
}
`,
			wantErr: false,
		},
		{
			name:    "data source without blocks merge test",
			base:    []string{"base/data_without_block.tf"},
			overlay: []string{"overlay/data_without_block.tf"},
			expect: `data "aws_ami" "ubuntu" {
  executable_users   = ["self"]
  include_deprecated = true
  most_recent        = true
  name_regex         = "^myami-\\d{3}"
  owners             = ["099720109477"]
}
`,
			wantErr: false,
		},
		{
			name:    "data source with append block merge test",
			base:    []string{"base/data_with_block.tf"},
			overlay: []string{"overlay/data_with_block.tf"},
			expect: `data "aws_ami" "ubuntu" {
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    # tfustomize:block_merge:name
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}
`,
			wantErr: false,
		},
		{
			name:    "data source with merge block test",
			base:    []string{"base/data_with_block.tf"},
			overlay: []string{"overlay/data_with_block_merge.tf"},
			expect: `data "aws_ami" "ubuntu" {
  filter {
    name   = "name_is_updated"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-24.04-amd64-server-*"]
  }
}
`,
			wantErr: false,
		},
		{
			name:    "data source and resource",
			base:    []string{"base/data_and_resource.tf"},
			overlay: []string{"overlay/data_and_resource.tf"},
			expect: `data "aws_ami" "ubuntu" {
  most_recent = false
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}
resource "aws_instance" "web" {
  ami               = data.aws_ami.ubuntu.id
  availability_zone = "ap-northeast-1a"
  instance_type     = "t3.large"
  tags = {
    Name = "HelloWorld"
  }
}
`,
			wantErr: false,
		},
		{
			name:    "all types of blocks",
			base:    []string{"base/all_blocks.tf"},
			overlay: []string{"overlay/all_blocks.tf"},
			expect: `import {
  to = aws_instance.example2
  id = "i-qwer5678"
}
removed {
  from = aws_instance.example2
  lifecycle {
    destroy = true
  }
}
moved {
  from = aws_instance.old_name2
  to   = aws_instance.new_name2
}
locals {
  a = 1
  b = 2
}
data "aws_ami" "example" {
  most_recent = false
  owners      = ["self"]
}
module "vpc" {
  cidr    = "10.0.0.0/16"
  name    = "staging-vpc"
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.77.0"
}
output "instance_ip_addr" {
  value = aws_instance.example.public_ip
}
provider "aws" {
  region = "ap-northeast-1"
}
resource "aws_instance" "example" {
  ami           = "ami-0c94855ba95c574c8"
  instance_type = "t2.medium"
}
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}
variable "image_id" {
  default     = "ami-0c94855ba95c574c8"
  description = "foo"
}
moved {
  from = aws_instance.old_name
  to   = aws_instance.new_name
}
import {
  to = aws_instance.example
  id = "i-abcd1234"
}
removed {
  from = aws_instance.example
  lifecycle {
    destroy = false
  }
}
`,
			wantErr: false,
		},
	}

	testDir := "../test"
	parser := api.HCLParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePaths, err := parser.CollectHCLFilePaths(testDir, tt.base)
			if err != nil {
				t.Fatal(err)
			}
			baseHCL, err := parser.ConcatFiles(basePaths)
			if err != nil {
				t.Fatal(err)
			}
			overlayPaths, err := parser.CollectHCLFilePaths(testDir, tt.overlay)
			if err != nil {
				t.Fatal(err)
			}
			overlayHCL, err := parser.ConcatFiles(overlayPaths)
			if err != nil {
				t.Fatal(err)
			}

			result, err := parser.MergeFileBlocks(baseHCL, overlayHCL)
			if err != nil && !tt.wantErr {
				t.Errorf("MergeFileBlocks() error = %v, wantErr %v", err, tt.wantErr)
				assert.Fail(t, "unexpected error")
			} else {
				assert.Equal(t, tt.expect, regexpFormatNewLines.ReplaceAllString(string(hclwrite.Format(result.Bytes())), "\n"))
			}
		})
	}
}
