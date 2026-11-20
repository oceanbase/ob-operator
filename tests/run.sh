#!/usr/bin/env bash
set -eo pipefail

if [[ $# != 1 ]]; then
    echo "Usage: $0 CASE_DIRECTORY" >&2
    exit 2
fi
case_dir=$1
if [[ -d "$case_dir" ]]; then case_dir=$(cd -- "$case_dir" && pwd); fi
cd -- "$(dirname -- "$0")"
source env.sh

shopt -s nullglob
cases=("$case_dir"/check_P*.sh)
if (( ${#cases[@]} == 0 )); then
    echo "No scenarios found in $case_dir" >&2
    exit 2
fi

result=0
for scenario in "${cases[@]}"; do
    date
    echo "Running $scenario"
    # Legacy scenarios can report failed assertions and then exit zero.
    if ! bash "$scenario" 2>&1 | awk '
      { print; fflush(); line = tolower($0) }
      line ~ /^[[:space:]]*case[ -]+fail(ed)?([:[:space:]]|$)/ { failed = 1 }
      line ~ /^[[:space:]]*case[ -]+pass(ed)?([:[:space:]]|$)/ { passed = 1 }
      END { if (failed || !passed) exit 1 }
    '; then
        echo "Scenario failed or did not report completion: $scenario" >&2
        result=1
    fi
done
exit "$result"
