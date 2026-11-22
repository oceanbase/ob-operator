#!/usr/bin/env python3
"""Publish fixed-channel image tools as self-contained, version-pinned scripts."""
import argparse
import hashlib
from pathlib import Path


def build(output, manifest_url, manifest_sha256=None):
    if manifest_sha256 and (len(manifest_sha256) != 64 or any(c not in '0123456789abcdef' for c in manifest_sha256)):
        raise ValueError('manifest SHA256 must contain 64 lowercase hexadecimal characters')
    source = Path(__file__).with_name('offline_common.py').read_text()
    main = "\nif __name__ == '__main__':\n    main()\n"
    if not source.endswith(main):
        raise ValueError('Unexpected offline_common entry point')
    implementation = source[:-len(main)]
    output.mkdir(parents=True, exist_ok=True)
    targets = [output / ('package-offline-' + channel + '.py') for channel in ['compose', 'helm']]
    targets.append(output / 'TOOLS-SHA256SUMS')
    if any(path.exists() for path in targets):
        raise FileExistsError('Release tools already exist; use a new output directory')
    result = []
    for channel in ['compose', 'helm']:
        target = output / ('package-offline-' + channel + '.py')
        content = implementation + ("\nif __name__ == '__main__':\n"
                                    "    main(expected_channel=%r, default_manifest=%r, default_manifest_sha256=%r)\n"
                                    % (channel, manifest_url, manifest_sha256))
        with target.open('x') as stream:
            stream.write(content)
        target.chmod(0o755)
        result.append(hashlib.sha256(target.read_bytes()).hexdigest() + '  ' + target.name)
    checksum = output / 'TOOLS-SHA256SUMS'
    with checksum.open('x') as stream:
        stream.write('\n'.join(result) + '\n')
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--manifest-url', required=True, help='Immutable versioned release-manifest.json URL or local path')
    parser.add_argument('--manifest-sha256', help='Optional trusted manifest SHA256 embedded into the scripts')
    args = parser.parse_args()
    for line in build(args.output, args.manifest_url, args.manifest_sha256):
        print(line)
