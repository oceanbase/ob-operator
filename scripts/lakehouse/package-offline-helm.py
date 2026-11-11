#!/usr/bin/env python3
"""Prepare helm images from an explicitly selected locked release."""
from release_offline import main

if __name__ == "__main__":
    main(expected_channel="helm")
