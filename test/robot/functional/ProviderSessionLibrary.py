"""Session configuration through the CLI and a loopback-only provider."""
import json
import os
import shutil
import subprocess
import tempfile
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse


class ProviderSessionLibrary:
    ROBOT_LIBRARY_SCOPE = "SUITE"

    def prepare_session_provider(self, binary, execution_platform="native"):
        self.prefix = [shutil.which(binary) or str(Path(binary).resolve())]
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

        self.server = ThreadingHTTPServer(("0.0.0.0" if execution_platform == "docker" else "127.0.0.1", 0), API)
        self.worker = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.worker.start()
        endpoint = f"http://127.0.0.1:{self.server.server_port}"
        if execution_platform == "docker":
            self.prefix = ["docker", "--log-level", "error", "compose", "-p", "execrun", "run", "--rm", "--no-deps", "-T",
                           "--user", f"{os.getuid()}:{os.getgid()}", "--workdir", str(self.root),
                           "-v", str(self.root) + ":" + str(self.root), "--entrypoint", "stackql", "stackqlsrv"]
            endpoint = endpoint.replace("127.0.0.1", "host.docker.internal")
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
        result = subprocess.run(self.prefix + ["exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
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
        sql = self.setting("secret_store", {"path": (self.root / "secrets.json").as_posix()})
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
        result = subprocess.run(self.prefix + ["exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
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
            result = subprocess.run(self.prefix + ["exec", statement] + self.flags, text=True, capture_output=True, timeout=30)
            assert result.returncode != 0 and "not supported" in result.stderr, result.stderr
            assert len(self.calls) == before, "unsupported scope reached provider"

    def dollar_assignment_target_is_rejected(self):
        before = len(self.calls)
        result = subprocess.run(self.prefix + ["exec", 'SET "$.auth.widgets.password" = \'value\';'] + self.flags, text=True, capture_output=True, timeout=30)
        assert result.returncode != 0 and "targets must omit" in result.stderr, result.stderr
        assert len(self.calls) == before, "invalid assignment target reached provider"

    def scalar_variables_preserve_sql_expressions(self):
        sql = '''SET "variables.word" = 'hello'; SET "variables.number" = 9007199254740993;
                 SET "variables.yes" = true; SET "variables.empty" = null; SET "variables.negative" = -17;
                 SELECT upper($word) || '-world' AS label, $number AS exact_number,
                        $yes = true AS truth, $empty IS NULL AS is_null, $negative AS negative;
                 SET "variables.word" = 'changed'; SELECT upper($word) AS changed;'''
        result = self.run_sql(sql, [])
        for expected in ['HELLO-world', '9007199254740993', 'CHANGED', '-17']:
            assert expected in result.stdout, result.stdout

    def variable_predicate_resolves_before_provider_dispatch(self):
        sql = self.setting("providers.variable", {"provider": "widgets", "tenancy": "variable", "auth": {"type": "bearer", "credentials": "variable-token"}, "params": {"region": "default"}})
        sql += '''SET "variables.region" = 'east'; SELECT id FROM variable.things.widgets WHERE region = upper($region);'''
        self.run_sql(sql, [("Bearer variable-token", "EAST")])

    def variable_errors_prevent_provider_dispatch(self):
        for expression in ['$missing', '"$providers.missing"', '"$auth.widgets.password"']:
            before = len(self.calls)
            result = subprocess.run(self.prefix + ["exec", 'SELECT ' + expression + ';'] + self.flags, text=True, capture_output=True, timeout=30)
            assert result.returncode != 0 and result.stderr.strip(), (result.stdout, result.stderr)
            assert len(self.calls) == before, "invalid variable reached provider"
        sql = '''SET "variables.removed" = 'old'; UNSET "variables.removed"; SELECT $removed;'''
        result = subprocess.run(self.prefix + ["exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
        assert result.returncode != 0 and 'not defined' in result.stderr, result.stderr

    def secret_variables_require_runtime_opt_in(self):
        sql = self.setting("secret_store", {"path": (self.root / "dereference-secrets.json").as_posix()})
        sql += self.setting("secrets.demo", "private-demo-value")
        sql += 'SELECT upper("$secrets.demo") AS value;'
        before = len(self.calls)
        denied = subprocess.run(self.prefix + ["exec", sql] + self.flags, text=True, capture_output=True, timeout=30)
        assert denied.returncode != 0 and 'secret variable dereferencing is disabled' in denied.stderr, denied.stderr
        assert 'private-demo-value' not in denied.stdout + denied.stderr, "secret leaked without opt-in"
        allowed = subprocess.run(self.prefix + ["exec", sql, "--allow-secret-dereference"] + self.flags, text=True, capture_output=True, timeout=30)
        assert allowed.returncode == 0 and 'PRIVATE-DEMO-VALUE' in allowed.stdout, allowed.stdout + allowed.stderr
        assert len(self.calls) == before, "secret expression reached provider"
