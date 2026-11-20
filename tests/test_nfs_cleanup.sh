#!/usr/bin/env bash

set -eu
cd "$(dirname "$0")"
source util.sh
unset NFS_BASE_PATH
source env.sh
[[ -z "$NFS_BASE_PATH" ]]

scratch=$(mktemp -d)
trap 'rm -rf -- "$scratch"' EXIT
export trace="$scratch/trace" USER_NAME=tester NFS_SERVER=example.invalid
# Only record the files selected by the remote shell. No files are deleted.
mkdir "$scratch/bin"
cat > "$scratch/bin/sudo" <<'STUB'
#!/bin/sh
printf '%s\n' "$(pwd -P)" "$@" > "$trace"
STUB
chmod +x "$scratch/bin/sudo"
export PATH="$scratch/bin:$PATH"
ssh() {
    [[ $# == 3 && $1 == -- && $2 == tester@example.invalid ]]
    touch "$scratch/ssh-called"
    sh -c "$3"
}

for NFS_BASE_PATH in '' '/' '///' '/./' 'relative/path' "$scratch/../"; do
    if clean_nfs_test_directory 2>/dev/null; then exit 1; fi
done
[[ ! -e "$scratch/ssh-called" ]]

NFS_BASE_PATH="$scratch/root-link"
ln -s / "$NFS_BASE_PATH"
if clean_nfs_test_directory; then exit 1; fi
[[ ! -e "$trace" ]]

NFS_BASE_PATH="$scratch/"'测试 spaces and quotes " '\'' [wildcard]* $HOME'
mkdir -- "$NFS_BASE_PATH"
touch -- "$NFS_BASE_PATH/payload"
clean_nfs_test_directory
printf '%s\n' "$NFS_BASE_PATH" rm -rf -- ./payload | diff -u - "$trace"
