"""Exercise the installer without network access or real agent configuration."""
import hashlib
import io
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest


INSTALLER = Path(__file__).resolve().parents[1] / "install.sh"


class InstallerTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.cli = self.root / "cully"
        self.cli.write_text('''#!/bin/sh
case "$1" in
  version) echo v-test ;;
  help) echo 'cully agent setup' ;;
  agent) printf '%s\\n' "$*" > "$TEST_ROOT/setup" ;;
esac
''')
        self.cli.chmod(0o755)
        with tarfile.open(self.root / "archive.tar.gz", "w:gz") as archive:
            content = self.cli.read_bytes()
            entry = tarfile.TarInfo("cully")
            entry.size = len(content)
            entry.mode = 0o755
            archive.addfile(entry, io.BytesIO(content))
        digest = hashlib.sha256((self.root / "archive.tar.gz").read_bytes()).hexdigest()
        (self.root / "checksums.txt").write_text(f"{digest}  cully_linux_amd64.tar.gz\n")
        self.stub("uname", 'case "$1" in -s) echo Linux ;; -m) echo x86_64 ;; esac')
        self.stub("curl", '''
printf '%s\\n' "$*" >> "$TEST_ROOT/curl-calls"
[ "${DOWNLOAD_FAIL:-0}" = 0 ] || exit 28
source_file=archive.tar.gz
while [ "$#" -gt 0 ]; do
  case "$1" in
    */checksums.txt) source_file=checksums.txt ;;
    -o) shift; destination="$1" ;;
  esac
  shift
done
cp "$TEST_ROOT/$source_file" "$destination"
''')
        self.stub("go", '''
printf '%s\\n' "$*" > "$TEST_ROOT/go-call"
cp "$TEST_ROOT/cully" "$GOBIN/cully"
''')
        self.env = dict(os.environ, TEST_ROOT=str(self.root),
                        CLAUDE_CONFIG_DIR=str(self.root / "claude"),
                        PATH=str(self.bin) + os.pathsep + os.environ["PATH"])

    def stub(self, name, body):
        path = self.bin / name
        path.write_text("#!/bin/sh\nset -eu\n" + body + "\n")
        path.chmod(0o755)

    def run_installer(self, *args):
        return subprocess.run(["sh", str(INSTALLER), *args], env=self.env,
                              capture_output=True, text=True, timeout=10)

    def test_binary_install_forwards_agent_and_mcp_options(self):
        result = self.run_installer("--agent", "codex", "--mcp-url",
                                    "https://example.com/mcp", "--oauth")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.root / "setup").read_text().strip(),
                         "agent setup codex --mcp-url https://example.com/mcp --oauth")
        self.assertFalse((self.root / "go-call").exists())
        calls = (self.root / "curl-calls").read_text()
        self.assertIn("--progress-bar --connect-timeout 10 --max-time 120", calls)
        self.assertIn("--speed-limit 1024 --speed-time 30", calls)
        self.assertIn("--max-time 30", calls)

    def test_failed_download_does_not_compile_or_install(self):
        self.env["DOWNLOAD_FAIL"] = "1"
        result = self.run_installer()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("--from-source", result.stderr)
        self.assertFalse((self.root / "go-call").exists())
        self.assertFalse((self.root / "claude" / "bin" / "cully").exists())

    def test_explicit_source_install_uses_selected_ref(self):
        for version, ref in [("latest", "main"), ("v0.3.0", "v0.3.0")]:
            with self.subTest(version=version):
                self.env["CULLY_VERSION"] = version
                result = self.run_installer("--from-source", "--agent", "cursor")
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertIn("several minutes", result.stdout)
                self.assertEqual((self.root / "go-call").read_text().strip(),
                                 f"install -v github.com/mcp-runtime/cully/cmd/cully@{ref}")
                self.assertFalse((self.root / "curl-calls").exists())

    def test_checksum_mismatch_aborts_install(self):
        (self.root / "checksums.txt").write_text("bad  cully_linux_amd64.tar.gz\n")
        result = self.run_installer()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("checksum mismatch", result.stderr)
        self.assertFalse((self.root / "claude" / "bin" / "cully").exists())


if __name__ == "__main__":
    unittest.main()
