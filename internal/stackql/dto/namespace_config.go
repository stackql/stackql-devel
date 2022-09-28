package dto

import "gopkg.in/yaml.v2"

type NamespaceCfg struct {
	RegexpStr         string `json:"regex" yaml:"regex"`
	TTL               int    `json:"ttl" yaml:"ttl"`
	NamespaceTemplate string `json:"template" yaml:"template"`
}

func GetNamespaceCfg(s string) (*NamespaceCfg, error) {
	rv := &NamespaceCfg{}
	err := yaml.Unmarshal([]byte(s), rv)
	return rv, err
}
