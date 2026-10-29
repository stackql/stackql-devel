package accessmanagement

import (
	"context"

	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stackql-labs/omnisdk/public/authentication"
	"github.com/stackql-labs/omnisdk/public/authorization"
	"github.com/stackql-labs/omnisdk/public/principal"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

// Config is host-owned configuration, never accepted from SQL or provider predicates.
type Config struct {
	Issuer            string               `json:"issuer"`
	Audience          string               `json:"audience"`
	Keys              []VerificationKey    `json:"keys"`
	Rules             []authorization.Rule `json:"rules"`
	MaxSessionSeconds int                  `json:"max_session_seconds"`
}
type VerificationKey struct {
	ID        string `json:"kid"`
	Algorithm string `json:"algorithm"`
	PEM       string `json:"public_key_pem"`
}

// Manager establishes a fresh directory snapshot for every statement.
type Manager interface {
	Establish(context.Context) (authorization.Session, context.Context, context.CancelFunc, error)
}
type manager struct {
	authenticator authentication.Authenticator
	directory     principal.Directory
	authorizer    authorization.Authorizer
	lifetime      time.Duration
}

const defaultSessionLifetime = 5 * time.Minute

type evidenceKey struct{}

// WithEvidence marks a transport boundary, including when no credential was supplied.
// Remote transports must never fall back to the server process's environment token.
func WithEvidence(ctx context.Context, evidence authentication.Evidence) context.Context {
	return context.WithValue(ctx, evidenceKey{}, evidence)
}

//nolint:nilnil // No configuration explicitly disables caller protection.
func NewManager(path string, engine sqlengine.SQLEngine) (Manager, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("cannot read caller access configuration")
	}
	var cfg Config
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&cfg); err != nil {
		return nil, errors.New("invalid caller access configuration")
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("invalid trailing caller access configuration")
	}
	directory, err := NewDirectory(engine)
	if err != nil {
		return nil, err
	}
	return newManager(cfg, directory)
}

func newManager(cfg Config, directory principal.Directory) (Manager, error) {
	if len(cfg.Keys) == 0 || len(cfg.Rules) == 0 || (cfg.MaxSessionSeconds < 0 || cfg.MaxSessionSeconds > 86400) {
		return nil, errors.New("caller access requires trusted keys, policy and a valid session lifetime")
	}
	keys := map[string]any{}
	algorithms := map[string]string{}
	var allowed []string
	for _, spec := range cfg.Keys {
		if spec.ID == "" || keys[spec.ID] != nil {
			return nil, errors.New("verification keys require unique nonempty IDs")
		}
		var key any
		var err error
		switch spec.Algorithm {
		case "EdDSA":
			key, err = jwt.ParseEdPublicKeyFromPEM([]byte(spec.PEM))
		case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
			key, err = jwt.ParseRSAPublicKeyFromPEM([]byte(spec.PEM))
		case "ES256", "ES384", "ES512":
			key, err = jwt.ParseECPublicKeyFromPEM([]byte(spec.PEM))
		default:
			return nil, errors.New("unsupported caller JWT algorithm")
		}
		if err != nil {
			return nil, errors.New("invalid caller verification key")
		}
		if value, ok := key.(*rsa.PublicKey); ok && value.N.BitLen() < 2048 {
			return nil, errors.New("RSA key must have at least 2048 bits")
		}

		keys[spec.ID], algorithms[spec.ID] = key, spec.Algorithm
		allowed = append(allowed, spec.Algorithm)
	}
	verifier, err := authentication.NewJWT(cfg.Issuer, cfg.Audience, allowed, func(token *jwt.Token) (any, error) {
		id, ok := token.Header["kid"].(string)
		if !ok || keys[id] == nil || algorithms[id] != token.Method.Alg() {
			return nil, errors.New("untrusted caller verification key")
		}
		return keys[id], nil
	})
	if err != nil {
		return nil, err
	}
	policy, err := authorization.NewCasbin(cfg.Rules)
	if err != nil {
		return nil, err
	}
	lifetime := time.Duration(cfg.MaxSessionSeconds) * time.Second
	if lifetime == 0 {
		lifetime = defaultSessionLifetime
	}
	return &manager{verifier, directory, policy, lifetime}, nil
}

func (m *manager) Establish(ctx context.Context) (authorization.Session, context.Context, context.CancelFunc, error) {
	evidence, remote := ctx.Value(evidenceKey{}).(authentication.Evidence)
	if !remote {
		evidence.Token = os.Getenv("STACKQL_CALLER_TOKEN")
	}
	identity, err := m.authenticator.Authenticate(ctx, evidence)
	if err != nil {
		return nil, nil, nil, errors.New("caller authentication failed")
	}
	// Only inspect expiry after the maintained verifier has authenticated this exact token.
	claims := jwt.RegisteredClaims{}
	if _, _, err = jwt.NewParser().ParseUnverified(evidence.Token, &claims); err != nil || claims.ExpiresAt == nil {
		return nil, nil, nil, errors.New("caller expiry is required")
	}
	deadline := time.Now().Add(m.lifetime)
	if claims.ExpiresAt.Time.Before(deadline) {
		deadline = claims.ExpiresAt.Time
	}
	bounded, cancel := context.WithDeadline(ctx, deadline)
	caller, err := authorization.NewSession(bounded, identity, m.directory, m.authorizer)
	if err != nil {
		cancel()
		return nil, nil, nil, errors.New("caller identity resolution failed")
	}
	return caller, caller.Context(bounded), cancel, nil
}
