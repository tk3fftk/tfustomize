# Pulls the image
resource "docker_image" "ubuntu" {
  name = "ubuntu:24.04"
}

# Create a container
resource "docker_container" "foo" {
  name  = "foo-production"
}
