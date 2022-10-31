package dto

import (
	"github.com/stackql/stackql/internal/stackql/constants"
	"gopkg.in/yaml.v2"
)

type SQLBackendCfg struct {
	DbEngine       string `json:"dbEngine" yaml:"dbEngine"`
	DbFilePath     string `json:"dbFilepath" yaml:"dbFilepath"`
	DbInitFilePath string `json:"dbInitFilepath" yaml:"dbInitFilepath"`
	SQLDialect     string `json:"sqlDialect" yaml:"sqlDialect"`
}

func GetSQLBackendCfg(s string) (SQLBackendCfg, error) {
	rv := SQLBackendCfg{}
	err := yaml.Unmarshal([]byte(s), &rv)
	if rv.DbEngine == "" {
		rv.DbEngine = constants.DefaultDbEngine
	}
	if rv.SQLDialect == "" {
		rv.SQLDialect = constants.DefaultSQLDialect
	}
	return rv, err
}
