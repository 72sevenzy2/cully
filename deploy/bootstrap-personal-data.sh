#!/bin/sh
set -eu
test "$(id -u)" = 0 || { echo 'Run bootstrap as root.' >&2; exit 1; }
test "$#" = 1 || { echo 'Usage: bootstrap-personal-data.sh DEPLOY_PUBLIC_KEY_FILE' >&2; exit 1; }
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
docker network inspect workspace_workspace >/dev/null
docker compose version >/dev/null
install -d -m 700 /opt/cully-personal-data
test -f /opt/cully-personal-data/.env || { echo 'Configure /opt/cully-personal-data/.env before bootstrap.' >&2; exit 1; }
chmod 600 /opt/cully-personal-data/.env
install -m 644 "$script_dir/compose.personal-data.yaml" /opt/cully-personal-data/compose.yaml
install -m 644 "$script_dir/self-hosted/mem0-init-db.sql" /opt/cully-personal-data/mem0-init-db.sql
install -d -m 755 /usr/local/libexec
install -m 755 "$script_dir/personal_data_controller.py" /usr/local/libexec/cully-personal-data-deploy
install -d -m 700 /root/.ssh
touch /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys
python3 - "$1" <<'PY'
from pathlib import Path
import os
import re
import sys
import tempfile
key = Path(sys.argv[1]).read_text().strip()
if not re.fullmatch(r'ssh-ed25519 [A-Za-z0-9+/=]+(?: [^\r\n]*)?', key):
    raise SystemExit('Expected one Ed25519 public key')
entry = 'restrict,command="/usr/local/libexec/cully-personal-data-deploy" ' + key
path = Path('/root/.ssh/authorized_keys')
lines = path.read_text().splitlines()
# The one-time password bootstrap adds this public key without restrictions.
# Replace that entry after setup so the stored CI key can only deploy Cully.
new_lines = [line for line in lines if line != key and line != entry]
new_lines.append(entry)
with tempfile.NamedTemporaryFile(mode='w', dir=path.parent, delete=False) as f:
    f.write('\n'.join(new_lines) + '\n')
    temporary = Path(f.name)
os.chmod(temporary, 0o600)
os.replace(temporary, path)
PY
echo 'Restricted personal data deployment controller installed.'
