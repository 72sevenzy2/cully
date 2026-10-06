#!/bin/sh
set -eu
test "$(id -u)" = 0 || { echo 'Run bootstrap as root.' >&2; exit 1; }
test "$#" = 1 || { echo 'Usage: bootstrap-website.sh DEPLOY_PUBLIC_KEY_FILE' >&2; exit 1; }
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
docker network inspect workspace_workspace >/dev/null
docker compose version >/dev/null
install -d -m 755 /opt/cully-web /usr/local/libexec
install -m 644 "$script_dir/compose.website.yaml" /opt/cully-web/compose.yaml
install -m 755 "$script_dir/website_controller.py" /usr/local/libexec/cully-web-deploy
install -d -m 700 /root/.ssh
touch /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys
python3 - "$1" <<'PY'
from pathlib import Path
import re
import sys
key = Path(sys.argv[1]).read_text().strip()
if not re.fullmatch(r'ssh-ed25519 [A-Za-z0-9+/=]+(?: [^\r\n]*)?', key):
    raise SystemExit('Expected one Ed25519 public key')
entry = 'restrict,command="/usr/local/libexec/cully-web-deploy" ' + key
path = Path('/root/.ssh/authorized_keys')
text = path.read_text()
if entry not in text.splitlines():
    with path.open('a') as f:
        f.write(('' if not text or text.endswith('\n') else '\n') + entry + '\n')
PY
echo 'Restricted website deploy key and Docker controller installed.'
