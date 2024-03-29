package api

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type TfustomizeConfig struct {
	Tfustomize Tfustomize `hcl:"tfustomize,block"`
	Resources  Resource   `hcl:"resources,block"`
	Patches    Patch      `hcl:"patches,block"`
}

type Tfustomize struct {
	SyntaxVersion string `hcl:"syntax_version,optional"`
}

type Resource struct {
	Paths []string `hcl:"paths,attr"`
}

type Patch struct {
	Paths []string `hcl:"paths,attr"`
}

func LoadConfig(configPath string) (TfustomizeConfig, error) {
	return decodeConfigFromFile(configPath)
}

func decodeConfigFromFile(path string) (tfusconf TfustomizeConfig, err error) {
	err = hclsimple.DecodeFile(path, nil, &tfusconf)
	if err != nil {
		return
	}
	return
}
