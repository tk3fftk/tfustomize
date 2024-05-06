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
