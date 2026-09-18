#!/bin/bash
set -e
echo "Pulling from GitHub..."
git pull
echo "Building and restarting service..."
export XDG_RUNTIME_DIR=/run/user/$(id -u)
systemctl --user restart cptx-ee-build.service
systemctl --user restart cptx-ee.service
echo "✓ Deployed"
