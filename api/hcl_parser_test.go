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
	"github.com/zclconf/go-cty/cty"
)

var regexpFormatNewLines = regexp.MustCompile(`\n{2,}`)

func TestPatchFileAttributes(t *testing.T) {
	tests := []struct {
		name    string
		base    map[string]string
		overlay map[string]string
		expect  string
		wantErr bool
	}{
		{
			name: "no overlay",
			base: map[string]string{
				"foo": "bar",
			},
			overlay: map[string]string{},
			expect: `foo = "bar"
`,
			wantErr: false,
		},
		{
			name: "single var patch",
			base: map[string]string{
				"foo": "bar",
			},
			overlay: map[string]string{
				"foo": "baz",
			},
			expect: `foo = "baz"
`,
			wantErr: false,
		},
		{
			name: "multi var and single patch",
			base: map[string]string{
				"foo":  "bar",
				"hoge": "fuga",
			},
			overlay: map[string]string{
				"foo": "baz",
			},
			expect: `foo  = "baz"
hoge = "fuga"
`,
			wantErr: false,
		},
		{
			name: "multi var and single patch and add new var",
			base: map[string]string{
				"foo":  "bar",
				"hoge": "fuga",
			},
			overlay: map[string]string{
				"foo":  "baz",
				"nyan": "meow",
			},
			expect: `foo  = "baz"
hoge = "fuga"
nyan = "meow"
`,
			wantErr: false,
		},
	}

	parser := api.HCLParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseFile := hclwrite.NewEmptyFile()
			overlayFile := hclwrite.NewEmptyFile()

			baseBody := baseFile.Body()
			overlayBody := overlayFile.Body()

			for k, v := range tt.base {
				baseBody.SetAttributeValue(k, cty.StringVal(v))
			}
			for k, v := range tt.overlay {
				overlayBody.SetAttributeValue(k, cty.StringVal(v))
			}

			_, err := parser.PatchFileAttributes(baseFile, overlayFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeFileBlocks() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expect, string(hclwrite.Format(baseFile.Bytes())))
		})
	}
}

func TestConcatFile(t *testing.T) {
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

			hclFile, err := parser.ConcatFile(dir, fileNames)
			if (err != nil) || tt.wantErr {
				t.Errorf("%q. ConcatFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			assert.Equal(t, tt.expect, string(hclwrite.Format(hclFile.Bytes())))
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
  most_recent        = true
  name_regex         = "^myami-\\d{3}"
  owners             = ["099720109477"]
  include_deprecated = true
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
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
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
			baseHCL, err := parser.ConcatFile(testDir, tt.base)
			if err != nil {
				t.Fatal(err)
			}
			overlayHCL, err := parser.ConcatFile(testDir, tt.overlay)
			if err != nil {
				t.Fatal(err)
			}

			result, err := parser.MergeFileBlocks(baseHCL, overlayHCL)
			if (err != nil) || tt.wantErr {
				t.Errorf("MergeFileBlocks() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.expect, regexpFormatNewLines.ReplaceAllString(string(hclwrite.Format(result.Bytes())), "\n"))
		})
	}
}
