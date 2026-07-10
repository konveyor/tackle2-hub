#!/bin/bash
set -euo pipefail
cd $(dirname $0)/..

#
# Make the seed resources available
#
SEED_PROJECT=konveyor/tackle2-seed
SEED_REF=${SEED_REF:-main}
SEED_ROOT=./.run/seed-${SEED_REF}
SEED_RESOURCES=${SEED_ROOT}/resources

if [[ ! -d "${SEED_ROOT}" || ! -d "${SEED_ROOT}/.git" ]]; then
  echo "Cloning seed ${SEED_PROJECT} to ${SEED_ROOT}"
  git clone --no-single-branch https://github.com/${SEED_PROJECT} ${SEED_ROOT}
fi
echo "Updating seed ${SEED_ROOT} to ${SEED_REF}"
pushd ${SEED_ROOT}
git fetch --all --tags -p
git checkout ${SEED_REF}
popd

#
# Setup HUB environment variables and run the hub
#
echo "Ensure directories exist for DB and BUCKET"
mkdir -p ./.run/dev ./.run/bucket

export DISCONNECTED=true
export AUTH_REQUIRED=true
export API_PORT=8080
export DB_SEED_PATH=${SEED_RESOURCES}
export DB_PATH=./.run/dev/hub.db
export BUCKET_PATH=./.run/bucket

echo "Running HUB"
set +e
if ! make run; then
  if [ -n "${TMUX:-}" ] || [ ! -t 0 ]; then
    echo ""
    echo "Press any key to exit..."
    read -n 1 -s -r
  fi
  exit 1
fi

