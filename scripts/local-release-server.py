#!/usr/bin/env python3
"""
local-release-server.py
=======================
A local HTTP server that mimics the HashiCorp releases.hashicorp.com structure
for Packer binaries. It reads binaries from the bin/ directory, zips them
on demand, computes SHA256 checksums, and serves them so the hcp-sbom
provisioner can download them without hitting the real release server.

URL layout served:
  GET /packer/<version>/packer_<version>_<os>_<arch>.zip   -> zip of the binary
  GET /packer/<version>/packer_<version>_SHA256SUMS        -> checksum file

Binary resolution order (bin/ directory, relative to repo root):
  1. bin/packer-<os>-<arch>        (e.g. bin/packer-linux-amd64)
  2. bin/packer-<os>-<arch>.exe    (Windows)
  3. bin/packer                    (fallback — whatever single binary is present)

Usage:
  python3 scripts/local-release-server.py [--port 3231] [--bin-dir bin/]
"""

import argparse
import hashlib
import http.server
import io
import os
import re
import sys
import zipfile
from pathlib import Path
from typing import Optional

REPO_ROOT = Path(__file__).resolve().parent.parent


def find_binary(bin_dir, goos, goarch):
    # type: (Path, str, str) -> Optional[Path]
    """Return the path to the best-matching packer binary for the given OS/arch.

    pkg/ layout: pkg/<goos>_<goarch>/packer  (or packer.exe on Windows)
    """
    subdir = bin_dir / "{}_{}" .format(goos, goarch)
    candidates = [
        subdir / "packer.exe",
        subdir / "packer",
    ]
    for path in candidates:
        if path.is_file():
            return path
    return None


def make_zip(binary_path, goos):
    # type: (Path, str) -> bytes
    """Zip the binary and return the raw zip bytes."""
    binary_name = "packer.exe" if goos == "windows" else "packer"
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, mode="w", compression=zipfile.ZIP_DEFLATED) as zf:
        zf.write(str(binary_path), arcname=binary_name)
    return buf.getvalue()


def sha256_bytes(data):
    # type: (bytes) -> str
    return hashlib.sha256(data).hexdigest()


class ReleaseHandler(http.server.BaseHTTPRequestHandler):
    bin_dir = REPO_ROOT / "pkg"

    ZIP_RE = re.compile(r"^/packer/([^/]+)/packer_([^/]+)_([^/]+)_([^/]+)\.zip$")
    SUMS_RE = re.compile(r"^/packer/([^/]+)/packer_([^/]+)_SHA256SUMS$")

    def log_message(self, fmt, *args):
        print("[server] {} - {}".format(self.address_string(), fmt % args), file=sys.stderr)

    def send_bytes(self, data, content_type="application/octet-stream"):
        self.send_response(200)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        path = self.path.split("?")[0]

        m = self.ZIP_RE.match(path)
        if m:
            version, _, goos, goarch = m.group(1), m.group(2), m.group(3), m.group(4)
            binary = find_binary(self.bin_dir, goos, goarch)
            if binary is None:
                self._not_found("no binary found for {}/{} in {}".format(goos, goarch, self.bin_dir))
                return
            zip_data = make_zip(binary, goos)
            print("[server] serving zip for {}/{} v{} from {} ({} KB)".format(
                goos, goarch, version, binary, len(zip_data) // 1024), file=sys.stderr)
            self.send_bytes(zip_data, "application/zip")
            return

        m = self.SUMS_RE.match(path)
        if m:
            version = m.group(1)
            lines = []
            # pkg/<goos>_<goarch>/packer[.exe]
            for subdir in sorted(self.bin_dir.iterdir()):
                if not subdir.is_dir():
                    continue
                parts = subdir.name.split("_", 1)
                if len(parts) != 2:
                    continue
                goos, goarch = parts
                binary = find_binary(self.bin_dir, goos, goarch)
                if binary is None:
                    continue
                zip_data = make_zip(binary, goos)
                chk = sha256_bytes(zip_data)
                fname = "packer_{}_{}_{}.zip".format(version, goos, goarch)
                lines.append("{}  {}".format(chk, fname))
            body = "\n".join(lines) + "\n"
            self.send_bytes(body.encode(), "text/plain")
            return

        self._not_found("unrecognised path: {}".format(path))

    def _not_found(self, reason):
        print("[server] 404 {}".format(reason), file=sys.stderr)
        self.send_response(404)
        self.end_headers()
        self.wfile.write("404 not found: {}\n".format(reason).encode())


def main():
    parser = argparse.ArgumentParser(description="Local Packer release server")
    parser.add_argument("--port", type=int, default=3231)
    parser.add_argument("--bin-dir", default=str(REPO_ROOT / "pkg"))
    args = parser.parse_args()

    ReleaseHandler.bin_dir = Path(args.bin_dir).resolve()
    if not ReleaseHandler.bin_dir.is_dir():
        print("error: bin-dir {} does not exist".format(ReleaseHandler.bin_dir), file=sys.stderr)
        sys.exit(1)

    server = http.server.HTTPServer(("127.0.0.1", args.port), ReleaseHandler)
    print("[server] listening on http://127.0.0.1:{}".format(args.port), file=sys.stderr)
    print("[server] serving binaries from {}".format(ReleaseHandler.bin_dir), file=sys.stderr)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n[server] stopped", file=sys.stderr)


if __name__ == "__main__":
    main()
