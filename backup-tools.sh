#!/usr/bin/env sh
#shellcheck disable=SC3040 # sh in Alpine does support pipefail
set -euo pipefail

usage() {
  printf '\n'
  printf 'USAGE: %s status      - Report the internal status of the backup process.\n' "$0"
  printf '       %s create-now  - Create a backup immediately, outside the usual schedule.\n' "$0"
  printf '\n'
}

do_curl() {
  METHOD="$1"
  shift
  URL="http://127.0.0.1:8080$1"
  shift
  if ! curl -s --fail-with-body -X "$METHOD" "$URL" "$@"; then
    echo "ERROR: could not $METHOD $URL" >&2
    return 1
  fi
}

cmd_status() {
  do_curl GET /v1/status
}

cmd_create_now() {
  echo "Creating backup... (This might take a while. Check the container log for details.)" >&2
  do_curl POST /v1/backup-now
}

case "${1:-unset}" in
  (help|--help|-h) usage ;;
  (status)         cmd_status ;;
  (create-now)     cmd_create_now ;;
  (*)              usage && exit 1 ;;
esac
