package locking

import (
	"sync"
)

type LockerCfg struct {
	ProviderName string `json:"providerName" yaml:"providerName"`
}

type ResourceLocker interface{}

func GetReourceLocker(cfg LockerCfg) (ResourceLocker, error) {
	return newProviderResourceLocker(cfg.ProviderName), nil
}

type providerResourceLocker struct {
	m            sync.Mutex
	providerName string
}

func newProviderResourceLocker(providerName string) ResourceLocker {
	return &providerResourceLocker{
		providerName: providerName,
	}
}
