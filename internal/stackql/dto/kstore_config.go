package dto

import (
	"gopkg.in/yaml.v2"
)

type KStoreCfg struct {
	IsEager bool `json:"isEager" yaml:"isEager"`
}

func GetKStoreCfg(s string) (KStoreCfg, error) {
	rv := KStoreCfg{}
	err := yaml.Unmarshal([]byte(s), &rv)
	return rv, err
}
