tfustomize {
  syntax_version = "v1"
}

resources {
  pathes = [
    "../staging/main.tf",
  ]
}

patches {
  pathes = [
    "main.tf",
  ]
}
