# example

This document provides instructions for generating Terraform configuration files using `tfustomize`, and then applying those configurations to manage infrastructure in two environments: `staging` and `production`. These instructions use [docker provider](https://registry.terraform.io/providers/kreuzwerker/docker/latest) so `docker` is required in your environment.


- The directory structure is following:

```sh
example
├── README.md
├── production
│   ├── main.tf
│   └── tfustomization.hcl
└── staging
    ├── main.tf
    └── tfustomization.hcl
```

- `main.tf` and ` tfustomization.hcl` files in each dir are following:

```sh
# staging
$ cat staging/tfustomization.hcl
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
$ cat staging/main.tf
terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

provider "docker" {
  host = "unix:///var/run/docker.sock"
}

resource "docker_image" "ubuntu" {
  name = "ubuntu:latest"
}

resource "docker_container" "foo" {
  image   = docker_image.ubuntu.image_id
  name    = "foo-staging"
  command = ["/bin/bash", "-c", "sleep 100"]
}

# production
$ cat production/tfustomization.hcl
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
    "main.tf",
  ]
}
$ cat production/main.tf
resource "docker_image" "ubuntu" {
  name = "ubuntu:24.04"
}

resource "docker_container" "foo" {
  name  = "foo-production"
}
```


- Generate the Terraform configuration file using `tfustomize`.

```sh
tfustomize build staging
tfustomize build production
```

- Run `terraform` in each directories.

```sh
cd staging/generated
terraform init
terraform plan
terraform apply
cd ../../
```

```sh
cd production/generated
terraform init
terraform plan
terraform apply
cd ../../
```

- Check the resulting Docker images and running containers.

```sh
# Confirm pulled two images
$ docker image ls | grep ubuntu
ubuntu                                                                                  latest                                     a50ab9f16797   4 days ago      69.2MB
ubuntu                                                                                  24.04                                      1e6914c26da3   5 days ago      99.6MB

# Confirm running two containers
$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS              PORTS     NAMES
17cf2b1cb30e   1e6914c26da3   "/bin/bash -c 'sleep…"   About a minute ago   Up About a minute             foo-production
5150bb3a0d0f   a50ab9f16797   "/bin/bash -c 'sleep…"   About a minute ago   Up About a minute             foo-staging
```
- Check the output files.

```sh
# staging
$ cat staging/generated/main.tf
provider "docker" {
  host = "unix:///var/run/docker.sock"
}
resource "docker_image" "ubuntu" {
  name = "ubuntu:latest"
}
resource "docker_container" "foo" {
  image   = docker_image.ubuntu.image_id
  name    = "foo-staging"
  command = ["/bin/bash", "-c", "sleep 100"]
}
terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

# production
$ cat production/generated/main.tf
provider "docker" {
  host = "unix:///var/run/docker.sock"
}
resource "docker_image" "ubuntu" {
  name = "ubuntu:24.04"
}
resource "docker_container" "foo" {
  command = ["/bin/bash", "-c", "sleep 100"]
  image   = docker_image.ubuntu.image_id
  name    = "foo-production"
}
terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}
```