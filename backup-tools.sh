#!/usr/bin/env ash

# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0
#
# shellcheck shell=dash

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
  if curl --no-progress-meter -o /tmp/curl-output --fail-with-body -X "$METHOD" -u "$PGSQL_USER:$PGPASSWORD" "$URL" "$@"; then
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
  if [ "${PGSQL_USER:-unset}" = backup ]; then
    # when using the postgresql-ng chart, this container has a restricted user
    # with read-only privileges that are enough for creating backups, but not
    # for restoring them
    echo "Please enter credentials for the PostgreSQL superuser to proceed."
    echo -n "Username [postgres]: "
    read -r PG_SUPERUSER_NAME
    if [ "$PG_SUPERUSER_NAME" = "" ]; then
      PG_SUPERUSER_NAME=postgres
    fi
    echo -n "Password (get this from the respective Kubernetes Secret or by exec'ing in the postgresql Pod and running: \`cat /postgres-password; echo\`): "
    read -r PG_SUPERUSER_PASSWORD
    if [ "$PG_SUPERUSER_PASSWORD" = "" ]; then
      echo "ERROR: No password given." >&2
      exit 1
    fi
    # to ensure that special characters in the input do not break the JSON payload,
    # use jq to convert the raw input into string literals
    PG_SUPERUSER_NAME_JSON="$(echo -n "$PG_SUPERUSER_NAME" | jq --raw-input --slurp .)"
    PG_SUPERUSER_PASSWORD_JSON="$(echo -n "$PG_SUPERUSER_PASSWORD" | jq --raw-input --slurp .)"
    do_curl POST "/v1/restore/${BACKUP_ID}" -d "{\"superuser\":{\"username\":$PG_SUPERUSER_NAME_JSON,\"password\":$PG_SUPERUSER_PASSWORD_JSON}}"

    echo "If the restore was successful and you are using postgres-ng, you must now restart the postgresql pod to apply proper permissions!"
    echo
    echo "If the restore failed, check if any other pod is holding an active postgres connection and preventing the database from being dropped."
  else
    # when using the legacy postgres chart, this container uses the superuser
    # credentials, so restore works directly
    # TODO: remove this case once everyone has been migrated to postgresql-ng
    do_curl POST "/v1/restore/${BACKUP_ID}" -d "{\"superuser\":null}"
  fi
}

case "${1:-unset}" in
  (help|--help|-h) usage ;;
  (status)         cmd_status ;;
  (create-now)     cmd_create_now ;;
  (list)           cmd_list ;;
  (restore)        cmd_restore "${2:-unset}" ;;
  (*)              usage && exit 1 ;;
esac
