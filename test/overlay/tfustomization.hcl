tfustomize {
  syntax_version = "v1"
}

resources {
  pathes = [
    "../base/main.tf",
    "../base/provider.tf",
  ]
}

patches {
  pathes = [
    "./overlay.tf",
    "./provider.tf",
  ]
}
