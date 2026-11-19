#!/usr/bin/env bash
set -eu
entry=$(cd -- "$(dirname -- "$0")" && pwd)/run.sh
scratch=$(mktemp -d)
trap 'rm -r -- "$scratch"' EXIT
cd "$scratch"
mkdir -p 'cases space' empty
d='cases space'
verify() {
  local expected="$1" label="$2" rc=0
  bash "$entry" "$d" > "$scratch/out" 2>&1 || rc=$?
  if [[ $rc != "$expected" ]]; then cat "$scratch/out"; exit 1; fi
  echo "PASS $label (rc=$rc)"
}
printf '%s\n' 'echo "case passed: SQL readback"' > "$d/check_P0_ok.sh"
verify 0 success
printf '%s\n' 'echo "case failed"' 'echo "case passed"' > "$d/check_P0_bad.sh"
verify 1 legacy_failure_with_trailing_success
grep -q 'SQL readback' "$scratch/out"
printf '%s\n' 'echo "case passed"' 'exit 27' > "$d/check_P0_bad.sh"
verify 1 nonzero_script
printf '%s\n' 'echo "did not reach checks"' > "$d/check_P0_bad.sh"
verify 1 no_completion
printf '%s\n' 'echo "case-pass"' > "$d/check_P0_bad.sh"
verify 0 historical_marker
d=empty
verify 2 empty_suite
