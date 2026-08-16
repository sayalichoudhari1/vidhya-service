#!/usr/bin/env bash

if [ type "${1}" > /dev/null 2>&1 ]; then
  ## first argument is an actual OS command so run it
  exec "$@"
else
  /dbtool/scripts/setup.sh
  /dbtool/scripts/wait-for-it.sh ${DB_HOST}:${DB_PORT} -t 120

  if [ $? -eq 0 ]; then
    exec /liquibase/docker-entrypoint.sh \
      "--logLevel=DEBUG" \
      "--defaultsFile=./artifacts/liquibase.properties" \
      "--changeLogFile=./artifacts/dbchangelog.xml" \
      "$@"
  fi
fi
