#!/usr/bin/env python3
"""Restricted SSH controller: deploy only the Cully website Compose project."""
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

ROOT = Path('/opt/cully-web')
REVISION = re.compile(r'[0-9a-f]{40}')


def validate_request(request):
    if not isinstance(request, dict) or request.get('action') not in ('deploy', 'rollback'):
        raise ValueError('unsupported deployment action')
    if not isinstance(request.get('revision'), str) or not REVISION.fullmatch(request['revision']):
        raise ValueError('invalid deployment revision')
    if request['action'] == 'deploy':
        token = request.get('token')
        user = request.get('registry_user')
        if not isinstance(token, str) or not 1 <= len(token) <= 8192 or '\n' in token:
            raise ValueError('missing registry credential')
        if not isinstance(user, str) or not re.fullmatch(r'[A-Za-z0-9_\[\]-]{1,100}', user):
            raise ValueError('invalid registry user')


def write_state(root, state):
    with tempfile.NamedTemporaryFile(mode='w', dir=root, delete=False) as f:
        json.dump(state, f)
        temporary = Path(f.name)
    os.replace(temporary, root / 'state.json')


def compose(root, revision, runner, docker_config=None):
    env = dict(os.environ, CULLY_WEB_REVISION=revision)
    if docker_config is not None:
        env['DOCKER_CONFIG'] = docker_config
    runner(['docker', 'compose', '--project-name', 'cully-web', '--file',
            str(root / 'compose.yaml'), 'up', '-d', '--wait', '--wait-timeout', '120'],
           env=env, check=True)


def execute(request, root=ROOT, runner=subprocess.run):
    validate_request(request)
    state_file = root / 'state.json'
    state = json.loads(state_file.read_text()) if state_file.exists() else {}
    for revision in state.values():
        if revision is not None and (not isinstance(revision, str) or not REVISION.fullmatch(revision)):
            raise ValueError('invalid saved deployment state')
    current = state.get('current')
    if request['action'] == 'rollback':
        if current != request['revision']:
            raise ValueError('refusing to roll back a different deployment')
        previous = state.get('previous')
        if not previous:
            raise ValueError('no previous website release available')
        compose(root, previous, runner)
        write_state(root, {'current': previous, 'previous': current})
        return previous

    revision = request['revision']
    with tempfile.TemporaryDirectory(prefix='cully-registry-') as directory:
        # The job-scoped registry token is supplied via stdin and removed with
        # this temporary Docker config. It never appears in command arguments.
        runner(['docker', '--config', directory, 'login', 'ghcr.io', '--username',
                request['registry_user'], '--password-stdin'],
               input=request['token'], text=True, check=True, capture_output=True)
        image = 'ghcr.io/mcp-runtime/cully-web:' + revision
        runner(['docker', '--config', directory, 'pull', image], check=True)
        try:
            compose(root, revision, runner, directory)
        except subprocess.CalledProcessError:
            if current:
                compose(root, current, runner)
            raise
    previous = state.get('previous') if revision == current else current
    write_state(root, {'current': revision, 'previous': previous})
    return revision


def main():
    if os.environ.get('SSH_ORIGINAL_COMMAND') != 'deploy':
        raise ValueError('only the deploy SSH command is allowed')
    raw = sys.stdin.buffer.read(16385)
    if len(raw) > 16384:
        raise ValueError('deployment request too large')
    revision = execute(json.loads(raw))
    print('Website revision active:', revision)


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        # Never print the input request, token, or subprocess captured output.
        print('Website deployment failed:', type(error).__name__, file=sys.stderr)
        sys.exit(1)
