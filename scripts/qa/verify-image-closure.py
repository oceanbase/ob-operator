#!/usr/bin/env python3
"""Build-time QA bundle gate. Requires PyYAML; never pulls or runs an image.

Pass the image lock and ALL rendered manifests (Operator, lakehouse stage=cluster,
Dashboard, cert-manager). --bundle validates the self-contained archive manifest
and internal checksums. Explicit archive/sidecar arguments support older bundles.
ConfigMap inspection images are included; optional SQL analyzer is not.
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


def verify(lock, manifests, archives=(), sidecar_archives=(), archive_references=None):
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
                actual = archive_image_ids(path)
                if archive_references is not None and set(actual) != archive_references[path]:
                    raise ValueError("Archive references differ from bundle manifest: " + str(path))
                for ref, image_id in actual.items():
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


def bundle_file(bundle, name):
    """Require an existing regular file reached without leaving the bundle."""
    relative = Path(name)
    if relative.is_absolute() or ".." in relative.parts or not relative.parts:
        raise ValueError("Unsafe bundle path: " + name)
    path = bundle
    for part in relative.parts:
        path = path / part
        if path.is_symlink():
            raise ValueError("Bundle path must not contain a symlink: " + name)
    if not path.is_file():
        raise ValueError("Missing bundle file: " + name)
    return path


def verify_bundle(bundle, manifests):
    """Verify a fresh single-tar extraction, without any external image inputs."""
    bundle = Path(bundle).resolve()
    checksums = {}
    for line in bundle_file(bundle, "SHA256SUMS").read_text().splitlines():
        expected, separator, name = line.partition("  ")
        if not separator or not re.fullmatch(r"[0-9a-f]{64}", expected) or name in checksums:
            raise ValueError("Invalid or duplicate bundle checksum entry")
        path = bundle_file(bundle, name)
        digest = hashlib.sha256()
        with path.open("rb") as source:
            for chunk in iter(lambda: source.read(8 * 1024 * 1024), b""):
                digest.update(chunk)
        if digest.hexdigest() != expected:
            raise ValueError("Bundle checksum mismatch: " + name)
        checksums[name] = expected
    for name in ("image-lock.json", "images/archives.json"):
        if name not in checksums:
            raise ValueError("Bundle metadata is not covered by SHA256SUMS: " + name)
    lock = json.loads(bundle_file(bundle, "image-lock.json").read_text())
    if any(item["delivery"] != "archive" for item in lock["images"]):
        raise ValueError("Single-tar bundle requires archive delivery for every image")
    saved = json.loads(bundle_file(bundle, "images/archives.json").read_text())
    if saved.get("schemaVersion") != 1 or not saved.get("archives"):
        raise ValueError("Expected schemaVersion 1 and a nonempty bundle archive manifest")
    references, names, all_refs = {}, set(), set()
    for item in saved["archives"]:
        if item["base"] != "bundle":
            raise ValueError("Single-tar archive must use base=bundle")
        path = bundle_file(bundle, item["path"])
        if item["name"] in names or path in references:
            raise ValueError("Duplicate bundle archive name or path")
        if not re.fullmatch(r"[0-9a-f]{64}", item["sha256"]) or checksums.get(item["path"]) != item["sha256"]:
            raise ValueError("Archive SHA256 differs from verified bundle checksum: " + item["path"])
        declared = item["references"]
        if not isinstance(declared, list) or not declared or not all(isinstance(ref, str) and ref for ref in declared):
            raise ValueError("Expected nonempty archive references")
        refs = {normalize_ref(ref) for ref in declared}
        if len(refs) != len(declared) or refs & all_refs:
            raise ValueError("Duplicate bundle archive reference")
        names.add(item["name"])
        all_refs.update(refs)
        references[path] = refs
    return verify(lock, manifests, list(references), archive_references=references)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("lock", type=Path)
    parser.add_argument("manifests", type=Path, nargs="+")
    parser.add_argument("--archive", type=Path, action="append", default=[])
    parser.add_argument("--sidecar-archive", type=Path, action="append", default=[])
    parser.add_argument("--bundle", type=Path, help="Verify all checksums and archives inside a single-tar extraction")
    args = parser.parse_args()
    try:
        manifests = [path.read_text() for path in args.manifests]
        if args.bundle:
            if args.archive or args.sidecar_archive or args.lock.resolve() != (args.bundle / "image-lock.json").resolve():
                raise ValueError("--bundle requires its own image-lock.json and no external archive arguments")
            count = verify_bundle(args.bundle, manifests)
        else:
            count = verify(json.loads(args.lock.read_text()), manifests, args.archive, args.sidecar_archive)
    except (ValueError, KeyError, OSError, yaml.YAMLError, tarfile.TarError) as error:
        parser.exit(1, f"Image closure FAILED: {error}\n")
    print(f"PASS image closure: {count} rendered image references accounted for")
