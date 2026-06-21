#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   sudo TASKS_BINARY=/tmp/tasks ./scripts/update-vps.sh

SERVICE_NAME="tasks"
SERVICE_USER="tasksuser"
APP_BIN="/opt/tasks/tasks"
BACKUP_BIN="/opt/tasks/tasks.old"
BINARY_SRC="${TASKS_BINARY:-/tmp/tasks}"
TIMESTAMP="$(date +%Y%m%d%H%M%S)"

require_root() {
	if [[ "${EUID}" -ne 0 ]]; then
		echo "Run this script with sudo on the VPS."
		exit 1
	fi
}

require_file() {
	local path="$1"
	if [[ ! -f "${path}" ]]; then
		echo "Required file not found: ${path}"
		exit 1
	fi
}

require_root
require_file "${BINARY_SRC}"
require_file "${APP_BIN}"

systemctl stop "${SERVICE_NAME}"

if [[ -e "${BACKUP_BIN}" ]]; then
	mv "${BACKUP_BIN}" "${BACKUP_BIN}.bak.${TIMESTAMP}"
fi
cp -a "${APP_BIN}" "${BACKUP_BIN}"

install -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${BINARY_SRC}" "${APP_BIN}"

if ! systemctl start "${SERVICE_NAME}"; then
	echo "Service failed to start after update. Run rollback-vps.sh to restore ${BACKUP_BIN}."
	systemctl status "${SERVICE_NAME}" --no-pager || true
	exit 1
fi

if ! systemctl is-active --quiet "${SERVICE_NAME}"; then
	echo "Service is not active after update. Run rollback-vps.sh to restore ${BACKUP_BIN}."
	systemctl status "${SERVICE_NAME}" --no-pager || true
	exit 1
fi

systemctl status "${SERVICE_NAME}" --no-pager
