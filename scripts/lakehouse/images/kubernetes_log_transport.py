"""Bounded Kubernetes log transport for the pinned obdiag 5.0.1 package.

Only diagnostic temporary archives are written remotely. Observer images,
executables and configuration are never changed. Python >=3.6 is required in
the target container; tar/gzip are optional optimizations for archive creation.
"""
import base64
import hashlib
import json
import os
import shlex
import tempfile
import time

MAX_ARCHIVE_BYTES = 2 * 1024 * 1024 * 1024
PROBE_CODE = '''import base64,gzip,hashlib,io,json,os,stat,sys,tarfile
assert sys.version_info >= (3,6)
payload=b"lakehouse-diagnostic-probe"
buffer=io.BytesIO()
with tarfile.open(fileobj=buffer,mode="w:gz") as archive:
    member=tarfile.TarInfo("probe")
    member.size=len(payload)
    archive.addfile(member,io.BytesIO(payload))
raw=buffer.getvalue()
assert gzip.decompress(raw)
with tarfile.open(fileobj=io.BytesIO(base64.b64decode(base64.b64encode(raw))),mode="r:gz") as archive:
    actual=archive.extractfile("probe").read()
assert hashlib.sha256(actual).digest()==hashlib.sha256(payload).digest()
'''

PACK_CODE = '''import os,stat,sys,tarfile,tempfile
source,destination=sys.argv[1:]
source=os.path.abspath(source)
if os.path.islink(source) or not os.path.isdir(source) or destination!=source+".tar.gz":
    raise ValueError("invalid log archive path")
entries=[]
size=0
for parent,dirs,files in os.walk(source,followlinks=False):
    for name in dirs+files:
        path=os.path.join(parent,name)
        info=os.lstat(path)
        if not (stat.S_ISDIR(info.st_mode) or stat.S_ISREG(info.st_mode)):
            raise ValueError("non-regular log archive member")
        if stat.S_ISREG(info.st_mode):
            size+=info.st_size
            if size>2147483648:
                raise ValueError("log archive exceeds byte limit")
        entries.append(path)
fd,temporary=tempfile.mkstemp(prefix=".obdiag-pack-",dir=os.path.dirname(source))
os.close(fd)
try:
    with tarfile.open(temporary,"w:gz") as archive:
        for path in sorted(entries):
            archive.add(path,arcname=os.path.relpath(path,os.path.dirname(source)),recursive=False)
    os.replace(temporary,destination)
finally:
    if os.path.exists(temporary): os.unlink(temporary)
'''

DOWNLOAD_CODE = '''import base64,hashlib,json,os,stat,sys
path=sys.argv[1]
fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW)
with os.fdopen(fd,"rb") as stream:
    info=os.fstat(stream.fileno())
    if not stat.S_ISREG(info.st_mode) or info.st_size<=0 or info.st_size>2147483648:
        raise ValueError("invalid diagnostic archive")
    print(json.dumps({"size":info.st_size}),flush=True)
    checksum=hashlib.sha256()
    size=0
    while size<info.st_size:
        block=stream.read(min(49152,info.st_size-size))
        if not block: raise ValueError("truncated diagnostic archive")
        size+=len(block)
        checksum.update(block)
        print(base64.b64encode(block).decode("ascii"),flush=True)
    final=os.fstat(stream.fileno())
    if final.st_size!=info.st_size or final.st_mtime_ns!=info.st_mtime_ns:
        raise ValueError("diagnostic archive changed during transfer")
    print(json.dumps({"size":size,"sha256":checksum.hexdigest()}),flush=True)
'''


class TransportError(Exception):
    pass


def _path(value):
    if (not isinstance(value, str) or not value.startswith("/") or "\x00" in value
            or ".." in value.split("/") or os.path.normpath(value) == "/"):
        raise TransportError("invalid diagnostic path")
    return os.path.normpath(value)


