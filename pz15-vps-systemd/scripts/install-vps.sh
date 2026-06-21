#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   sudo TASKS_BINARY=/tmp/tasks TASKS_ENV=/tmp/tasks.env TASKS_UNIT=/tmp/tasks.service ./scripts/install-vps.sh

SERVICE_NAME="tasks"
SERVICE_USER="tasksuser"
APP_DIR="/opt/tasks"
CONFIG_DIR="/etc/tasks"
APP_BIN="${APP_DIR}/tasks"
ENV_FILE="${CONFIG_DIR}/tasks.env"
UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

BINARY_SRC="${TASKS_BINARY:-/tmp/tasks}"
ENV_SRC="${TASKS_ENV:-${PROJECT_DIR}/deploy/env/tasks.env.example}"
UNIT_SRC="${TASKS_UNIT:-${PROJECT_DIR}/deploy/systemd/tasks.service}"
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

backup_file() {
	local target="$1"
	if [[ -e "${target}" ]]; then
		local backup="${target}.bak.${TIMESTAMP}"
		cp -a "${target}" "${backup}"
		echo "Backup created: ${backup}"
	fi
}

require_root
require_file "${BINARY_SRC}"
require_file "${ENV_SRC}"
require_file "${UNIT_SRC}"

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
	useradd --system --no-create-home --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

install -d -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${APP_DIR}"
install -d -o root -g root -m 0755 "${CONFIG_DIR}"

backup_file "${APP_BIN}"
install -o "${SERVICE_USER}" -g "${SERVICE_USER}" -m 0755 "${BINARY_SRC}" "${APP_BIN}"

backup_file "${ENV_FILE}"
install -o root -g root -m 0600 "${ENV_SRC}" "${ENV_FILE}"

backup_file "${UNIT_FILE}"
install -o root -g root -m 0644 "${UNIT_SRC}" "${UNIT_FILE}"

systemctl daemon-reload

if ! systemctl enable --now "${SERVICE_NAME}"; then
	echo "Service failed to start. Check status and logs below."
	systemctl status "${SERVICE_NAME}" --no-pager || true
	exit 1
fi

systemctl status "${SERVICE_NAME}" --no-pager
