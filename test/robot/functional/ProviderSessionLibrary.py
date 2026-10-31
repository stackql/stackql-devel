"""Session configuration through the CLI and a loopback-only provider."""
import json
import shutil
import subprocess
import tempfile
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse


class ProviderSessionLibrary:
    ROBOT_LIBRARY_SCOPE = "SUITE"

    def prepare_session_provider(self, binary):
        self.binary = str(Path(binary).resolve())
        self.root = Path(tempfile.mkdtemp(prefix="stackql-session-"))
        self.calls = []
        owner = self

        class API(BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass

            def do_GET(self):
                token = self.headers.get("Authorization", "")
                region = parse_qs(urlparse(self.path).query).get("region", [""])[0]
                owner.calls.append((token, region))
                payload = json.dumps([{"id": region or "account"}]).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(payload)))
                self.end_headers()
                self.wfile.write(payload)

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), API)
        self.worker = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.worker.start()
        endpoint = f"http://127.0.0.1:{self.server.server_port}"
        provider = {"id": "widgets", "name": "widgets", "version": "v00.00.00000",
                    "providerServices": {"things": {"id": "things:v00.00.00000", "name": "things", "preferred": True,
                                                     "service": {"$ref": "widgets/v00.00.00000/services/things.yaml"}}},
                    "config": {"auth": {"type": "bearer"}}}
        doc = {"openapi": "3.0.0", "info": {"title": "things", "version": "v1"}, "servers": [{"url": endpoint}],
               "paths": {"/widgets": {"get": {"operationId": "list_widgets",
                         "parameters": [{"name": "region", "in": "query", "required": True, "schema": {"type": "string"}}],
                         "responses": {"200": {"content": {"application/json": {"schema": {"type": "array", "items": {"type": "object", "properties": {"id": {"type": "string"}}}}}}}}}}},
               "components": {"x-stackQL-resources": {"widgets": {"id": "widgets.things.widgets", "name": "widgets",
                              "methods": {"list": {"operation": {"$ref": "#/paths/~1widgets/get"}, "response": {"mediaType": "application/json", "openAPIDocKey": "200"}}},
                              "sqlVerbs": {"select": [{"$ref": "#/components/x-stackQL-resources/widgets/methods/list"}]}}}}}
        registry = self.root / "registry"
        version = registry / "src/widgets/v00.00.00000"
        (version / "services").mkdir(parents=True)
        (version / "provider.yaml").write_text(json.dumps(provider))
        (version / "services/things.yaml").write_text(json.dumps(doc))
        self.flags = ["--approot", str(self.root / "app"), "--registry", json.dumps({"localDocRoot": str(registry), "verifyConfig": {"nopVerify": True}}),
                      "--preview", json.dumps({"unstable": True, "endpoint": endpoint}), "--output", "json"]

    @staticmethod
    def setting(path, value, scope=""):
        raw = value if isinstance(value, str) else json.dumps(value)
        return f'SET {scope} "{path}" = ' + "'" + raw.replace("'", "''") + "';"

    def run_sql(self, sql, expected_calls):
        before = len(self.calls)
        result = subprocess.run([self.binary, "exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
        assert result.returncode == 0, result.stderr
        assert not result.stderr.strip(), result.stderr
        assert self.calls[before:] == expected_calls, (self.calls[before:], result.stdout, result.stderr)
        assert "east-token" not in result.stdout and "west-token" not in result.stdout, "credential leaked"
        return result

    def provider_aliases_survive_session_edits(self):
        sql = ""
        for name in ["east", "west"]:
            sql += self.setting("providers." + name, {"provider": "widgets", "tenancy": name, "auth": {"type": "bearer", "credentials": name + "-token"}, "params": {"region": name}, "tuning": {"max_concurrent": 1}})
        sql += "SHOW SERVICES IN east; SHOW RESOURCES IN east.things; SHOW METHODS IN east.things.widgets; DESCRIBE east.things.widgets; SELECT * FROM east.things.widgets; SELECT id FROM west.things.widgets;"
        sql += self.setting("providers.east.auth.credentials", "changed-token")
        sql += "SELECT id FROM east.things.widgets WHERE region = 'override';"
        self.run_sql(sql, [("Bearer east-token", "east"), ("Bearer west-token", "west"), ("Bearer changed-token", "override")])

    def secret_credentials_rotate_and_reset(self):
        sql = self.setting("secret_store", {"path": str(self.root / "secrets.json")})
        sql += self.setting("secrets.account", json.dumps({"type": "bearer", "credentials": "first-token"}))
        sql += self.setting("providers.secure", {"provider": "widgets", "tenancy": "secure-account", "auth_secret": "account", "params": {"region": "secure"}})
        sql += "SELECT id FROM secure.things.widgets;"
        sql += self.setting("secrets.account", json.dumps({"type": "bearer", "credentials": "rotated-token"}))
        sql += "SELECT id FROM secure.things.widgets; UNSET \"providers.secure\"; UNSET \"secrets.account\";"
        self.run_sql(sql, [("Bearer first-token", "secure"), ("Bearer rotated-token", "secure")])
        assert json.loads((self.root / "secrets.json").read_text()) == {}, "secret was not deleted"

    def auth_dto_changes_and_unsets(self):
        sql = self.setting("auth.widgets", {"type": "api_key", "credentialsenvvar": "STACKQL_UNUSED_CREDENTIAL"})
        sql += 'UNSET "auth.widgets.credentialsenvvar";'
        sql += self.setting("auth.widgets", {"type": "null_auth"})
        sql += "SELECT id FROM widgets.things.widgets WHERE region = 'plain';"
        self.run_sql(sql, [("", "plain")])

    def unset_alias_auth_fails_before_request(self):
        sql = self.setting("providers.disabled", {"provider": "widgets", "tenancy": "disabled-account", "auth": {"type": "bearer", "credentials": "explicit-token"}, "params": {"region": "disabled"}})
        sql += 'UNSET "providers.disabled.auth"; SELECT id FROM disabled.things.widgets;'
        before = len(self.calls)
        result = subprocess.run([self.binary, "exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
        assert "requires resolved explicit credentials" in result.stderr, result.stderr
        assert len(self.calls) == before, "missing auth reached the provider"

    def stop_session_provider(self):
        if hasattr(self, "server"):
            self.server.shutdown()
            self.server.server_close()
            self.worker.join(timeout=5)
        if hasattr(self, "root"):
            shutil.rmtree(self.root)

    def explicit_session_scope(self):
        binding = {"provider":"widgets", "tenancy":"scoped", "auth":{"type":"bearer","credentials":"scoped-token"}, "params":{"region":"scoped"}}
        sql = self.setting("providers.scoped", binding, "SESSION")
        sql += "SELECT id FROM scoped.things.widgets; UNSET SESSION \"providers.scoped\", \"auth.widgets.password\";"
        self.run_sql(sql, [("Bearer scoped-token", "scoped")])

    def unsupported_configuration_scope(self):
        for statement in ['SET GLOBAL "auth.widgets.password" = \'value\';', 'UNSET GLOBAL "auth.widgets.password";']:
            before = len(self.calls)
            result = subprocess.run([self.binary, "exec", statement] + self.flags, text=True, capture_output=True, timeout=30)
            assert result.returncode != 0 and "not supported" in result.stderr, result.stderr
            assert len(self.calls) == before, "unsupported scope reached provider"

    def dollar_assignment_target_is_rejected(self):
        before = len(self.calls)
        result = subprocess.run([self.binary, "exec", 'SET "$.auth.widgets.password" = \'value\';'] + self.flags, text=True, capture_output=True, timeout=30)
        assert result.returncode != 0 and "targets must omit" in result.stderr, result.stderr
        assert len(self.calls) == before, "invalid assignment target reached provider"
