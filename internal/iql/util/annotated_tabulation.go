package util

import (
	"infraql/internal/iql/dto"

	"infraql/internal/pkg/openapistackql"
)

type AnnotatedTabulation struct {
	tab  *openapistackql.Tabulation
	hIds *dto.HeirarchyIdentifiers
}

func NewAnnotatedTabulation(tab *openapistackql.Tabulation, hIds *dto.HeirarchyIdentifiers) AnnotatedTabulation {
	return AnnotatedTabulation{
		tab:  tab,
		hIds: hIds,
	}
}

func (at AnnotatedTabulation) GetTabulation() *openapistackql.Tabulation {
	return at.tab
}

func (at AnnotatedTabulation) GetHeirarchyIdentifiers() *dto.HeirarchyIdentifiers {
	return at.hIds
}
