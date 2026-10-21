#!/usr/bin/env python3
"""Build-time QA bundle gate. Requires PyYAML; never pulls or runs an image.

Pass the image lock and ALL rendered manifests (Operator, lakehouse stage=cluster,
Dashboard, cert-manager). Optionally verify saved Docker archives as well.
MinIO/mc remain separate delivery sidecars; diagnostics are part of the main
bundle. ConfigMap inspection images are included; optional SQL analyzer is not.
"""
import argparse
import hashlib
import json
import re
import tarfile
from pathlib import Path

import yaml


def normalize_ref(value):
    value = value[10:] if value.startswith("docker.io/") else value
    return value[8:] if value.startswith("library/") else value


def archive_image_ids(path):
    """Read actual Docker-save config bytes, not tag-only or manifest digests.

    Streaming traversal avoids retaining the large image layers in memory. Config
    JSON and manifest.json are the only payloads read into memory.
    """
    configs, manifests = {}, []
    with tarfile.open(str(path), "r|*") as archive:
        for member in archive:
            name = member.name.lstrip("./")
            if not member.isfile():
                continue
            if name == "manifest.json":
                if member.size > 8 * 1024 * 1024:
                    raise ValueError("Unreasonably large Docker-save manifest")
                manifests.append(json.load(archive.extractfile(member)))
            elif member.size <= 16 * 1024 * 1024 and (name.endswith(".json") or name.startswith("blobs/sha256/")):
                # Docker 26's save format uses OCI-style blob paths, while
                # older save files name configs <sha>.json. Small layers are
                # discarded immediately instead of accumulating their bytes.
                payload = archive.extractfile(member).read()
                try:
                    metadata = json.loads(payload)
                except (ValueError, UnicodeDecodeError):
                    continue
                if isinstance(metadata, dict) and "architecture" in metadata and "os" in metadata:
                    configs[name] = {"id": "sha256:" + hashlib.sha256(payload).hexdigest(),
                                     "architecture": metadata["architecture"], "os": metadata["os"]}
    if len(manifests) != 1 or not isinstance(manifests[0], list):
        raise ValueError("Expected one Docker-save manifest.json in " + str(path))
    images = {}
    for entry in manifests[0]:
        config_name = entry["Config"].lstrip("./")
        if config_name not in configs:
            raise ValueError("Missing image config %s in %s" % (config_name, path))
        metadata = configs[config_name]
        if metadata.get("architecture") != "amd64" or metadata.get("os") != "linux":
            raise ValueError("Image archive is not Linux/amd64: " + config_name)
        image_id = metadata["id"]
        for reference in entry.get("RepoTags") or []:
            reference = normalize_ref(reference)
            if reference in images and images[reference] != image_id:
                raise ValueError("Conflicting image tag in archive: " + reference)
            images[reference] = image_id
    return images


def image_refs(value):
    if isinstance(value, dict):
        if value.get("kind") == "ConfigMap":
            embedded = (value.get("data") or {}).get("config.yaml")
            if isinstance(embedded, str):
                configuration = yaml.safe_load(embedded)
                if isinstance(configuration, dict):
                    yield from image_refs(configuration.get("inspection", {}))
        for key, child in value.items():
            if key == "image" and isinstance(child, str):
                yield child
            else:
                yield from image_refs(child)
    elif isinstance(value, list):
        for child in value:
            yield from image_refs(child)


def verify(lock, manifests, archives=(), sidecar_archives=()):
    inventory = {}
    for item in lock["images"]:
        ref = normalize_ref(item["reference"])
        if ref in inventory:
            raise ValueError(f"Duplicate image reference: {ref}")
        if not re.fullmatch(r"sha256:[0-9a-f]{64}", item["image_id"]):
            raise ValueError(f"Missing immutable image ID: {ref}")
        delivery = item["delivery"]
        if delivery not in ("archive", "delivery-sidecar", "online-or-matching-existing-cache"):
            raise ValueError(f"Unknown delivery mode: {ref}")
        if item["role"] in ("minio", "mc") and delivery == "archive":
            raise ValueError("MinIO and mc must remain outside the main archive; use delivery-sidecar")
        inventory[ref] = item
    required = set()
    for manifest in manifests:
        for doc in yaml.safe_load_all(manifest):
            required.update(normalize_ref(ref) for ref in image_refs(doc))
    if not required:
        raise ValueError("No rendered image references supplied")
    missing = required - inventory.keys()
    if missing:
        raise ValueError("Images missing from lock: " + ", ".join(sorted(missing)))
    if archives or sidecar_archives:
        inventories = {"archive": {}, "delivery-sidecar": {}}
        for mode, paths in (("archive", archives), ("delivery-sidecar", sidecar_archives)):
            saved = inventories[mode]
            for path in paths:
                for ref, image_id in archive_image_ids(path).items():
                    if ref in saved and saved[ref] != image_id:
                        raise ValueError(f"Conflicting saved image: {ref}")
                    saved[ref] = image_id
        for ref, item in inventory.items():
            mode = item["delivery"]
            if mode in inventories:
                if inventories[mode].get(ref) != item["image_id"]:
                    raise ValueError(f"Missing/mismatched image in {mode} Docker archives: {ref}")
                other = "delivery-sidecar" if mode == "archive" else "archive"
                if ref in inventories[other]:
                    raise ValueError(f"Image stored in wrong delivery archive: {ref}")
            elif any(ref in saved for saved in inventories.values()):
                raise ValueError(f"Online-only image unexpectedly bundled: {ref}")
        extra = set().union(*inventories.values()) - inventory.keys()
        if extra:
            raise ValueError("Untracked archived images: " + ", ".join(sorted(extra)))
    return len(required)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("lock", type=Path)
    parser.add_argument("manifests", type=Path, nargs="+")
    parser.add_argument("--archive", type=Path, action="append", default=[])
    parser.add_argument("--sidecar-archive", type=Path, action="append", default=[])
    args = parser.parse_args()
    try:
        count = verify(json.loads(args.lock.read_text()),
                       [path.read_text() for path in args.manifests], args.archive, args.sidecar_archive)
    except (ValueError, KeyError, OSError, yaml.YAMLError, tarfile.TarError) as error:
        parser.exit(1, f"Image closure FAILED: {error}\n")
    print(f"PASS image closure: {count} rendered image references accounted for")
