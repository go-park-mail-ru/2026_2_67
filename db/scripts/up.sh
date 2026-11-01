#!/bin/bash

echo_env_var() {
  if [[ -n "$2" ]]; then
    echo "$1: $2"
  else
    echo "Env var '$1' not declared" >&2
    return 1
  fi
  return 0
}

ROOT_DIR=$(realpath "$(dirname "$0")/../../")
MIGRATIONS_DIR="$ROOT_DIR/db/migrations"

set -a
source "$ROOT_DIR/.env"
set +a

if ! echo_env_var "DB_NAME" "$DB_NAME"; then
  exit 1
fi

if ! echo_env_var "DB_USER" "$DB_USER"; then
  exit 1
fi

cd "$MIGRATIONS_DIR" || exit 1

for f in $(find . -type f -name "*.up.sql" | sort); do
  echo "Migrate: $f"
  if ! sudo -u "$DB_USER" psql -d "$DB_NAME" -f "$f"; then
    echo "Error" >&2
  fi
done
