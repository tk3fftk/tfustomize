# tfustomize

![test](https://github.com/tk3fftk/tfustomize/actions/workflows/test.yaml/badge.svg) ![goreleaser](https://github.com/tk3fftk/tfustomize/actions/workflows/goreleaser.yaml/badge.svg) ![release](https://img.shields.io/github/v/release/tk3fftk/tfustomize)

`tfustomize` is yet another way to reuse resource configurations with Terraform.
`tfustomize` is literally inspired by `kustomize`, a Kubernetes manifests management tool.

## Motivation

[Terraform Modules](https://developer.hashicorp.com/terraform/language/modules) is looks almost the only way to reuse resource configurations with Terraform. There is other option [`terragrunt`](https://terragrunt.gruntwork.io/), but it's looks expanding and wrapping the use of Terraform module.  
Terraform modules are effective when they are widly spreaded as open sources, or when they are officially provided by kind of Platform Engineers for internal use in private environments. However, creating custom modules for one product seems like overkill.  
It was quite labor-intensive to transition from a copy-paste style of management to using custom modules, so I created something like a Terraform version of kustomize as a proof of concept.  
[Override Files](https://developer.hashicorp.com/terraform/language/files/override) feature is looks similar concept with `tfustomize` but overriding is not for reusing purpose.

## Installation

```sh
# using Go
go install github.com/tk3fftk/tfustomize@latest

# from Releases
curl -L https://github.com/tk3fftk/tfustomize/releases/download/v0.2.0/tfustomize_0.2.0_Linux_x86_64.tar.gz | tar xvz
```

## Usage

💡[example](./example/) provides instructions for generating Terraform configuration files using `tfustomize`.

`tfustomize` reads a `tfustomization.hcl` file and generated a directory and a file as a result of merging.
Users can use the output file to run `terraform plan` and `terraform apply`.

```sh
Usage:
  tfustomize build [dir] [flags]

Flags:
  -h, --help             help for build
  -o, --out string       Output directory (default "generated")
  -f, --outfile string   Output filename (default "main.tf")
  -p, --print            Print the result to the console instead of writing to a file

Global Flags:
  -d, --debug   Enable debug mode
```

The format of `tfustomization.hcl` is following.

```hcl
tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "../base",
  ]
}

patches {
  paths = [
    "./main.tf",
  ]
}
```

- `tfustomize` block:
  - Currently it's just a placeholder.
- `resources` block:
  - Specify "base" configuration files.
  - directory or file name are available.
- `patches` block:
  - Specify "overlay" configuration files.
  - directory or file name are available.

### Merging Behavior and Limitation

- A Top-level block has the same block type and labels in base and overlay will be merged.
  - Except `moved`, `import`, `removed` block. These will be appended.
- `locals` blocks will be merged.
- Within a top-level block, an attribute argument within an overlay block will be replaced any argument of the same name in the base block.
- Within a top-level block, any block will be appended by default.
  - To merge a block, use an anotation `# tfustimize:block_merge:<key>` both a base and an overlay like below.

```hcl
# base
data "aws_ami" "ubuntu" {
  filter {
    # tfustomize:block_merge:name
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

# overlay
data "aws_ami" "ubuntu" {
  filter {
    name   = "arch"
    values = ["arm64"]
  }
  filter {
    # tfustomize:block_merge:name
    name   = "name_is_updated"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-24.04-amd64-server-*"]
  }
}

# output
data "aws_ami" "ubuntu" {
  filter {
    name   = "arch"
    values = ["arm64"]
  }
  filter {
    name   = "name_is_updated"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-24.04-amd64-server-*"]
  }
}
```

- [Limitation] The output order is randomized inside block level order. (https://github.com/tk3fftk/tfustomize/issues/6)

### A sample Terraform directory structure with `tfustomize`

```bash
someApp
├── production
│   ├── backend.tf
│   ├── main.tf
│   ├── outputs.tf
│   ├── tfustomization.hcl
│   └── variables.tf
└── staging
    ├── backend.tf
    ├── main.tf
    ├── outputs.tf
    ├── tfustomization.hcl
    └── variables.tf
```

- `staging/tfustomization.hcl`

```hcl
tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "./",
  ]
}

patches {
  paths = []
}
```

- `production/tfustomization.hcl`

```hcl
tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "../staging",
  ]
}

patches {
  paths = [
    "./",
  ]
}
```
