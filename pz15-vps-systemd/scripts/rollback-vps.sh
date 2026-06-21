#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   sudo ./scripts/rollback-vps.sh

SERVICE_NAME="tasks"
SERVICE_USER="tasksuser"
APP_BIN="/opt/tasks/tasks"
BACKUP_BIN="/opt/tasks/tasks.old"
TIMESTAMP="$(date +%Y%m%d%H%M%S)"

require_root() {
	if [[ "${EUID}" -ne 0 ]]; then
		echo "Run this script with sudo on the VPS."
		exit 1
	fi
}

require_root

if [[ ! -f "${BACKUP_BIN}" ]]; then
	echo "Backup binary not found: ${BACKUP_BIN}"
	exit 1
fi

systemctl stop "${SERVICE_NAME}"

if [[ -f "${APP_BIN}" ]]; then
	cp -a "${APP_BIN}" "${APP_BIN}.failed.${TIMESTAMP}"
fi

install -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${BACKUP_BIN}" "${APP_BIN}"
systemctl start "${SERVICE_NAME}"
systemctl status "${SERVICE_NAME}" --no-pager
