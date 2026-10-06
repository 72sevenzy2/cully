import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from website_controller import execute, validate_request

A = 'a' * 40
B = 'b' * 40


class DeploymentTests(unittest.TestCase):
    def test_rejects_untrusted_requests(self):
        for request in ({}, {'action': 'shell', 'revision': A},
                        {'action': 'deploy', 'revision': '../../root'},
                        {'action': 'deploy', 'revision': A, 'registry_user': 'a;id', 'token': 'test'}):
            with self.assertRaises(ValueError):
                validate_request(request)

    def test_healthy_deploy_and_rollback(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = []
            def runner(args, **kwargs):
                calls.append((args, kwargs))
            request = dict(action='deploy', revision=A, token='test-token', registry_user='tester')
            execute(request, root, runner)
            request['revision'] = B
            execute(request, root, runner)
            self.assertEqual(json.loads((root / 'state.json').read_text()), {'current': B, 'previous': A})
            self.assertNotIn('test-token', str([args for args, kwargs in calls]))
            execute({'action': 'rollback', 'revision': B}, root, runner)
            self.assertEqual(json.loads((root / 'state.json').read_text())['current'], A)
            with self.assertRaises(ValueError):
                execute({'action': 'rollback', 'revision': B}, root, runner)

    def test_unhealthy_deploy_restores_previous(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'state.json').write_text(json.dumps({'current': A, 'previous': None}))
            activated = []
            def runner(args, **kwargs):
                if 'compose' in args:
                    revision = kwargs['env']['CULLY_WEB_REVISION']
                    activated.append(revision)
                    if revision == B:
                        raise subprocess.CalledProcessError(1, args)
            with self.assertRaises(subprocess.CalledProcessError):
                execute(dict(action='deploy', revision=B, token='test', registry_user='tester'), root, runner)
            self.assertEqual(activated, [B, A])
            self.assertEqual(json.loads((root / 'state.json').read_text())['current'], A)


if __name__ == '__main__':
    unittest.main()
