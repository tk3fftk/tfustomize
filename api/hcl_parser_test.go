package api_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/tk3fftk/tfustomize/api"
	"github.com/zclconf/go-cty/cty"
)

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
		base    *hclwrite.File
		overlay *hclwrite.File
		wantErr bool
	}{
		{
			name:    "valid merge",
			base:    hclwrite.NewEmptyFile(),
			overlay: hclwrite.NewEmptyFile(),
			wantErr: false,
		},
		// Add more test cases here
	}

	parser := api.HCLParser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.MergeFileBlocks(tt.base, tt.overlay)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeFileBlocks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
