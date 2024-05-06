resource "docker_image" "ubuntu" {
  name = "ubuntu:24.04"
}

resource "docker_container" "foo" {
  name  = "foo-production"
}
