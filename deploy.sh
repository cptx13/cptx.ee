#!/usr/bin/env bash
# Poll GitHub; if main moved, rebuild and restart the server.
set -euo pipefail
cd "$HOME/cptx-v2"
export XDG_RUNTIME_DIR=/run/user/$(id -u)

git fetch --quiet origin main
local_sha=$(git rev-parse HEAD)
remote_sha=$(git rev-parse origin/main)
[ "$local_sha" = "$remote_sha" ] && exit 0

echo "deploying $local_sha -> $remote_sha"
git reset --hard origin/main
systemctl --user restart cptx-ee-build.service
systemctl --user restart cptx-ee.service
echo "deployed $(git rev-parse --short HEAD)"
