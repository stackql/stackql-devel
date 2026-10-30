"""Black-box caller access fixtures. Tokens stay in subprocess environments, never Robot results."""
import base64
import yaml
import concurrent.futures
import json
import datetime
import ipaddress
import socket
import shutil
import uuid
from urllib.parse import urlsplit, urlunsplit, parse_qsl, urlencode
import ssl
import urllib.request
import urllib.error
from cryptography import x509
from cryptography.x509.oid import NameOID
import os
import pathlib
import sqlite3
import subprocess
import tempfile
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import ed25519


class AccessManagementKeywords:
    ROBOT_LIBRARY_SCOPE = "SUITE"

    def verify_caller_access(self, executable, backend="{}", execution_platform="native"):
        with tempfile.TemporaryDirectory(prefix="stackql-access-") as directory:
            root = pathlib.Path(directory)
            docker = execution_platform == "docker"
            cfg = json.loads(backend)
            postgres = cfg.get("dbEngine") == "postgres_tcp"
            directory_schema = None
            if postgres:
                import psycopg
                dsn = cfg["dsn"]
                if docker:
                    subprocess.run(["docker", "compose", "-p", "execrun", "up", "-d", "postgres_stackql"], capture_output=True, check=True, timeout=120)
                    result = subprocess.run(["docker", "compose", "-p", "execrun", "port", "postgres_stackql", "5432"], capture_output=True, text=True, check=True, timeout=30)
                    published = result.stdout.strip().splitlines()[0].rsplit(":", 1)[1]
                    url = urlsplit(dsn)
                    credentials = url.netloc.rsplit("@", 1)[0]
                    dsn = urlunsplit((url.scheme, credentials + "@127.0.0.1:" + published, url.path, url.query, url.fragment))
                for attempt in range(50):
                    try:
                        db = psycopg.connect(dsn, autocommit=True)
                        break
                    except psycopg.OperationalError:
                        if not docker or attempt == 49:
                            raise
                        time.sleep(0.2)
                directory_schema = "stackql_access_test_" + uuid.uuid4().hex
                db.execute(psycopg.sql.SQL("CREATE SCHEMA {}").format(psycopg.sql.Identifier(directory_schema)))
                db.execute(psycopg.sql.SQL("SET search_path TO {}, public").format(psycopg.sql.Identifier(directory_schema)))
                url = urlsplit(cfg["dsn"])
                options = dict(parse_qsl(url.query))
                options["options"] = "-csearch_path=" + directory_schema + ",public"
                cfg["dsn"] = urlunsplit((url.scheme, url.netloc, url.path, urlencode(options), url.fragment))
            else:
                cfg = {"dsn": str(root / "access.db")}
                db = sqlite3.connect(cfg["dsn"], check_same_thread=False)
            key = ed25519.Ed25519PrivateKey.generate()
            public = key.public_key().public_bytes(serialization.Encoding.PEM, serialization.PublicFormat.SubjectPublicKeyInfo).decode()
            config = {"issuer": "access-test", "audience": "stackql", "keys": [{"kid": "local", "algorithm": "EdDSA", "public_key_pem": public}], "rules": [
                {"subject": "group:readers", "resource": "operation/accessdemo/list/listItems", "action": "execute"},
                {"subject": "group:readers", "resource": "credential/accessdemo", "action": "use"},
                {"subject": "group:operators", "resource": "**", "action": "*"},
            ]}
            config_file = root / "access.json"
            config_file.write_text(json.dumps(config))
            ddl = [
                "CREATE TABLE IF NOT EXISTS stackql_access_principals (id TEXT PRIMARY KEY, attributes TEXT NOT NULL DEFAULT '{}')",
                "CREATE TABLE IF NOT EXISTS stackql_access_groups (id TEXT PRIMARY KEY)",
                "CREATE TABLE IF NOT EXISTS stackql_access_identities (issuer TEXT NOT NULL, subject TEXT NOT NULL, principal_id TEXT NOT NULL REFERENCES stackql_access_principals(id), PRIMARY KEY (issuer,subject))",
                "CREATE TABLE IF NOT EXISTS stackql_access_memberships (principal_id TEXT NOT NULL REFERENCES stackql_access_principals(id), group_id TEXT NOT NULL REFERENCES stackql_access_groups(id), PRIMARY KEY (principal_id,group_id))",
            ]
            for statement in ddl:
                db.execute(statement)
            for table in ["memberships", "identities", "principals", "groups"]:
                db.execute("DELETE FROM stackql_access_" + table)
            for statement in [
                "INSERT INTO stackql_access_principals VALUES ('reader','{}'),('operator','{}'),('denied','{}')",
                "INSERT INTO stackql_access_groups VALUES ('readers'),('operators')",
                "INSERT INTO stackql_access_identities VALUES ('access-test','reader','reader'),('access-test','operator','operator'),('access-test','denied','denied')",
                "INSERT INTO stackql_access_memberships VALUES ('reader','readers'),('operator','operators')",
            ]:
                db.execute(statement)
            if not postgres:
                db.commit()
            counts = {"read": 0, "mutation": 0}
            managed = {}

            class API(BaseHTTPRequestHandler):
                def log_message(self, *_):
                    pass

                def do_GET(self):
                    counts["read"] += 1
                    self.send_response(200)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    identity = self.path.rsplit("/", 1)[-1]
                    self.wfile.write(json.dumps(managed.get(identity, {}) if self.path.startswith("/items/") else [{"id":"known-item"}]).encode())

                def do_POST(self):
                    counts["mutation"] += 1
                    body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", "0"))) or b"{}")
                    if body.get("name") == "upstream-reject":
                        self.send_response(400)
                        self.send_header("Content-Type", "application/json")
                        self.end_headers()
                        self.wfile.write(b'{"error":"test rejection"}')
                        return
                    self.send_response(200)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    identity = "managed-" + str(counts["mutation"])
                    managed[identity] = dict(body, id=identity)
                    self.wfile.write(json.dumps(managed[identity]).encode())

                def do_PATCH(self):
                    counts["mutation"] += 1
                    identity = self.path.rsplit("/", 1)[-1]
                    body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", "0"))) or b"{}")
                    managed[identity].update(body)
                    self.send_response(200)
                    self.send_header("Content-Type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps(managed[identity]).encode())

                def do_DELETE(self):
                    counts["mutation"] += 1
                    managed.pop(self.path.rsplit("/", 1)[-1], None)
                    self.send_response(204)
                    self.end_headers()

            server = ThreadingHTTPServer(("0.0.0.0" if docker else "127.0.0.1", 0), API)
            thread = threading.Thread(target=server.serve_forever, daemon=True)
            thread.start()
            doc_dir = root / "src" / "accessdemo" / "v0.0.1"
            doc_dir.mkdir(parents=True)
            (doc_dir / "provider.yaml").write_text("name: accessdemo\nproviderServices:\n  service:\n    service: {$ref: service.yaml}\n")
            (doc_dir / "service.yaml").write_text("""openapi: 3.0.0
servers: [{url: 'http://127.0.0.1:PORT'}]
paths:
  /items:
    get:
      operationId: listItems
      responses:
        '200':
          content:
            application/json:
              schema: {type: array, items: {type: object, properties: {id: {type: string}}}}
    post:
      operationId: createItem
      responses:
        '200':
          content:
            application/json:
              schema: {type: object, properties: {id: {type: string}}}
components:
  x-stackQL-resources:
    items:
      methods:
        list:
          operation: {$ref: '#/paths/~1items/get'}
          response: {mediaType: application/json}
        create:
          operation: {$ref: '#/paths/~1items/post'}
          response: {mediaType: application/json}
      sqlVerbs:
        select: [{$ref: '#/components/x-stackQL-resources/items/methods/list'}]
        insert: [{$ref: '#/components/x-stackQL-resources/items/methods/create'}]
""".replace("PORT", str(server.server_port)))

            service_file = doc_dir / "service.yaml"
            service = yaml.safe_load(service_file.read_text())
            if docker:
                service["servers"][0]["url"] = service["servers"][0]["url"].replace("127.0.0.1", "host.docker.internal")
            schema = {"type": "object", "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}
            response = {"200": {"content": {"application/json": {"schema": schema}}}}
            body = {"required": True, "content": {"application/json": {"schema": schema}}}
            service["paths"]["/items"]["post"]["requestBody"] = body
            service["paths"]["/items/{id}"] = {
                "get": {"operationId": "getItem", "parameters": [{"name": "id", "in": "path", "required": True, "schema": {"type": "string"}}], "responses": response},
                "patch": {"operationId": "patchItem", "parameters": [{"name": "id", "in": "path", "required": True, "schema": {"type": "string"}}], "requestBody": body, "responses": response},
                "delete": {"operationId": "deleteItem", "parameters": [{"name": "id", "in": "path", "required": True, "schema": {"type": "string"}}], "responses": {"204": {"description": "deleted"}}},
            }
            resource = service["components"]["x-stackQL-resources"]["items"]
            for name, verb in [("get", "get"), ("update", "patch"), ("delete", "delete")]:
                resource["methods"][name] = {"operation": {"$ref": "#/paths/~1items~1{id}/" + verb}, "response": {"mediaType": "application/json"}}
            resource["sqlVerbs"]["select"].append({"$ref": "#/components/x-stackQL-resources/items/methods/get"})
            resource["sqlVerbs"]["update"] = [{"$ref": "#/components/x-stackQL-resources/items/methods/update"}]
            resource["sqlVerbs"]["delete"] = [{"$ref": "#/components/x-stackQL-resources/items/methods/delete"}]
            service_file.write_text(yaml.safe_dump(service))

            def token(subject="reader", updates=None, remove=(), signer=None, kid="local"):
                claims = {"iss": "access-test", "sub": subject, "aud": "stackql", "iat": int(time.time()), "exp": int(time.time()) + 300, "groups": ["operators"]}
                claims.update(updates or {})
                for field in remove:
                    claims.pop(field, None)
                encode = lambda value: base64.urlsafe_b64encode(json.dumps(value, separators=(",", ":")).encode()).rstrip(b"=")
                message = encode({"alg": "EdDSA", "kid": kid}) + b"." + encode(claims)
                signature = (signer or key).sign(message)
                return (message + b"." + base64.urlsafe_b64encode(signature).rstrip(b"=")).decode()

            prefix = (["docker", "compose", "-p", "execrun", "run", "--rm", "--no-deps", "-T", "--user", f"{os.getuid()}:{os.getgid()}", "--workdir", str(root), "-v", str(root) + ":" + str(root), "-e", "STACKQL_CALLER_TOKEN", "--entrypoint", "stackql", "stackqlsrv"]
                      if docker else [shutil.which(executable) or str(pathlib.Path(executable).resolve())])
            base = prefix + ["--approot", str(root), "--sqlBackend", json.dumps(cfg), "--registry", json.dumps({"localDocRoot": str(root), "verifyConfig": {"nopVerify": True}}), "--preview", '{"unstable":true}', "--output", "json"]

            def run(sql, credential=None, protected=True):
                env = os.environ.copy()
                env.pop("STACKQL_CALLER_TOKEN", None)
                if credential is not None:
                    env["STACKQL_CALLER_TOKEN"] = credential
                command = base + (["--access.config", str(config_file)] if protected else []) + ["exec", sql]
                result = subprocess.run(command, capture_output=True, text=True, env=env, timeout=45)
                # Separate CLI processes initialize the shared PostgreSQL host catalog.
                # Retry only this fixture's known read when concurrent initialization deadlocks.
                for _ in range(2):
                    if sql != "SELECT id FROM stackql_unstable_accessdemo.service.items;" or "40P01" not in result.stdout + result.stderr:
                        break
                    result = subprocess.run(command, capture_output=True, text=True, env=env, timeout=45)
                output = result.stdout + result.stderr
                if credential and credential in output:
                    raise AssertionError("caller token leaked into output")
                return result.returncode, output

            def success(sql, credential, expected):
                code, output = run(sql, credential)
                assert code == 0 and expected in output, "authorized query failed: " + output

            def denied(sql, credential=None, concurrent=False):
                before = counts.copy()
                code, output = run(sql, credential)
                assert code != 0, "protected query unexpectedly succeeded: " + sql + " output: " + output
                assert counts["mutation"] == before["mutation"] and (concurrent or counts["read"] == before["read"]), "denied query reached provider: " + sql

            try:
                query = "SELECT id FROM stackql_unstable_accessdemo.service.items;"
                code, output = run(query, protected=False)
                assert code == 0 and "known-item" in output, "unprotected compatibility failed: " + output
                reader = token()
                success("SELECT whoami() AS principal, groups() AS memberships;", reader, "readers")
                success(query, reader, "known-item")
                success("SELECT 1 AS ordinary_integer;", reader, "ordinary_integer")
                invalid = [None, token(signer=ed25519.Ed25519PrivateKey.generate()), token(updates={"iss": "other"}), token(updates={"aud": "other"}), token(remove=["exp"]), token(updates={"exp": int(time.time()) - 5}), token(updates={"nbf": int(time.time()) + 600}), token(updates={"iat": int(time.time()) + 600}), token(kid="unknown"), token(subject="unknown")]
                for credential in invalid:
                    denied(query, credential)
                for sql in ["DELETE FROM stackql_unstable_accessdemo.service.items;", "INSERT INTO stackql_unstable_accessdemo.service.items (id) VALUES ('new');", "EXEC stackql_unstable_accessdemo.service.items.create;", "SELECT * FROM stackql_preview.iac.run WHERE collection = 'test' AND resources = '[]';", "SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'test';", "SELECT * FROM stackql_access_principals;", "SELECT pg_read_file('/etc/passwd');"]:
                    denied(sql, reader)
                denied(query, token(subject="denied"))
                with concurrent.futures.ThreadPoolExecutor(max_workers=2) as executor:
                    futures = [executor.submit(success, query, reader, "known-item"), executor.submit(denied, "SELECT 1", token(subject="denied"), True)]
                    for future in futures:
                        future.result()
                self._verify_https(base, config_file, root, key, token, invalid, query, db, postgres, counts)
                db.execute("DELETE FROM stackql_access_memberships WHERE principal_id = 'reader'")
                if not postgres:
                    db.commit()
                denied(query, reader)
                db.execute("UPDATE stackql_access_principals SET attributes = 'null' WHERE id = 'operator'")
                if not postgres:
                    db.commit()
                denied(query, token(subject="operator"))
                assert counts["mutation"] == 0, "reader performed provider mutation"
                db.execute("UPDATE stackql_access_principals SET attributes = '{}' WHERE id = 'operator'")
                if not postgres:
                    db.commit()
                operator = token(subject="operator")
                def converge(label, keys):
                    resources = [{"key": name, "provider": "accessdemo", "address": "service.items", "identity": "id", "addressed_by": "id", "desired": {"name": {"from": "input", "attribute": "label"}}} for name in keys]
                    spec = json.dumps(resources).replace("'", "''")
                    inputs = json.dumps({"label": label})
                    sql = "SELECT * FROM stackql_preview.iac.run WHERE collection = 'acceptance' AND state = 'audit' AND run_id = '" + label + "' AND resources = '" + spec + "' AND inputs = '" + inputs + "';"
                    success(sql, operator, "applied")
                    assert all(item.get("name") == label for item in managed.values()), "IaC reported applied without converging desired values"
                converge("first", ["one", "two"])
                assert len(managed) == 2, "IaC create did not retain both resources"
                db.execute("INSERT INTO stackql_access_principals VALUES ('operator2','{}')")
                db.execute("INSERT INTO stackql_access_identities VALUES ('access-test','operator2','operator2')")
                db.execute("INSERT INTO stackql_access_memberships VALUES ('operator2','operators')")
                if not postgres:
                    db.commit()
                denied("SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'acceptance' AND state = 'audit' AND keys = '[\"one\"]';", token(subject="operator2"))
                denied("SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'acceptance' AND state = '../../escape';", operator)
                converge("second", ["one", "two"])
                converge("third", ["one", "two"])
                denied("SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'acceptance' AND state = 'audit' AND keys = '[]';", operator)
                success("SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'acceptance' AND state = 'audit' AND run_id = 'prune' AND keys = '[\"two\"]';", operator, "destroyed")
                assert len(managed) == 1, "partial delete changed the retained resource set"
                converge("fourth", ["one"])
                success("SELECT * FROM stackql_preview.iac.destroy WHERE collection = 'acceptance' AND state = 'audit' AND run_id = 'teardown';", operator, "destroyed")
                assert not managed, "IaC teardown left managed resources"
                spec = json.dumps([{"key": "reject", "provider": "accessdemo", "address": "service.items", "identity": "id", "addressed_by": "id", "desired": {"name": "upstream-reject"}}])
                sql = "SELECT * FROM stackql_preview.iac.run WHERE collection = 'rejection' AND run_id = 'reject' AND resources = '" + spec + "';"
                code, output = run(sql, protected=False)
                assert code != 0 and "failed" in output.lower(), "unsuccessful IaC returned success: " + output
                assert not managed, "failed create unexpectedly retained resources"

            finally:
                server.shutdown()
                server.server_close()
                if directory_schema:
                    db.execute(psycopg.sql.SQL("DROP SCHEMA {} CASCADE").format(psycopg.sql.Identifier(directory_schema)))
                db.close()

    def _verify_https(self, base, config_file, root, key, token, invalid, query, db, postgres, counts):
        subject = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, "localhost")])
        now = datetime.datetime.now(datetime.timezone.utc)
        certificate = (x509.CertificateBuilder().subject_name(subject).issuer_name(subject)
                       .public_key(key.public_key()).serial_number(x509.random_serial_number())
                       .not_valid_before(now - datetime.timedelta(minutes=1))
                       .not_valid_after(now + datetime.timedelta(hours=1))
                       .add_extension(x509.SubjectAlternativeName([x509.IPAddress(ipaddress.ip_address("127.0.0.1"))]), False)
                       .sign(key, None))
        cert_path, key_path = root / "server.crt", root / "server.key"
        cert_path.write_bytes(certificate.public_bytes(serialization.Encoding.PEM))
        key_path.write_bytes(key.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
        key_path.chmod(0o600)
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        config = {"server": {"transport": "http", "address": "0.0.0.0:" + str(port),
                             "protocol_version": "2026-07-28", "mode": "full_access",
                             "tls_cert_file": str(cert_path), "tls_key_file": str(key_path)},
                  "enabled_tools": ["run_select_query"]}
        env = os.environ.copy()
        # Deliberately set a privileged process token: remote clients must never inherit it.
        env["STACKQL_CALLER_TOKEN"] = token(subject="operator")
        container = "stackql-access-" + uuid.uuid4().hex
        server_base = list(base)
        docker = server_base[0] == "docker"
        if docker:
            index = server_base.index("stackqlsrv")
            server_base[index:index] = ["--publish", "127.0.0.1:" + str(port) + ":" + str(port), "--name", container]
        command = server_base + ["--access.config", str(config_file), "--mcp.config", json.dumps(config), "mcp"]
        tls = ssl.create_default_context(cafile=str(cert_path))
        with (root / "mcp.log").open("w+") as log:
            process = subprocess.Popen(command, stdout=log, stderr=log, env=env)
            try:
                def call(credential, sql=query):
                    headers = {"Content-Type": "application/json", "Accept": "application/json, text/event-stream", "MCP-Protocol-Version": "2026-07-28", "Mcp-Method": "tools/call", "Mcp-Name": "run_select_query"}
                    if credential:
                        headers["Authorization"] = "Bearer " + credential
                    payload = {"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"_meta": {"io.modelcontextprotocol/protocolVersion": "2026-07-28", "io.modelcontextprotocol/clientCapabilities": {}, "io.modelcontextprotocol/clientInfo": {"name": "access-test", "version": "1"}}, "name": "run_select_query", "arguments": {"sql": sql}}}
                    request = urllib.request.Request("https://127.0.0.1:" + str(port), data=json.dumps(payload).encode(), headers=headers)
                    try:
                        with urllib.request.urlopen(request, context=tls, timeout=20) as response:
                            return response.status, response.read().decode()
                    except urllib.error.HTTPError as error:
                        return error.code, error.read().decode()
                for _ in range(100):
                    if process.poll() is not None:
                        log.seek(0)
                        raise AssertionError("protected HTTPS server failed to start: " + log.read())
                    try:
                        status, _ = call(None)
                        assert status == 401, "missing HTTP token accepted"
                        break
                    except (OSError, urllib.error.URLError):
                        time.sleep(0.1)
                else:
                    raise AssertionError("protected HTTPS server did not start")
                reader = token()
                status, output = call(reader)
                assert status == 200 and "known-item" in output, "authorized HTTPS read failed: " + output
                for credential in invalid:
                    before = counts.copy()
                    status, _ = call(credential)
                    assert status == 401 and counts == before, "invalid HTTP credential reached execution"
                def read():
                    status, output = call(reader)
                    assert status == 200 and "known-item" in output, "concurrent HTTPS reader failed: " + output
                def deny():
                    status, output = call(token(subject="denied"))
                    assert "known-item" not in output and (status != 200 or '"isError":true' in output.replace(" ", "")), "concurrent caller inherited access"
                with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
                    for future in [pool.submit(read), pool.submit(deny)]:
                        future.result()
                db.execute("DELETE FROM stackql_access_memberships WHERE principal_id = 'reader'")
                if not postgres:
                    db.commit()
                status, output = call(reader)
                assert "known-item" not in output and (status != 200 or '"isError":true' in output.replace(" ", "")), "HTTP membership revocation ignored: " + output
                db.execute("INSERT INTO stackql_access_memberships VALUES ('reader','readers')")
                if not postgres:
                    db.commit()
            finally:
                process.terminate()
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=10)
                if docker:
                    subprocess.run(["docker", "rm", "-f", container], capture_output=True, timeout=30)
                log.seek(0)
                text = log.read()
                assert env["STACKQL_CALLER_TOKEN"] not in text, "server token leaked into log"
