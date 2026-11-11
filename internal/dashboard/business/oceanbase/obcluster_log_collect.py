"""Dashboard collector for the pinned obdiag 5.0.1 result format (Python >=3.6).

obdiag can exit zero after writing a summary containing failed nodes. Validate
the complete result before publishing an attachment; never archive credentials.
"""
import gzip
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile


class CollectionError(Exception):
    pass


def targets(directory):
    with open(os.path.join(directory, "nodes.json")) as stream:
        items = json.load(stream)
    if not isinstance(items, list) or not items:
        raise CollectionError("target manifest is empty")
    names = []
    for item in items:
        if not isinstance(item, dict):
            raise CollectionError("invalid target manifest")
        namespace, pod = item.get("namespace"), item.get("pod_name")
        if not all(isinstance(value, str) and re.fullmatch(r"[a-z0-9][a-z0-9.-]*", value)
                   for value in (namespace, pod)):
            raise CollectionError("invalid target identity")
        # obdiag v5.0.1 KubernetesClient.get_name and SshClient.get_name.
        names.append(("kubernetes_" + namespace + "_" + pod).replace(".", "_"))
    if len(names) != len(set(names)):
        raise CollectionError("duplicate target identity")
    return set(names)


def summary_rows(path):
    with open(path, encoding="utf-8") as stream:
        lines = stream.read().splitlines()
    header = False
    rows = []
    for line in lines:
        line = line.strip()
        if not line or (line.startswith("+") and set(line) <= set("+-")):
            continue
        if not (line.startswith("|") and line.endswith("|")):
            raise CollectionError("unrecognized obdiag result summary")
        columns = [value.strip() for value in line[1:-1].split("|")]
        if not header:
            if len(columns) == 1 and columns[0].startswith("Gather observer Log Summary on "):
                continue
            if columns != ["Node", "Status", "Size", "info"]:
                raise CollectionError("unrecognized obdiag result columns")
            header = True
            continue
        if len(columns) != 4:
            raise CollectionError("invalid obdiag result row")
        rows.append(columns)
    if not header or not rows:
        raise CollectionError("obdiag result has no node rows")
    return rows


def validate_payload(path):
    count = 0
    with tarfile.open(path, "r:gz") as archive:
        for member in archive:
            parts = member.name.split("/")
            if member.name.startswith("/") or ".." in parts or not (member.isfile() or member.isdir()):
                raise CollectionError("unsafe node archive member")
            if not member.isfile():
                continue
            stream = archive.extractfile(member)
            compressed = parts[-1].endswith(".gz")
            if compressed:
                stream = gzip.GzipFile(fileobj=stream)
            with stream:
                size = 0
                while True:
                    data = stream.read(1024 * 1024)
                    if not data:
                        break
                    size += len(data)
            if not compressed and size != member.size:
                raise CollectionError("truncated node archive")
            # zstd requires a non-stdlib decoder. Keep such files in the bundle,
            # but require at least one verifiably non-empty plain/gzip log.
            if size and not parts[-1].endswith(".zst") and re.fullmatch(r"(?:observer|rootservice|election|alert)\.log(?:\..+)?", parts[-1]):
                count += 1
        # Consume the gzip trailer as well, detecting truncated/corrupt streams.
        while archive.fileobj.read(1024 * 1024):
            pass
    if count == 0:
        raise CollectionError("node archive has no non-empty OceanBase log")
    return count


def validate_results(directory, expected):
    packs = [os.path.join(directory, name) for name in os.listdir(directory)
             if name.startswith("obdiag_gather_pack_")]
    if len(packs) != 1 or not os.path.isdir(packs[0]) or os.path.islink(packs[0]):
        raise CollectionError("expected exactly one obdiag result directory")
    pack = packs[0]
    rows = summary_rows(os.path.join(pack, "result_summary.txt"))
    seen, files, log_count = set(), [], 0
    for node, status, size, info in rows:
        if node not in expected or node in seen:
            raise CollectionError("unexpected or duplicate node result")
        seen.add(node)
        if status != "Success" or not info.startswith("file save in "):
            raise CollectionError("obdiag reported a failed node; no attachment published")
        payload = info[len("file save in "):]
        filename = os.path.basename(payload)
        if (os.path.dirname(os.path.abspath(payload)) != os.path.abspath(pack)
                or not filename.startswith("observer_log_" + node + "_")
                or not filename.endswith(".tar.gz") or os.path.islink(payload)
                or not os.path.isfile(payload)):
            raise CollectionError("missing or invalid node payload")
        log_count += validate_payload(payload)
        files.append(payload)
    if seen != expected:
        raise CollectionError("obdiag omitted one or more target nodes")
    # Publish only the validated files, never config.yaml, kubeconfig or stdout.
    return [os.path.join(pack, "result_summary.txt")] + files, log_count


def collect(start, end, directory, attachment):
    os.umask(0o077)
    temporary = None
    try:
        expected = targets(directory)
        with open(os.path.join(directory, "collector-output.log"), "wb") as output:
            status = subprocess.call([
                "obdiag", "gather", "log", "--from", start, "--to", end,
                "--store_dir", directory, "-c", os.path.join(directory, "config.yaml")
            ], stdout=output, stderr=subprocess.STDOUT, timeout=1500)
        if status != 0:
            raise CollectionError("obdiag command failed; no attachment published")
        files, log_count = validate_results(directory, expected)
        descriptor, temporary = tempfile.mkstemp(prefix=".log-attachment-", dir=os.path.dirname(attachment))
        os.close(descriptor)
        with tarfile.open(temporary, "w:gz") as archive:
            for path in files:
                archive.add(path, arcname=os.path.relpath(path, directory), recursive=False)
        with open(temporary, "rb") as stream:
            os.fsync(stream.fileno())
        # Cleanup must succeed before publication: a failed Job must not leave
        # a downloadable attachment or report success while retaining secrets.
        shutil.rmtree(directory)
        os.replace(temporary, attachment)
        temporary = None
        print("Log collection succeeded: {} nodes, {} non-empty log files".format(len(expected), log_count))
    finally:
        if temporary and os.path.exists(temporary):
            os.unlink(temporary)
        # Runs on unsuccessful summaries, archive failures and process failures.
        # The output directory belongs exclusively to this one-shot Job.
        shutil.rmtree(directory, ignore_errors=True)


if __name__ == "__main__":
    try:
        collect(*sys.argv[1:])
    except CollectionError as error:
        print("Log collection failed: " + str(error), file=sys.stderr)
        sys.exit(1)
    except Exception:
        # Do not disclose subprocess output, paths or configuration values.
        print("Log collection failed: invalid output or local I/O failure; no attachment published", file=sys.stderr)
        sys.exit(1)
