import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("online_builder", HERE / "build-online-package.py")
BUILDER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(BUILDER)


class OnlinePackageTest(unittest.TestCase):
    def test_light_bundle_contains_tools_but_no_image_payload_or_private_state(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source, baseline = root / "source", root / "offline"
            for folder in [source / "scripts/lakehouse", baseline / "charts", baseline / "values", baseline / "prerequisites", baseline / "images", baseline / ".lakehouse-install"]:
                folder.mkdir(parents=True, exist_ok=True)
            (source / "LICENSE").write_text("fixture license")
            (source / "install.sh").write_text("#!/bin/sh\nexit 0\n")
            (source / "scripts/lakehouse/ONLINE-INSTALL.txt").write_text("ONLINE_BUNDLE_NAME ONLINE_ARCHIVE_NAME")
            (baseline / ".lakehouse-install/secret").write_text("NEVER_COPY_THIS")
            payload = b"large-image-placeholder-not-to-be-packaged"
            digest = hashlib.sha256(payload).hexdigest()
            (baseline / "images/core.tar.gz").write_bytes(payload)
            (baseline / "images/archives.json").write_text(json.dumps({"schemaVersion":1,"archives":[{"name":"core","base":"bundle","path":"images/core.tar.gz","sha256":digest,"references":["fixture:1"]}]}))
            (baseline / "image-lock.json").write_text(json.dumps({"source_sha":"f"*40,"images":[{"role":"observer","reference":"fixture:1","image_id":"sha256:"+"a"*64}]}))
            executable = root / "tool"
            header = bytearray(64); header[:5]=b"\x7fELF\x02"; header[18]=62
            executable.write_bytes(header)
            license_ = root / "LICENSE"; license_.write_text("fixture")
            args = argparse.Namespace(source=source,source_archive=None,baseline=baseline,installer=executable,helm=executable,jq=executable,
                                      helm_license=license_,jq_license=license_,output=root/"output",version="test",archive_url="http://127.0.0.1/archives",helm_version="3-test",jq_version="1-test")
            with patch.object(BUILDER.BASE,"source_identity",return_value=("a"*40,1000000000)):
                BUILDER.build(args)
            record=json.loads((args.output/"build-result.json").read_text())
            self.assertEqual(record["image_payload_files"],0)
            custom=root/"custom-client";custom.mkdir()
            client=custom/"kubectl"
            client.write_text("#!/bin/sh\nprintf 'custom kubectl %s' \"$1\"\n")
            client.chmod(0o755)
            shim=next(args.output.glob("*/tools/bin/kubectl"))
            result=subprocess.check_output(["/bin/bash",str(shim),"version"],env={**os.environ,"PATH":str(shim.parent)+os.pathsep+str(custom)},text=True)
            self.assertEqual(result,"custom kubectl version")
            with tarfile.open(record["package"]) as archive:
                names=archive.getnames()
                self.assertTrue(any(n.endswith("tools/bin/helm") for n in names))
                self.assertTrue(any(n.endswith("tools/bin/jq") for n in names))
                self.assertFalse(any(".lakehouse-install" in n or n.endswith("core.tar.gz") for n in names))
                name=next(n for n in names if n.endswith("images/archives.json"))
                manifest=json.load(archive.extractfile(name))["archives"][0]
                self.assertEqual(manifest["path"],"images/cache/"+digest+".tar.gz")
                self.assertEqual(manifest["size_bytes"],len(payload))
                self.assertEqual(manifest["url"],"http://127.0.0.1/archives/"+digest+".tar.gz")


if __name__ == "__main__":
    unittest.main()
