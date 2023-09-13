#!/usr/bin/env sh
#shellcheck disable=SC3040 # sh in Alpine does support pipefail
set -euo pipefail

usage() {
  printf '\n'
  printf 'USAGE: %s status        - Report the internal status of the backup process.\n' "$0"
  printf '       %s create-now    - Create a backup immediately, outside the usual schedule.\n' "$0"
  printf '       %s list          - List backups that we can restore from.\n' "$0"
  printf '       %s restore <id>  - Restore a backup using an ID from the "list" command.\n' "$0"
  printf '\n'
}

do_curl() {
  METHOD="$1"
  shift
  URL="http://127.0.0.1:8080$1"
  shift

  # when curl succeeds (2xx status), print the response body on stdin;
  # otherwise print the response body on stderr
  rm -f /tmp/curl-output
  if curl --no-progress-meter -o /tmp/curl-output --fail-with-body -X "$METHOD" "$URL" "$@"; then
    cat /tmp/curl-output
  else
    echo "ERROR: could not $METHOD $URL" >&2
    test -f /tmp/curl-output && cat /tmp/curl-output >&2
    return 1
  fi
}

pretty_print_json() {
  jq .
}

cmd_status() {
  do_curl GET /v1/status | pretty_print_json
}

cmd_create_now() {
  echo "Creating backup... (This might take a while. Check the container log for details.)" >&2
  do_curl POST /v1/backup-now | pretty_print_json
}

cmd_list() {
  do_curl GET /v1/backups | jq -r '.[] | "- ID: \u001B[1;36m\(.id)\u001B[0m, taken at \(.readable_time), size \u001B[0;36m\(.total_size_bytes)\u001B[0m bytes, contains database \(.database_names | map("\u001B[0;36m\(.)\u001B[0m") | join(" and "))"'
  printf '\n'
  #shellcheck disable=SC2016 # non-expansion of the backticks is intentional here
  printf 'To restore one of these backups, pick its ID from the list above and run `%s restore <id>`' "$0"
  printf '\n'
}

cmd_restore() {
  BACKUP_ID="$1"
  if [ "$BACKUP_ID" = "unset" ]; then
    usage && exit 1
  fi
  do_curl POST "/v1/restore/${BACKUP_ID}"
}

case "${1:-unset}" in
  (help|--help|-h) usage ;;
  (status)         cmd_status ;;
  (create-now)     cmd_create_now ;;
  (list)           cmd_list ;;
  (restore)        cmd_restore "${2:-unset}" ;;
  (*)              usage && exit 1 ;;
esac
