package openapistackql

import (
	"embed"
	"fmt"
)

//go:embed embeddedproviders/googleapis.com/* embeddedproviders/googleapis.com/bigquery/* embeddedproviders/googleapis.com/cloudresourcemanager/* embeddedproviders/googleapis.com/compute/* embeddedproviders/googleapis.com/container/*
var googleProvider embed.FS

//go:embed embeddedproviders/okta/* embeddedproviders/okta/*/*
var oktaProvider embed.FS

func GetEmbeddedProvider(prov string) (embed.FS, error) {
	switch prov {
	case "google":
		return googleProvider, nil
	case "okta":
		return oktaProvider, nil
	}
	return embed.FS{}, fmt.Errorf("no such embedded provider: '%s'", prov)
}
