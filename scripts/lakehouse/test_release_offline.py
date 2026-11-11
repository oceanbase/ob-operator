"""Release metadata and offline identity regressions; runtime tests use real archives."""
import argparse
import hashlib
import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import release_offline as offline


class ReleaseOfflineTests(unittest.TestCase):
    def fixture(self, root):
        image = {'reference': 'example/busybox@sha256:' + 'a' * 64,
                 'image_id': 'sha256:' + 'b' * 64}
        asset = root / 'compose.yaml'
        asset.write_text('services: {}\n')
        document = {'schemaVersion': 1, 'releaseVersion': 'test', 'platform': 'linux/amd64',
                    'channels': {'compose': {'images': [image], 'configuration': {
                        'url': asset.name, 'sha256': hashlib.sha256(asset.read_bytes()).hexdigest()}}}}
        manifest = root / 'release-manifest.json'
        manifest.write_text(json.dumps(document))
        return manifest, document, image

    def test_manifest_and_asset_hashes_are_independently_checked(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            path, document, image = self.fixture(root)
            args = argparse.Namespace(release_manifest=str(path),
                                      manifest_sha256=hashlib.sha256(path.read_bytes()).hexdigest(),
                                      expected_channel='compose')
            bundle, lock, release = offline.manifest_bundle(args)
            try:
                self.assertEqual(lock['images'], [image])
                self.assertEqual((bundle / 'compose.yaml').stat().st_mode & 0o777, 0o600)
                self.assertEqual(release['sha256'], args.manifest_sha256)
            finally:
                args._release_context.cleanup()
            (root / 'compose.yaml').write_text('tampered')
            with self.assertRaisesRegex(ValueError, 'artifact SHA256 mismatch'):
                offline.manifest_bundle(args)
            args._release_context.cleanup()
            args.manifest_sha256 = '0' * 64
            with self.assertRaisesRegex(ValueError, 'manifest SHA256 mismatch'):
                offline.manifest_bundle(args)

    def test_conflicting_identity_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path, document, image = self.fixture(Path(directory))
            document['channels']['compose']['images'].append({**image, 'image_id': 'sha256:' + 'c' * 64})
            path.write_text(json.dumps(document))
            args = argparse.Namespace(release_manifest=str(path), expected_channel='compose')
            with self.assertRaisesRegex(ValueError, 'Conflicting image identity'):
                offline.manifest_bundle(args)

    def test_remote_manifest_cannot_read_local_assets(self):
        with self.assertRaisesRegex(ValueError, 'cannot reference local files'):
            offline.resolve_asset('https://example.invalid/release.json', 'file:///etc/passwd')

    def test_digest_aliases_are_explicit_and_never_pull(self):
        reference = 'example/busybox@sha256:' + 'a' * 64
        selected = {'channel': 'compose', 'images': [{'reference': reference, 'image_id': 'sha256:' + 'b' * 64}],
                    '_compose_config': {'services': {'test': {'image': reference}}}}
        aliases = offline.offline_references(selected)
        rendered = offline.yaml.safe_load(offline.offline_compose(selected, aliases))
        self.assertEqual(rendered['services']['test']['image'], aliases[reference])
        self.assertEqual(rendered['services']['test']['pull_policy'], 'never')
        self.assertEqual(selected['_compose_config']['services']['test']['image'], reference)
        self.assertEqual(offline.offline_references({**selected, 'channel': 'helm'}), {})

    def test_docker_descriptor_requires_verified_original_config(self):
        item = {'reference': 'example:1', 'image_id': 'sha256:' + 'a' * 64}
        descriptor = 'sha256:' + 'b' * 64
        actual = {'Id': descriptor, 'Descriptor': {'digest': descriptor}, 'Os': 'linux', 'Architecture': 'amd64'}
        with mock.patch.object(offline, 'run', return_value=json.dumps([actual])):
            with mock.patch.object(offline, 'saved_docker_config_matches', return_value=False):
                with self.assertRaisesRegex(ValueError, 'identity mismatch'):
                    offline.inspect_image('docker', item, 'linux/amd64')
            with mock.patch.object(offline, 'saved_docker_config_matches', return_value=True) as verified:
                self.assertEqual(offline.inspect_image('docker', item, 'linux/amd64'), actual)
                verified.assert_called_once_with('docker', 'example:1', item['image_id'], 'linux/amd64')

    def test_k3s_loader_checks_config_and_original_manifest(self):
        item = {'reference': 'example/test@sha256:' + 'a' * 64, 'image_id': 'sha256:' + 'b' * 64}
        script = offline.k3s_verification_script([item], 'linux/amd64')
        self.assertIn(item['image_id'] + '|linux/amd64', script)
        self.assertIn('Loaded K3s reference/manifest digest mismatch', script)
        self.assertIn('CONTAINERD_ADDRESS', script)

    def test_private_values_snapshot(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'values.yaml'
            offline.write_private(path, b'fixture: value\n')
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)
            with self.assertRaises(FileExistsError):
                offline.write_private(path, b'overwrite')


if __name__ == '__main__':
    unittest.main()
