#!/usr/bin/env bash

set -u -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKER_BIN="${PACKER_BIN:-$ROOT_DIR/bin/packer}"
TEMPLATE_DIR="${TEMPLATE_DIR:-$ROOT_DIR/examples/hcl/sbom-test}"
LOG_DIR="${LOG_DIR:-/tmp/packer-sbom-cases/logs}"
RELEASE_SERVER="${RELEASE_SERVER:-http://localhost:3231}"

if [[ ! -x "$PACKER_BIN" ]]; then
  echo "error: packer binary not executable: $PACKER_BIN" >&2
  exit 1
fi

if [[ ! -d "$TEMPLATE_DIR" ]]; then
  echo "error: template directory not found: $TEMPLATE_DIR" >&2
  exit 1
fi

mkdir -p "$LOG_DIR"
RUN_TS="$(date +%Y%m%d-%H%M%S)"
MASTER_LOG="$LOG_DIR/all-cases-localhost3231-${RUN_TS}.log"
SUMMARY_TSV="$LOG_DIR/summary-${RUN_TS}.tsv"

version_raw="$($PACKER_BIN version 2>/dev/null | awk 'NR==1{print $2}')"
version="${version_raw#v}"
if [[ -z "$version" ]]; then
  echo "error: could not detect packer version from $PACKER_BIN version" >&2
  exit 1
fi

sums_url="$RELEASE_SERVER/packer/$version/packer_${version}_SHA256SUMS"

start_local_release_server() {
  if [[ "$RELEASE_SERVER" != "http://localhost:3231" && "$RELEASE_SERVER" != "http://127.0.0.1:3231" ]]; then
    return
  fi

  if curl -sS --max-time 3 "$sums_url" >/dev/null 2>&1; then
    return
  fi

  echo "[info] starting local release server on localhost:3231" | tee -a "$MASTER_LOG"
  (
    cd "$ROOT_DIR" || exit 1
    python3 scripts/local-release-server.py --port 3231
  ) >"$LOG_DIR/local-release-server-${RUN_TS}.log" 2>&1 &
  server_pid=$!

  for _ in {1..20}; do
    if curl -sS --max-time 3 "$sums_url" >/dev/null 2>&1; then
      echo "[info] local release server ready (pid=$server_pid)" | tee -a "$MASTER_LOG"
      return
    fi
    sleep 1
  done

  echo "error: local release server did not become ready; see $LOG_DIR/local-release-server-${RUN_TS}.log" | tee -a "$MASTER_LOG" >&2
  exit 1
}

start_local_release_server

echo -e "template\texit_code\tcase_log" > "$SUMMARY_TSV"
echo "Run started at $(date)" | tee -a "$MASTER_LOG"
echo "PACKER_BIN=$PACKER_BIN" | tee -a "$MASTER_LOG"
echo "TEMPLATE_DIR=$TEMPLATE_DIR" | tee -a "$MASTER_LOG"
echo "RELEASE_SERVER=$RELEASE_SERVER" | tee -a "$MASTER_LOG"
echo "SUMMARY_TSV=$SUMMARY_TSV" | tee -a "$MASTER_LOG"
echo | tee -a "$MASTER_LOG"

mapfile -t templates < <(find "$TEMPLATE_DIR" -maxdepth 1 -type f -name '*.pkr.hcl' | sort)
if [[ ${#templates[@]} -eq 0 ]]; then
  echo "error: no templates found in $TEMPLATE_DIR" | tee -a "$MASTER_LOG" >&2
  exit 1
fi

for f in "${templates[@]}"; do
  name="$(basename "$f" .pkr.hcl)"
  case_log="$LOG_DIR/${name}-${RUN_TS}.log"

  echo "===== RUNNING: $f =====" | tee -a "$MASTER_LOG" "$case_log"

  set +e
  PACKER_RELEASE_SERVER="$RELEASE_SERVER" "$PACKER_BIN" build "$f" 2>&1 | tee -a "$MASTER_LOG" "$case_log"
  code=${PIPESTATUS[0]}
  set -e

  echo "===== EXIT: $code : $f =====" | tee -a "$MASTER_LOG" "$case_log"
  echo -e "$f\t$code\t$case_log" >> "$SUMMARY_TSV"
  echo | tee -a "$MASTER_LOG" "$case_log"
done

echo "Run finished at $(date)" | tee -a "$MASTER_LOG"
echo "MASTER_LOG=$MASTER_LOG" | tee -a "$MASTER_LOG"
echo "SUMMARY_TSV=$SUMMARY_TSV" | tee -a "$MASTER_LOG"
