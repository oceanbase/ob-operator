#!/usr/bin/env python3
"""Require a closed, checksummed set of local RPM inputs before installation."""
import hashlib
import re
import sys
from pathlib import Path


def verify(directory):
    directory = Path(directory)
    expected = {}
    for line in (directory / 'SHA256SUMS').read_text().splitlines():
        match = re.fullmatch(r'([0-9a-fA-F]{64}) [ *]([A-Za-z0-9][A-Za-z0-9._+-]*\.rpm)', line)
        if not match or match[2] in expected:
            raise ValueError('Invalid or duplicate RPM checksum entry')
        expected[match[2]] = match[1].lower()
    actual = {p.name for p in directory.glob('*.rpm')}
    if not expected or actual != set(expected):
        raise ValueError('RPM directory and SHA256SUMS must contain exactly the same files')
    for name, checksum in expected.items():
        path = directory / name
        if path.is_symlink() or not path.is_file():
            raise ValueError('RPM input must be a regular file: ' + name)
        digest = hashlib.sha256()
        with path.open('rb') as stream:
            for chunk in iter(lambda: stream.read(1024 * 1024), b''):
                digest.update(chunk)
        if digest.hexdigest() != checksum:
            raise ValueError('RPM checksum mismatch: ' + name)


if __name__ == '__main__':
    verify(sys.argv[1])
