#!/bin/bash
set -euo pipefail
cd $(dirname $0)/..

UI_IMAGE="${UI_IMAGE:-quay.io/konveyor/tackle2-ui:latest}"
UI_PORT="${UI_PORT:-9000}"
HUB_PORT="${HUB_PORT:-8080}"
SESSION="tackle-local"
CONTAINER_NAME="tackle-ui"
RUN_LOG_FILE="./.run/run_local_with_ui.log"
TMUXP_WORKSPACE="./.run/tmuxp_workspace.yml"

#
# Detect container runtime (prefer podman, fall back to docker). Ensure tmux is installed.
#
if command -v podman &>/dev/null; then
    RUNTIME=podman
elif command -v docker &>/dev/null; then
    RUNTIME=docker
else
    echo "Error: podman or docker is required" >&2
    exit 1
fi

if ! command -v tmux &>/dev/null; then
    echo "Error: tmux is required" >&2
    exit 1
fi

# Podman resolves the host gateway automatically via host.containers.internal.
# Docker on Linux needs an explicit --add-host mapping.
HUB_HOST=host.containers.internal
RUNTIME_EXTRA=""
if [[ "${RUNTIME}" == "docker" ]]; then
    HUB_HOST=host.docker.internal
    RUNTIME_EXTRA="--add-host=host.docker.internal:host-gateway"
fi

#
# Commands for each pane.
#
INFO_CMD="tail -f ${RUN_LOG_FILE}"
HUB_CMD="hack/run_local.sh"
UI_CMD="${RUNTIME} run \
  --rm \
  --name ${CONTAINER_NAME} \
  -p ${UI_PORT}:8080 \
  -e AUTH_REQUIRED=true \
  -e NODE_EXTRA_CA_CERTS=/opt/app-root/src/ca.crt \
  -e OIDC_CLIENT_ID=web-ui \
  -e OIDC_ISSUER=http://${HUB_HOST}:${HUB_PORT}/oidc \
  -e TACKLE_HUB_URL=http://${HUB_HOST}:${HUB_PORT} \
  ${RUNTIME_EXTRA} \
  ${UI_IMAGE}"

#
# Capture run information.
#
mkdir -p "$(dirname ${RUN_LOG_FILE})"
cat > ${RUN_LOG_FILE} <<EOF
  Container runtime: ${RUNTIME}
  Hub: http://localhost:${HUB_PORT}
  UI:  http://localhost:${UI_PORT}
EOF

#
# Clean up any previous run.
#
tmux kill-session -t "${SESSION}" 2>/dev/null || true
${RUNTIME} stop "${CONTAINER_NAME}" 2>/dev/null || true
${RUNTIME} rm "${CONTAINER_NAME}" 2>/dev/null || true

#
# Create a tmux session with info, HUB and UI panes.
#
tmux \
  new-session -s "${SESSION}" "${INFO_CMD}" \; \
  split-window -v -t 0 "${HUB_CMD}" \; \
  split-window -v -t 1 "${UI_CMD}" \; \
  select-pane -t 0 -T "info" \; \
  select-pane -t 1 -T "hub" \; \
  select-pane -t 2 -T "UI container" \; \
  resize-pane -t 0 -y 5 \; \
  resize-pane -t 1 -y 50% \; \
  resize-pane -t 2 -y 50% \; \
  set-option -t "${SESSION}" pane-border-status top \;
