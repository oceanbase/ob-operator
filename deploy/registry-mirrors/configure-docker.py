#!/usr/bin/env python3
"""Preview or explicitly apply optional Docker Hub mirror configuration."""

import argparse
import json
import os
from pathlib import Path
import stat
import sys
import tempfile
from urllib.parse import urlsplit


DEFAULT_MIRRORS = [
    "https://mirror.ccs.tencentyun.com",
    "https://registry.docker-cn.com",
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://docker.1panel.live",
    "https://docker.1ms.run",
    "https://dytt.online",
    "https://lispy.org",
    "https://docker.xiaogenban1993.com",
    "https://docker.yomansunter.com",
    "https://666860.xyz",
    "https://a.ussh.net",
    "https://hub.rat.dev",
    "https://docker.m.daocloud.io",
]


def unique_keys(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON key: " + key)
        result[key] = value
    return result


def read_config(path):
    if path.is_symlink():
        raise ValueError("refusing a symbolic link config: " + str(path))
    try:
        metadata = path.stat()
    except FileNotFoundError:
        return None, None, {}
    if not stat.S_ISREG(metadata.st_mode):
        raise ValueError("config must be a regular file: " + str(path))
    original = path.read_bytes()
    config = json.loads(original, object_pairs_hook=unique_keys)
    if not isinstance(config, dict):
        raise ValueError("Docker configuration must be a JSON object")
    return original, metadata, config


def merge_config(config, requested_mirrors):
    merged = dict(config)
    mirrors = merged.get("registry-mirrors", [])
    if not isinstance(mirrors, list) or any(
        not isinstance(item, str) or not item.strip() for item in mirrors
    ):
        raise ValueError("registry-mirrors must be an array of nonempty strings")
    mirrors = list(mirrors)
    for mirror in requested_mirrors:
        url = urlsplit(mirror)
        if (url.scheme != "https" or not url.hostname or url.username is not None
                or url.password is not None or url.query or url.fragment
                or any(character.isspace() for character in mirror)):
            raise ValueError("new mirrors must be HTTPS URLs without credentials or queries")
        mirror = mirror.rstrip("/")
        if mirror not in [item.rstrip("/") for item in mirrors]:
            mirrors.append(mirror)
    merged["registry-mirrors"] = mirrors
    for key, default in (("max-concurrent-downloads", 10), ("max-concurrent-uploads", 5)):
        value = merged.setdefault(key, default)
        if type(value) is not int or value < 1:
            raise ValueError(key + " must be a positive integer")
    return merged


def write_config(path, content, original, metadata):
    if not path.parent.is_dir():
        raise ValueError("config directory does not exist: " + str(path.parent))
    current, _, _ = read_config(path)
    if current != original:
        raise ValueError("config changed while preparing the update; preview again")
    backup = None
    if original is not None:
        fd, backup = tempfile.mkstemp(prefix=path.name + ".backup-", dir=path.parent)
        with os.fdopen(fd, "wb") as stream:
            stream.write(original)
            stream.flush()
            os.fsync(stream.fileno())
    fd, temporary = tempfile.mkstemp(prefix="." + path.name + ".", dir=path.parent)
    try:
        with os.fdopen(fd, "wb") as stream:
            if metadata is not None:
                os.fchmod(stream.fileno(), stat.S_IMODE(metadata.st_mode))
                current_stat = os.fstat(stream.fileno())
                if (current_stat.st_uid, current_stat.st_gid) != (metadata.st_uid, metadata.st_gid):
                    os.fchown(stream.fileno(), metadata.st_uid, metadata.st_gid)
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)
    return backup


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", type=Path, default=Path("/etc/docker/daemon.json"),
                        help="Docker daemon configuration path")
    parser.add_argument("--mirror", action="append", dest="mirrors",
                        help="HTTPS Docker Hub mirror; repeat to replace the default 14-address list")
    parser.add_argument("--apply", action="store_true",
                        help="write the merged config with a backup; never reloads Docker")
    args = parser.parse_args()
    try:
        original, metadata, config = read_config(args.config)
        merged = merge_config(config, args.mirrors or DEFAULT_MIRRORS)
        content = (json.dumps(merged, indent=2) + "\n").encode("utf-8")
        if not args.apply:
            sys.stdout.write(content.decode("utf-8"))
            print("Preview only; no files changed. Add --apply to write this configuration.", file=sys.stderr)
        elif config == merged and original is not None:
            print("Configuration already contains these settings; no files changed.")
        else:
            backup = write_config(args.config, content, original, metadata)
            print("Updated " + str(args.config))
            if backup:
                print("Backup: " + str(backup))
            print("Docker has not been reloaded. Validate the configuration before reloading.")
    except (OSError, ValueError) as error:
        print("Error: " + str(error), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
