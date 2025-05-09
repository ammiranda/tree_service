#!/bin/sh
# wait-for-it.sh

set -e

# Parse host and port from the first argument
host_port="$1"
host="${host_port%:*}"
port="${host_port#*:}"
shift
cmd="$@"

# Maximum number of retries
max_retries=30
retries=0

# Use environment variables for database connection
until PGPASSWORD=$DB_PASSWORD psql -h "$host" -p "$port" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  retries=$((retries + 1))
  if [ $retries -ge $max_retries ]; then
    >&2 echo "Postgres is still unavailable after $max_retries retries - giving up"
    exit 1
  fi
  >&2 echo "Postgres is unavailable - sleeping (attempt $retries/$max_retries)"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec $cmd 