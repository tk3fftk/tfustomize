tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "../staging/main.tf",
  ]
}

patches {
  paths = [
    "main.tf",
  ]
}
