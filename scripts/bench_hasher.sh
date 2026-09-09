#!/usr/bin/env bash
# SHA-256 vs BLAKE3 benches + ns/op discrepancy table.
# Usage: ./scripts/bench_hasher.sh [count]
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COUNT="${1:-1}"
FILTER='BenchmarkDigestSum256|BenchmarkNode_hash|BenchmarkMutableTree_SetWorkingHash'
OUT="${BENCH_OUT:-$ROOT/hasher.bench.txt}"
: > "$OUT"

echo "GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) GOMAXPROCS=${GOMAXPROCS:-}" | tee -a "$OUT"

run_pkg() {
  local dir="$1"
  echo "=== $dir ===" | tee -a "$OUT"
  (cd "$dir" && go test -mod=mod -run '^$' -bench "$FILTER" -benchmem -count="$COUNT" -timeout 20m .) | tee -a "$OUT"
}

run_pkg "$ROOT"
run_pkg "$ROOT/v2"

python3 - "$OUT" <<'PY'
import re, sys
from collections import defaultdict

path = sys.argv[1]
pat = re.compile(
    r"^Benchmark(\S+?)/(sha256|blake3)/(\S+?)(?:-\d+)?\s+\d+\s+([0-9.]+)\s+ns/op"
)
pkg_re = re.compile(r"^pkg:\s+(\S+)")
acc = defaultdict(list)
pkg = "v1"
for line in open(path):
    line = line.strip()
    mpkg = pkg_re.match(line)
    if mpkg:
        pkg = "v2" if mpkg.group(1).endswith("/v2") else "v1"
        continue
    m = pat.match(line)
    if not m:
        continue
    bench, algo, rest, ns = m.group(1), m.group(2), m.group(3), float(m.group(4))
    acc[(pkg, bench, rest, algo)].append(ns)

keys = sorted({(p, b, r) for p, b, r, _ in acc})
print()
print("hash algo discrepancy (median ns/op; ratio = blake3/sha256; <1 means BLAKE3 faster)")
print(f"{'bench':<48} {'sha256':>12} {'blake3':>12} {'ratio':>8}")
print("-" * 84)
for pkg, bench, rest in keys:
    sha = sorted(acc.get((pkg, bench, rest, "sha256"), []))
    b3 = sorted(acc.get((pkg, bench, rest, "blake3"), []))
    if not sha or not b3:
        continue
    med = lambda xs: xs[len(xs) // 2]
    s, b = med(sha), med(b3)
    ratio = b / s if s else float("nan")
    name = f"{pkg}/{bench}/{rest}"
    print(f"{name:<48} {s:12.2f} {b:12.2f} {ratio:8.3f}")
PY
