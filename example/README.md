# example

This document provides instructions for generating Terraform configuration files using a tool called `tfustomize`, and then applying those configurations to manage infrastructure in two environments: `staging` and `production`. These instructions use [docker provider](https://registry.terraform.io/providers/kreuzwerker/docker/latest) so `docker` is required in your environment.

- Generate the Terraform configuration file using `tfustomize`.

```sh
mkdir -p {staging/production}/generated

tfustomize build staging | tee staging/generated/main.tf
tfustomize build production | tee production/generated/main.tf
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
 $ docker image ls | grep ubuntu
ubuntu                                                                                  latest                                     a50ab9f16797   4 days ago      69.2MB
ubuntu                                                                                  24.04                                      1e6914c26da3   5 days ago      99.6MB

$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS              PORTS     NAMES
17cf2b1cb30e   1e6914c26da3   "/bin/bash -c 'sleep…"   About a minute ago   Up About a minute             foo-production
5150bb3a0d0f   a50ab9f16797   "/bin/bash -c 'sleep…"   About a minute ago   Up About a minute             foo-staging
```
