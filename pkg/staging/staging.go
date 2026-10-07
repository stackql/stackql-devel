package staging

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	queryIDSize  = 16
	nameHashSize = 16
	maxSlugSize  = 24
)

// Scope assigns isolated physical table names to the logical relations staged
// during one query. Names remains useful to the caller for query cleanup.
type Scope interface {
	TableName(logicalName string) (string, error)
	Names() []string
}

type queryScope struct {
	queryID string
	mu      sync.Mutex
	names   map[string]string
}

// NewScope creates a staging namespace for a single query.
func NewScope() (Scope, error) {
	queryID := make([]byte, queryIDSize)
	if _, err := rand.Read(queryID); err != nil {
		return nil, fmt.Errorf("staging: generate query namespace: %w", err)
	}
	return &queryScope{
		queryID: hex.EncodeToString(queryID),
		names:   make(map[string]string),
	}, nil
}

func (s *queryScope) TableName(logicalName string) (string, error) {
	if logicalName == "" {
		return "", fmt.Errorf("staging: logical table name must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if name, ok := s.names[logicalName]; ok {
		return name, nil
	}

	digest := sha256.Sum256([]byte(logicalName))
	name := fmt.Sprintf("__omni_stage_%s_%s_%s",
		s.queryID,
		nameSlug(logicalName),
		hex.EncodeToString(digest[:nameHashSize]),
	)
	s.names[logicalName] = name
	return name, nil
}

func (s *queryScope) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.names))
	for _, name := range s.names {
		names = append(names, name)
	}
	// Allocation order is intentionally hidden; stable ordering simplifies cleanup.
	sort.Strings(names)
	return names
}

func nameSlug(logicalName string) string {
	var slug strings.Builder
	slug.Grow(maxSlugSize)
	separator := false
	for _, char := range strings.ToLower(logicalName) {
		if slug.Len() == maxSlugSize {
			break
		}
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			slug.WriteRune(char)
			separator = false
		} else if slug.Len() > 0 && !separator {
			slug.WriteByte('_')
			separator = true
		}
	}
	return strings.Trim(slug.String(), "_")
}
