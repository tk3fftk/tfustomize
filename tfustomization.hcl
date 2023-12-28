tfustomize {
  syntax_version = "v1"
}

resources {
  pathes = [
    "./base.tf",
  ]
}

patches {
  pathes = [
    "./overlay.tf",
  ]
}
