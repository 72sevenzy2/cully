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
                        {'action': 'deploy', 'revision': A},
                        {'action': 'deploy-docs', 'revision': '../../root'},
                        {'action': 'deploy-website', 'revision': A, 'registry_user': 'a;id', 'token': 'test'}):
            with self.assertRaises(ValueError):
                validate_request(request)

    def test_healthy_deploy_and_rollback(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = []
            def runner(args, **kwargs):
                calls.append((args, kwargs))
            request = dict(action='deploy-website', revision=A, token='test-token', registry_user='tester')
            execute(request, root, runner)
            request['revision'] = B
            execute(request, root, runner)
            self.assertEqual(json.loads((root / 'state.json').read_text()), {'current': B, 'previous': A})
            self.assertNotIn('test-token', str([args for args, kwargs in calls]))
            execute({'action': 'rollback-website', 'revision': B}, root, runner)
            self.assertEqual(json.loads((root / 'state.json').read_text())['current'], A)
            with self.assertRaises(ValueError):
                execute({'action': 'rollback-website', 'revision': B}, root, runner)

    def test_docs_deploy_and_rollback_do_not_touch_website(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = []
            def runner(args, **kwargs):
                calls.append((args, kwargs))
            execute(dict(action='deploy-website', revision=A, token='test', registry_user='tester'), root, runner)
            calls.clear()
            execute(dict(action='deploy-docs', revision=B, token='test', registry_user='tester'), root, runner)
            execute({'action': 'rollback-docs', 'revision': B}, root, runner)
            self.assertIsNone(json.loads((root / 'docs-state.json').read_text())['current'])
            execute(dict(action='deploy-docs', revision=A, token='test', registry_user='tester'), root, runner)
            execute(dict(action='deploy-docs', revision=B, token='test', registry_user='tester'), root, runner)
            execute({'action': 'rollback-docs', 'revision': B}, root, runner)
            self.assertEqual(json.loads((root / 'state.json').read_text())['current'], A)
            self.assertEqual(json.loads((root / 'docs-state.json').read_text())['current'], A)
            self.assertTrue(all('cully-web' not in ' '.join(args) for args, _ in calls))
            self.assertTrue(any('compose.docs.yaml' in ' '.join(args) for args, _ in calls))

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
                execute(dict(action='deploy-website', revision=B, token='test', registry_user='tester'), root, runner)
            self.assertEqual(activated, [B, A])
            self.assertEqual(json.loads((root / 'state.json').read_text())['current'], A)


if __name__ == '__main__':
    unittest.main()
