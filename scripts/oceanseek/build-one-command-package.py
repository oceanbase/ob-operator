#!/usr/bin/env python3
"""Retired builder for the old all-in-one Helm archive."""

import argparse


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    _, _ = parser.parse_known_args()
    parser.error(
        'this source no longer builds the legacy Helm bundle. '
        'Use build-helm-release.py --lock <reviewed-image-lock> --output <directory> '
        'and prepare shared prerequisites separately. Already-built legacy bundles '
        'can still be used as migration inputs.'
    )


if __name__ == '__main__':
    main()
