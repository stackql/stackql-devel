package dto

import (
	"gopkg.in/yaml.v2"
)

type PGInternalCfg struct {
	IsEager bool `json:"isEager" yaml:"isEager"`
}

func GetPGInternalCfg(s string) (PGInternalCfg, error) {
	rv := PGInternalCfg{}
	err := yaml.Unmarshal([]byte(s), &rv)
	return rv, err
}
