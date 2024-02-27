tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "../base/main.tf",
    "../base/provider.tf",
  ]
}

patches {
  paths = [
    "./overlay.tf",
    "./provider.tf",
  ]
}
