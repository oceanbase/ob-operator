#!/usr/bin/env python3
"""Internal test artifact server. Serves only explicitly allowlisted files."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
from pathlib import Path
import shutil


def handler(routes):
    class Handler(BaseHTTPRequestHandler):
        def do_HEAD(self):
            self.serve(False)

        def do_GET(self):
            self.serve(True)

        def serve(self, body):
            if self.path == "/healthz":
                payload = b'{"status":"ready","scope":"internal-test-artifacts"}\n'
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(payload)))
                self.end_headers()
                if body:
                    self.wfile.write(payload)
                return
            # Exact matches only: no directory browsing, URL query, traversal,
            # arbitrary local path, credentials, upload or Kubernetes endpoint.
            source = routes.get(self.path)
            if source is None:
                self.send_error(404)
                return
            try:
                stream = Path(source).open("rb")
            except OSError:
                self.send_error(503, "Artifact unavailable")
                return
            with stream:
                self.send_response(200)
                self.send_header("Content-Type", "application/octet-stream")
                self.send_header("Content-Length", str(Path(source).stat().st_size))
                self.send_header("Cache-Control", "public, max-age=31536000, immutable")
                self.end_headers()
                if body:
                    try:
                        shutil.copyfileobj(stream, self.wfile, 1024 * 1024)
                    except (BrokenPipeError, ConnectionResetError):
                        pass
    return Handler


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", type=Path, required=True)
    parser.add_argument("--bind", required=True)
    parser.add_argument("--port", type=int, default=18088)
    args = parser.parse_args()
    routes = json.loads(args.config.read_text())
    if not routes or not all(key.startswith("/") and Path(value).is_file() for key, value in routes.items()):
        raise SystemExit("Every route must map to an existing explicitly selected artifact")
    ThreadingHTTPServer((args.bind, args.port), handler(routes)).serve_forever()
