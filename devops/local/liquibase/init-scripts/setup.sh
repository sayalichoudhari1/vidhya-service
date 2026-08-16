#!/usr/bin/env bash
set -e

# ===================================================================== fns ===

function tranform_file {
  local FILE_PATH=${1}
  local TMP_FILE=$(mktemp)
  cp --attributes-only --preserve "${FILE_PATH}" "${TMP_FILE}"
  cat "${FILE_PATH}" | envsubst '$DB_HOST,$DB_PORT,$DB_NAME,$DB_USER,$DB_PASS,$DB_MASTER_USER,$DB_MASTER_PASS' >  "${TMP_FILE}" \
    && mv "${TMP_FILE}" "${FILE_PATH}" >/dev/null
}

function list_dir_recurse {
  local DIR_PATH=${1}
  find "${DIR_PATH}" -type f 2>/dev/null
}

function print_array {
  local ITEMS=$@
  for item in "${ITEMS[*]}"; do
    echo "$item"
  done
}

# ================================================================== script ===

DB_ARTIFACT_PATH=${1:-/liquibase/artifacts}
DB_ARTIFACT_STAGE_PATH=${2:-/mnt/artifacts}

if [ -d "${DB_ARTIFACT_STAGE_PATH}" ]; then
  /bin/cp -rf "${DB_ARTIFACT_STAGE_PATH}" "${DB_ARTIFACT_PATH}"
fi

if [[ -f "${DB_ARTIFACT_PATH}/liquibase.properties" ]]; then
  echo "artifact liquibase files to transform..."
  ls -la "${DB_ARTIFACT_PATH}"/liquibase.properties || true
  echo ""

  tranform_file "${DB_ARTIFACT_PATH}/liquibase.properties"
fi

CHANGELOGS=$(list_dir_recurse "${DB_ARTIFACT_PATH}"/changelogs/)
if [[ "${#CHANGELOGS[@]}" -ne "0" ]]; then
  echo "artifact changelogs to transform..."
  print_array "${CHANGELOGS[*]}"
  echo ""

  for f in ${CHANGELOGS[@]}; do
    tranform_file "$f"
  done
fi

SCRIPTS=$(list_dir_recurse "${DB_ARTIFACT_PATH}"/scripts/)
if [[ "${#SCRIPTS[@]}" -ne "0" ]]; then
  echo "artifact scripts files to transform..."
  print_array "${SCRIPTS[*]}"
  echo ""

  for f in ${SCRIPTS[@]}; do
    tranform_file "$f"
  done
fi

