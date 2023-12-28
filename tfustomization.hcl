tfustomize {
  syntax_version = "v1"
}

resources {
  pathes = [
    "./base.hcl",
  ]
}

patches {
  pathes = [
    "./overlay.hcl",
  ]
}