def execute(client, command, sink=None, timeout=180, stream_factory=None):
    """Consume stdout incrementally and require the Kubernetes exit status."""
    if stream_factory is None:
        from kubernetes.stream import stream
        stream_factory = stream
    if timeout <= 0:
        raise TransportError("invalid exec timeout")
    websocket = stream_factory(client.client.connect_get_namespaced_pod_exec,
        client.pod_name, client.namespace, command=command, container=client.container_name,
        stderr=True, stdin=False, stdout=True, tty=False, _preload_content=False,
        _request_timeout=timeout)
    captured, errors, size = [], [], 0
    deadline = time.monotonic() + timeout
    def remaining_time():
        remaining = deadline - time.monotonic()
        if remaining <= 0:
            raise TransportError("diagnostic exec timed out")
        return remaining
    def drain():
        nonlocal size
        while websocket.peek_stdout():
            remaining_time()
            block = websocket.read_stdout()
            if sink is not None:
                sink(block)
            else:
                size += len(block)
                if size > 16384:
                    raise TransportError("unexpected exec output size")
                captured.append(block)
        while websocket.peek_stderr():
            remaining_time()
            message = websocket.read_stderr()
            if sum(map(len, errors)) < 4096:
                errors.append(message[:4096])
    try:
        while websocket.is_open():
            websocket.update(timeout=min(1, remaining_time()))
            drain()
        drain()
        remaining_time()
        if websocket.returncode != 0:
            raise TransportError("diagnostic exec failed or returned no exit status")
        return "".join(captured)
    finally:
        websocket.close()


def python_command(client):
    cached = getattr(client, "_lakehouse_python", None)
    if cached:
        return cached
    script = ("for py in /usr/libexec/platform-python python3; do "
              "if command -v \"$py\" >/dev/null 2>&1 && \"$py\" -B -c " + shlex.quote(PROBE_CODE) +
              "; then printf '%s\\n' \"$py\"; exit 0; fi; done; exit 1")
    found = execute(client, ["/bin/sh", "-c", script], timeout=30).strip()
    if found not in ("/usr/libexec/platform-python", "python3"):
        raise TransportError("target requires Python >=3.6 diagnostic standard libraries")
    client._lakehouse_python = found
    return found


def archive_logs(client, source, destination):
    source, destination = _path(source), _path(destination)
    if destination != source + ".tar.gz":
        raise TransportError("archive must belong to its diagnostic temporary directory")
    python = python_command(client)
    # All target containers must pass the Python download probe even when tar
    # is available. Failure of tar itself is never silently retried as success.
    available = execute(client, ["/bin/sh", "-c",
        "if command -v tar >/dev/null && command -v gzip >/dev/null; then printf yes; fi"], timeout=30)
    if available == "yes":
        execute(client, ["tar", "-czf", destination, "-C", os.path.dirname(source), os.path.basename(source)])
    else:
        execute(client, [python, "-B", "-c", PACK_CODE, source, destination])


class Decoder:
    def __init__(self, stream):
        self.stream = stream
        self.pending = ""
        self.expected = None
        self.size = 0
        self.checksum = hashlib.sha256()
        self.complete = False

    def feed(self, chunk):
        if isinstance(chunk, bytes):
            chunk = chunk.decode("ascii")
        self.pending += chunk
        while "\n" in self.pending:
            line, self.pending = self.pending.split("\n", 1)
            if self.complete or not line:
                raise TransportError("unexpected diagnostic stream data")
            if self.expected is None:
                header = json.loads(line)
                size = header.get("size")
                if type(size) is not int or not 0 < size <= MAX_ARCHIVE_BYTES or set(header) != {"size"}:
                    raise TransportError("invalid diagnostic stream header")
                self.expected = size
            elif line.startswith("{"):
                footer = json.loads(line)
                if (footer != {"size": self.size, "sha256": self.checksum.hexdigest()}
                        or self.size != self.expected):
                    raise TransportError("diagnostic stream checksum or size mismatch")
                self.complete = True
            else:
                if len(line) > 65536:
                    raise TransportError("oversized diagnostic stream frame")
                data = base64.b64decode(line, validate=True)
                self.size += len(data)
                if self.size > self.expected:
                    raise TransportError("diagnostic stream exceeds announced size")
                self.checksum.update(data)
                self.stream.write(data)
        if len(self.pending) > 65536:
            raise TransportError("unterminated diagnostic stream frame")

    def finish(self):
        if self.pending or not self.complete:
            raise TransportError("truncated diagnostic stream")


def download(client, remote_path, local_path):
    remote_path, local_path = _path(remote_path), _path(os.path.abspath(local_path))
    python = python_command(client)
    directory = os.path.dirname(local_path)
    if os.path.islink(local_path) or not os.path.isdir(directory):
        raise TransportError("invalid local diagnostic destination")
    descriptor, temporary = tempfile.mkstemp(prefix=".obdiag-download-", dir=directory)
    try:
        with os.fdopen(descriptor, "wb") as output:
            decoder = Decoder(output)
            execute(client, [python, "-B", "-c", DOWNLOAD_CODE, remote_path], sink=decoder.feed, timeout=300)
            decoder.finish()
            output.flush()
            os.fsync(output.fileno())
        os.replace(temporary, local_path)
        return True
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)
