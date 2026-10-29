package accessmanagement //nolint:testpackage // Exercises the private directory implementation.

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stackql-labs/omnisdk/public/authentication"
	"github.com/stackql-labs/omnisdk/public/authorization"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

//nolint:gocognit // Sequential session lifecycle assertions share an isolated directory fixture.
func TestCallerSessionsRefreshAndFailClosed(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(pub)
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := initializeSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`INSERT INTO stackql_access_principals VALUES ('alice','{}')`,
		`INSERT INTO stackql_access_groups VALUES ('readers')`,
		`INSERT INTO stackql_access_identities VALUES ('issuer','alice','alice')`,
		`INSERT INTO stackql_access_memberships VALUES ('alice','readers')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	directory, _ := newDirectory(db)
	m, err := newManager(Config{Issuer: "issuer", Audience: "stackql", Keys: []VerificationKey{{ID: "test", Algorithm: "EdDSA", PEM: string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))}}, Rules: []authorization.Rule{{Subject: "group:readers", Resource: "operation/demo/list/listItems", Action: "execute"}}}, directory)
	if err != nil {
		t.Fatal(err)
	}
	sign := func(modify func(jwt.MapClaims), kid string, signingKey any) string {
		claims := jwt.MapClaims{"iss": "issuer", "sub": "alice", "aud": "stackql", "exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(), "groups": []string{"admins"}}
		if modify != nil {
			modify(claims)
		}
		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		token.Header["kid"] = kid
		value, err := token.SignedString(signingKey)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	establish := func(token string) (authorization.Session, error) {
		caller, _, cancel, err := m.Establish(WithEvidence(context.Background(), authentication.Evidence{Token: token}))
		if cancel != nil {
			t.Cleanup(cancel)
		}
		return caller, err
	}
	valid := sign(nil, "test", key)
	caller, err := establish(valid)
	if err != nil {
		t.Fatal(err)
	}
	if len(caller.Principal().Groups()) != 1 || caller.Principal().Groups()[0] != "readers" {
		t.Fatal("JWT groups replaced directory membership")
	}
	for _, query := range []string{"SELECT 1", "SELECT count(*) FROM demo.service.items"} {
		stmt, _ := sqlparser.Parse(query)
		if err := CheckStatement(caller.Context(context.Background()), caller, stmt); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"DELETE FROM demo.service.items", "SELECT * FROM stackql_access_principals", "SELECT pg_read_file('/tmp/x')", "SELECT * FROM stackql_preview.iac.run"} {
		stmt, err := sqlparser.Parse(query)
		if err != nil {
			t.Fatal(err)
		}
		if CheckStatement(caller.Context(context.Background()), caller, stmt) == nil {
			t.Fatalf("allowed %s", query)
		}
	}
	_, other, _ := ed25519.GenerateKey(rand.Reader)
	for name, token := range map[string]string{
		"missing": "", "signature": sign(nil, "test", other), "unknown key": sign(nil, "other", key),
		"issuer":     sign(func(c jwt.MapClaims) { c["iss"] = "other" }, "test", key),
		"audience":   sign(func(c jwt.MapClaims) { c["aud"] = "other" }, "test", key),
		"expiry":     sign(func(c jwt.MapClaims) { delete(c, "exp") }, "test", key),
		"expired":    sign(func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Minute).Unix() }, "test", key),
		"future":     sign(func(c jwt.MapClaims) { c["iat"] = time.Now().Add(time.Hour).Unix() }, "test", key),
		"not before": sign(func(c jwt.MapClaims) { c["nbf"] = time.Now().Add(time.Hour).Unix() }, "test", key),
	} {
		if _, err := establish(token); err == nil {
			t.Fatalf("accepted %s", name)
		}
	}
	if _, err := db.Exec(`DELETE FROM stackql_access_memberships`); err != nil {
		t.Fatal(err)
	}
	next, err := establish(valid)
	if err != nil {
		t.Fatal(err)
	}
	stmt, _ := sqlparser.Parse("SELECT 1")
	if CheckStatement(next.Context(context.Background()), next, stmt) == nil {
		t.Fatal("revoked membership retained")
	}
	if len(caller.Principal().Groups()) != 1 {
		t.Fatal("existing snapshot mutated")
	}
	if _, err := db.Exec(`UPDATE stackql_access_principals SET attributes = 'null'`); err != nil {
		t.Fatal(err)
	}
	if _, err := establish(valid); err == nil {
		t.Fatal("accepted malformed attributes")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := establish(valid); err == nil {
		t.Fatal("directory outage allowed access")
	}
}
