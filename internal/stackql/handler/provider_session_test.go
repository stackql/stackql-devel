package handler

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stackql-labs/omnisdk/pkg/omnisdk"
	dto "github.com/stackql/stackql/internal/stackql/runtimeconfig"
)

func sessionTestHandler() *standardHandlerContext {
	return &standardHandlerContext{authMapMutex: &sync.Mutex{}, sessionCtxMutex: &sync.Mutex{}, authContexts: dto.AuthContexts{}, executionContext: context.Background()}
}

func TestProviderSessionBindings(t *testing.T) {
	hc := sessionTestHandler()
	binding := map[string]interface{}{"provider": "widgets", "tenancy": "account-a", "auth": map[string]interface{}{"type": "bearer", "credentials": "first"}, "tuning": map[string]interface{}{"max_concurrent": 1}}
	if err := hc.SetConfigAtPath("providers.production", binding, ""); err != nil {
		t.Fatal(err)
	}
	snapshot, err := hc.ProviderSessionArgs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ProviderAliases["production"].Auth.Credentials != "first" {
		t.Fatal("wrong credential")
	}
	if snapshot.ProviderAliases["stackql_unstable_production"].Auth.Credentials != "first" {
		t.Fatal("SQL alias missing")
	}
	if got := hc.ResolveProviderAddress("stackql_unstable_production.things.widgets"); got != "stackql_unstable_widgets.things.widgets" {
		t.Fatal(got)
	}
	if err := hc.SetConfigAtPath("providers.production.auth.credentials", "second", "session"); err != nil {
		t.Fatal(err)
	}
	if snapshot.ProviderAliases["production"].Auth.Credentials != "first" {
		t.Fatal("snapshot mutated")
	}
	if err := hc.SetConfigAtPath("providers.production.typo", "value", ""); err == nil {
		t.Fatal("unknown field accepted")
	}
	if err := hc.SetConfigAtPath("providers.production.auth", nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := hc.ProviderSessionArgs(context.Background()); err == nil {
		t.Fatal("unset auth borrowed credentials")
	}
	if err := hc.SetConfigAtPath("providers.production", nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := hc.ProviderSessionArgs(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestSessionSecretStore(t *testing.T) {
	hc := sessionTestHandler()
	cfg := map[string]interface{}{"path": filepath.Join(t.TempDir(), "secrets.json")}
	if err := hc.SetConfigAtPath("secret_store", cfg, ""); err != nil {
		t.Fatal(err)
	}
	auth := `{"type":"bearer","credentials":"from-store"}`
	if err := hc.SetConfigAtPath("secrets.production", auth, ""); err != nil {
		t.Fatal(err)
	}
	binding := map[string]interface{}{"provider": "widgets", "tenancy": "account-a", "auth_secret": "production"}
	if err := hc.SetConfigAtPath("providers.production", binding, ""); err != nil {
		t.Fatal(err)
	}
	snapshot, err := hc.ProviderSessionArgs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.ProviderAliases["production"].Auth.Credentials != "from-store" {
		t.Fatal("secret not resolved")
	}
	if err := hc.SetConfigAtPath("secrets.production", nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := hc.ProviderSessionArgs(context.Background()); err == nil {
		t.Fatal("deleted secret still available")
	}
}

func TestAuthSessionResetAndSnapshots(t *testing.T) {
	hc := sessionTestHandler()
	hc.authContexts["aws"] = &dto.AuthCtx{AuthConfig: omnisdk.AuthConfig{Type: "basic", Username: "user", Password: "initial"}}
	before, err := hc.GetAuthContext("aws")
	if err != nil {
		t.Fatal(err)
	}
	if err := hc.SetConfigAtPath("auth.aws.password", "updated", ""); err != nil {
		t.Fatal(err)
	}
	if before.Password != "initial" {
		t.Fatal("auth snapshot mutated")
	}
	if err := hc.SetConfigAtPath("auth.aws.password", nil, ""); err != nil {
		t.Fatal(err)
	}
	after, _ := hc.GetAuthContext("aws")
	if after.Password != "" || after.Username != "user" {
		t.Fatal("incorrect field reset")
	}
	if err := hc.SetConfigAtPath("auth.aws.unknown", "bad", ""); err == nil {
		t.Fatal("unknown auth field accepted")
	}
	if err := hc.SetConfigAtPath("auth.aws", map[string]interface{}{"type": "basic", "username": "replacement"}, ""); err != nil {
		t.Fatal(err)
	}
	after, _ = hc.GetAuthContext("aws")
	if after.Username != "replacement" {
		t.Fatal("replacement not applied")
	}
}

func TestClientSessionsDoNotShareMutations(t *testing.T) {
	parent := sessionTestHandler()
	parent.ReplaceAuthContext("widgets", &dto.AuthCtx{AuthConfig: omnisdk.AuthConfig{Type: "basic", Password: "original"}})
	first := NewClientContext(parent).(*standardHandlerContext)
	second := NewClientContext(parent).(*standardHandlerContext)
	if err := first.SetConfigAtPath("auth.widgets.password", "changed", ""); err != nil {
		t.Fatal(err)
	}
	auth, err := second.GetAuthContext("widgets")
	if err != nil || auth.Password != "original" {
		t.Fatal("auth leaked across client sessions")
	}
	binding := map[string]interface{}{"provider": "widgets", "tenancy": "account", "auth": map[string]interface{}{"type": "bearer", "credentials": "token"}}
	if err := first.SetConfigAtPath("providers.private", binding, ""); err != nil {
		t.Fatal(err)
	}
	args, err := second.ProviderSessionArgs(context.Background())
	if err != nil || len(args.ProviderAliases) != 0 {
		t.Fatal("provider binding leaked across client sessions")
	}
	queryClone := first.Clone().(*standardHandlerContext)
	args, err = queryClone.ProviderSessionArgs(context.Background())
	if err != nil || args.ProviderAliases["private"].Auth == nil {
		t.Fatal("query clone lost its client session")
	}
}
