# tfustomize

Customization of Terraform HCL

- Terraform directory structure

```bash
~/someApp
├── base
│   ├── backend.tf
│   ├── main.tf
│   ├── outputs.tf
│   ├── tfustomization.hcl
│   └── variables.tf
└── overlays
    ├── dev
    │   ├── backend.tf
    │   ├── main.tf
    │   ├── outputs.tf
    │   ├── tfustomization.hcl
    │   └── variables.tf
    └── prod
        ├── backend.tf
        ├── main.tf
        ├── outputs.tf
        ├── tfustomization.hcl
        └── variables.tf
```

- `tfustomization.hcl`

```hcl
tfustomize {
  syntax_version = "v1"
}

resources {
  paths = [
    "../../base/main.tf",
  ]
}

patches {
  paths = [
    "./main.tf",
  ]
}
```
