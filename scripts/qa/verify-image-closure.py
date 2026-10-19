#!/usr/bin/env python3
"""Build-time QA bundle gate. Requires PyYAML; never pulls or runs an image.

Pass the image lock and ALL rendered manifests (Operator, lakehouse stage=cluster,
Dashboard, cert-manager). Optionally verify saved Docker archives as well.
MinIO/mc remain online-only; this gate must not silently bundle their images.
"""
import argparse
import json
import re
import tarfile
from pathlib import Path

import yaml


def image_refs(value):
    if isinstance(value, dict):
        for key, child in value.items():
            if key == "image" and isinstance(child, str):
                yield child
            else:
                yield from image_refs(child)
    elif isinstance(value, list):
        for child in value:
            yield from image_refs(child)


def verify(lock, manifests, archives=()):
    inventory = {}
    for item in lock["images"]:
        ref = item["reference"]
        if ref in inventory:
            raise ValueError(f"Duplicate image reference: {ref}")
        if not re.fullmatch(r"sha256:[0-9a-f]{64}", item["image_id"]):
            raise ValueError(f"Missing immutable image ID: {ref}")
        delivery = item["delivery"]
        if delivery not in ("archive", "online-or-matching-existing-cache"):
            raise ValueError(f"Unknown delivery mode: {ref}")
        if item["role"] in ("minio", "mc") and delivery == "archive":
            raise ValueError("MinIO and mc must remain online-only in this QA bundle")
        inventory[ref] = item
    required = set()
    for manifest in manifests:
        for doc in yaml.safe_load_all(manifest):
            required.update(image_refs(doc))
    if not required:
        raise ValueError("No rendered image references supplied")
    missing = required - inventory.keys()
    if missing:
        raise ValueError("Images missing from lock: " + ", ".join(sorted(missing)))
    if archives:
        saved = {}
        for path in archives:
            with tarfile.open(path, "r|*") as archive:
                for member in archive:
                    if member.name != "manifest.json":
                        continue
                    for item in json.load(archive.extractfile(member)):
                        config = Path(item["Config"]).name
                        if config.endswith(".json"):
                            config = config[:-5]
                        for ref in item.get("RepoTags") or []:
                            if ref in saved and saved[ref] != config:
                                raise ValueError(f"Conflicting saved image: {ref}")
                            saved[ref] = config
        for ref, item in inventory.items():
            if item["delivery"] == "archive":
                if saved.get(ref) != item["image_id"][7:]:
                    raise ValueError(f"Missing/mismatched image in Docker archives: {ref}")
            elif ref in saved:
                raise ValueError(f"Online-only image unexpectedly bundled: {ref}")
        extra = saved.keys() - inventory.keys()
        if extra:
            raise ValueError("Untracked archived images: " + ", ".join(sorted(extra)))
    return len(required)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("lock", type=Path)
    parser.add_argument("manifests", type=Path, nargs="+")
    parser.add_argument("--archive", type=Path, action="append", default=[])
    args = parser.parse_args()
    try:
        count = verify(json.loads(args.lock.read_text()),
                       [path.read_text() for path in args.manifests], args.archive)
    except (ValueError, KeyError, OSError, yaml.YAMLError, tarfile.TarError) as error:
        parser.exit(1, f"Image closure FAILED: {error}\n")
    print(f"PASS image closure: {count} rendered image references accounted for")
