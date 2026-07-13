#!/usr/bin/env bash
#
# Measures cimcli's import time and validate time (import+validate combined)
# against the RealGrid-Merged CGMES test configuration, using wall-clock
# timing around the actual CLI binary (not `go test -bench`) so the numbers
# reflect what a real invocation of the tool costs.
#
# Usage:
#   scripts/bench_realgrid_cli.sh [-n repeat] [-d realgrid-dir]
#
# Defaults to 3 repetitions and CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged.

set -euo pipefail

repeat=3
realgrid_dir=""

while getopts "n:d:h" opt; do
	case "$opt" in
	n) repeat="$OPTARG" ;;
	d) realgrid_dir="$OPTARG" ;;
	h)
		echo "Usage: $0 [-n repeat] [-d realgrid-dir]"
		exit 0
		;;
	*)
		echo "Usage: $0 [-n repeat] [-d realgrid-dir]" >&2
		exit 1
		;;
	esac
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [ -z "$realgrid_dir" ]; then
	realgrid_dir="$repo_root/CGMES-Test-Configurations/v3.0/RealGrid/RealGrid-Merged"
fi

if [ ! -d "$realgrid_dir" ]; then
	echo "RealGrid directory not found: $realgrid_dir" >&2
	echo "(did you run 'git submodule update --init --recursive'?)" >&2
	exit 1
fi

mapfile -t files < <(find "$realgrid_dir" -maxdepth 1 -name '*.xml' | sort)
if [ "${#files[@]}" -eq 0 ]; then
	echo "No .xml files found in $realgrid_dir" >&2
	exit 1
fi

echo "Building cimcli..."
bin_dir="$(mktemp -d)"
trap 'rm -rf "$bin_dir"' EXIT
(cd "$repo_root" && go build -o "$bin_dir/cimcli" ./cmd/cimcli)
cimcli="$bin_dir/cimcli"

echo "Files (${#files[@]}):"
printf '  %s\n' "${files[@]}"
echo

now_ns() { date +%s%N; }

import_ms=()
validate_ms=()

for ((i = 1; i <= repeat; i++)); do
	echo "Run $i/$repeat"

	start=$(now_ns)
	if ! "$cimcli" import "${files[@]}" >/dev/null; then
		echo "  cimcli import failed" >&2
		exit 1
	fi
	end=$(now_ns)
	import_elapsed=$(((end - start) / 1000000))
	import_ms+=("$import_elapsed")
	echo "  import:            ${import_elapsed} ms"

	start=$(now_ns)
	# validate exits 1 when sh:Violation-severity findings are present, which
	# is expected for RealGrid — only treat a crash (exit >1) as a failure.
	set +e
	"$cimcli" validate "${files[@]}" >/dev/null
	status=$?
	set -e
	if [ "$status" -gt 1 ]; then
		echo "  cimcli validate failed (exit $status)" >&2
		exit 1
	fi
	end=$(now_ns)
	validate_elapsed=$(((end - start) / 1000000))
	validate_ms+=("$validate_elapsed")
	echo "  import+validate:   ${validate_elapsed} ms"

	validation_only=$((validate_elapsed - import_elapsed))
	echo "  validate only (derived): ${validation_only} ms"
	echo
done

avg() {
	local sum=0
	for v in "$@"; do sum=$((sum + v)); done
	echo $((sum / $#))
}

avg_import=$(avg "${import_ms[@]}")
avg_validate=$(avg "${validate_ms[@]}")
avg_validate_only=$((avg_validate - avg_import))

echo "=== Average over $repeat run(s) ==="
echo "import:                  ${avg_import} ms"
echo "import+validate:         ${avg_validate} ms"
echo "validate only (derived): ${avg_validate_only} ms"
