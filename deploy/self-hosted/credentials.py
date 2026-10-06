#!/usr/bin/env python3
"""Manage private credentials for the Docker self-hosted stack."""

import json
import os
import secrets
import shlex
import sys
import tempfile
from pathlib import Path
from urllib.parse import quote, unquote, urlsplit


FIELDS = {
    "CULLY_DB_PASSWORD": "db_password",
    "CULLY_DATA_API_TOKEN": "data_api_token",
    "CULLY_MCP_OWNER": "mcp_owner",
    "CULLY_MEM0_DB_PASSWORD": "mem0_db_password",
    "CULLY_MEM0_API_KEY": "mem0_api_key",
    "CULLY_MEM0_JWT_SECRET": "mem0_jwt_secret",
}


def config_path():
    return Path(os.environ.get("CULLY_CONFIG_PATH", Path.home() / ".cully" / "config.json"))


def read_env(path):
    values = {}
    if not path.exists():
        return values
    for line in path.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip().strip('"').strip("'")
    return values


def existing_secret(value):
    return bool(value) and not value.startswith("replace-with-")


def load(path):
    if path.is_symlink():
        raise ValueError(f"refusing symlinked config: {path}")
    if not path.exists():
        return {}
    if not path.is_file():
        raise ValueError(f"config is not a regular file: {path}")
    document = json.loads(path.read_text())
    if not isinstance(document, dict):
        raise ValueError("config.json must contain a JSON object")
    return document


def private_directory(path):
    path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    if path.parent.is_symlink():
        raise ValueError(f"refusing symlinked config directory: {path.parent}")
    if path.parent.name == ".cully":
        path.parent.chmod(0o700)
    elif path.parent.stat().st_mode & 0o077:
        raise ValueError(f"config directory must be private: {path.parent}")


def save(path, document):
    private_directory(path)
    fd, temporary = tempfile.mkstemp(prefix=".config-", dir=path.parent)
    try:
        with os.fdopen(fd, "w") as output:
            os.fchmod(output.fileno(), 0o600)
            json.dump(document, output, indent=2)
            output.write("\n")
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def ensure(path, env_path):
    document = load(path)
    private_directory(path)
    current = document.get("self_hosted", {})
    if not isinstance(current, dict):
        raise ValueError("self_hosted must be a JSON object")
    current = current.copy()
    env = read_env(env_path)
    current.setdefault("db_name", env.get("CULLY_DB_NAME") or "cully")
    current.setdefault("db_user", env.get("CULLY_DB_USER") or "cully")
    for key, field in FIELDS.items():
        if current.get(field):
            continue
        value = env.get(key, "")
        if field == "db_password" and not existing_secret(value):
            url = env.get("CULLY_DATABASE_URL", "")
            if existing_secret(url):
                value = unquote(urlsplit(url).password or "")
        if not existing_secret(value):
            value = ("local-" if field == "mcp_owner" else "") + secrets.token_urlsafe(36)
        current[field] = value
    required_fields = ["db_name", "db_user", *FIELDS.values()]
    if not all(isinstance(current.get(field), str) and current[field] for field in required_fields):
        raise ValueError("self_hosted config contains an empty or non-string value")
    if document.get("self_hosted") != current:
        document["self_hosted"] = current
        save(path, document)
    else:
        path.chmod(0o600)
    return current


def exports(current):
    values = {key: current[field] for key, field in FIELDS.items()}
    values["CULLY_DB_NAME"] = current["db_name"]
    values["CULLY_DB_USER"] = current["db_user"]
    values["CULLY_DATABASE_URL"] = (
        f"postgres://{quote(current['db_user'], safe='')}:"
        f"{quote(current['db_password'], safe='')}@db:5432/"
        f"{quote(current['db_name'], safe='')}"
    )
    for key, value in values.items():
        print(f"export {key}={shlex.quote(value)}")


def validate_oauth(mode, endpoint):
    values = read_env(Path(__file__).with_name(".env"))
    values.update({key: value for key, value in os.environ.items() if key.startswith("CULLY_AUTH_") or key in {"CULLY_MCP_HOST", "CULLY_AUTH_HOST", "KEYCLOAK_CLIENT_SECRET"}})
    required = ["CULLY_AUTH_ISSUER", "CULLY_AUTH_RESOURCE", "CULLY_JWKS_URL", "CULLY_MCP_HOST"]
    if mode == "mcp-auth":
        required += ["CULLY_AUTH_HOST", "KEYCLOAK_CLIENT_SECRET"]
    missing = [key for key in required if not values.get(key) or "example.com" in values[key] or values[key].startswith("replace-with-")]
    if missing:
        raise ValueError("set real OAuth values in .env: " + ", ".join(missing))
    for key in ("CULLY_AUTH_ISSUER", "CULLY_AUTH_RESOURCE", "CULLY_JWKS_URL"):
        if not values[key].startswith("https://"):
            raise ValueError(f"{key} must use HTTPS")
    if endpoint and endpoint != values["CULLY_AUTH_RESOURCE"]:
        raise ValueError("--mcp-url must equal CULLY_AUTH_RESOURCE")


def main():
    if len(sys.argv) < 2 or sys.argv[1] not in {"ensure", "export", "validate-oauth"}:
        raise ValueError("usage: credentials.py ensure|export|validate-oauth MODE [URL]")
    if sys.argv[1] == "validate-oauth":
        if len(sys.argv) not in {3, 4} or sys.argv[2] not in {"existing", "mcp-auth"}:
            raise ValueError("usage: credentials.py validate-oauth existing|mcp-auth [URL]")
        validate_oauth(sys.argv[2], sys.argv[3] if len(sys.argv) == 4 else "")
        return
    if len(sys.argv) != 2:
        raise ValueError("usage: credentials.py ensure|export")
    path = config_path()
    if sys.argv[1] == "ensure":
        ensure(path, Path(__file__).with_name(".env"))
        print(f"Self-hosted credentials ready in {path}")
    else:
        document = load(path)
        current = document.get("self_hosted", {})
        if not isinstance(current, dict) or any(not current.get(field) for field in FIELDS.values()):
            raise ValueError("run ./setup.sh to create the self-hosted credentials")
        exports(current)


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, KeyError, json.JSONDecodeError) as error:
        print(f"credentials: {error}", file=sys.stderr)
        sys.exit(1)
