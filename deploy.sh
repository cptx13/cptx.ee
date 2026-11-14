#!/usr/bin/env bash
# cptx.ee deploy — run every minute by cptx-ee-deploy.timer.
#
#   deploy.sh            poll GitHub; if main moved, deploy it safely
#   deploy.sh rollback   go back to the previous good version right now
#
# Safe deploy (B-01):
#   1. build the new commit to a temp path (cptx.ee.new) — the live binary is untouched
#   2. build fails  -> worktree goes back to the running commit; site stays up
#   3. build passes -> live binary becomes cptx.ee.prev, new one is swapped in, restart
#   4. health check -> if the new version doesn't answer, roll back automatically
# A commit that failed is not retried; the next push is tried as normal.
# Look for "FAILED" in: journalctl --user -u cptx-ee-deploy
set -euo pipefail

REPO="${CPTX_REPO:-$HOME/cptx-v2}"
BIN="${CPTX_BIN:-$HOME/.local/bin/cptx.ee}"
STATE="${CPTX_STATE_DIR:-$HOME/.local/state/cptx-ee}"
MISE="${MISE:-$HOME/.local/bin/mise}"
HEALTH_URL="${CPTX_HEALTH_URL:-http://127.0.0.1:8080/}"
HEALTH_WAIT="${CPTX_HEALTH_WAIT:-90}"   # seconds; startup renders the whole static site
SERVICE=cptx-ee.service

export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
mkdir -p "$STATE"
exec 9>"$STATE/lock"
flock -n 9 || { echo "another deploy is running; skipping"; exit 0; }
cd "$REPO"

short() { git rev-parse --short "$1"; }
read_state() { cat "$STATE/$1" 2>/dev/null || true; }

healthy() {
  local waited=0
  while [ "$waited" -lt "$HEALTH_WAIT" ]; do
    if systemctl --user is-active --quiet "$SERVICE" &&
       curl -fsS -o /dev/null --max-time 5 "$HEALTH_URL" 2>/dev/null; then
      return 0
    fi
    sleep 2; waited=$((waited + 2))
  done
  return 1
}

# Put back the previous binary + worktree and restart. $1 = good sha, $2 = bad sha
restore() {
  local good="$1" bad="$2"
  git reset --quiet --hard "$good"
  if [ -f "$BIN.prev" ]; then
    mv -f "$BIN" "$BIN.bad" 2>/dev/null || true
    mv -f "$BIN.prev" "$BIN"
  else
    echo "FAILED: no $BIN.prev to restore; restarting current binary on $(short "$good")"
  fi
  systemctl --user restart "$SERVICE" 9>&-   # do not leak the deploy lock
  echo "$good" > "$STATE/deployed-sha"
  echo "$bad"  > "$STATE/failed-sha"     # don't redeploy the bad commit until main moves again
  if healthy; then
    echo "rolled back to $(short "$good") — site healthy"
  else
    echo "FAILED: rolled back to $(short "$good") but the site is still not answering — check journalctl --user -eu $SERVICE"
  fi
}

# First run of this script: whatever is checked out is what's running.
[ -f "$STATE/deployed-sha" ] || git rev-parse HEAD > "$STATE/deployed-sha"
deployed=$(read_state deployed-sha)

# ---- manual rollback ------------------------------------------------------
if [ "${1:-}" = "rollback" ]; then
  prev=$(read_state prev-sha)
  [ -n "$prev" ] || { echo "FAILED: no previous deploy recorded; use git revert + push"; exit 1; }
  echo "manual rollback $(short "$deployed") -> $(short "$prev")"
  restore "$prev" "$deployed"
  echo "note: GitHub main still has $(short "$deployed"); it won't redeploy until the next push."
  echo "      to make it permanent: git revert <sha> && git push origin main"
  exit 0
fi

# ---- normal poll ----------------------------------------------------------
git fetch --quiet origin main
remote=$(git rev-parse origin/main)
[ "$remote" = "$deployed" ] && exit 0
[ "$remote" = "$(read_state failed-sha)" ] && exit 0

echo "deploying $(short "$deployed") -> $(short "$remote")"
git reset --quiet --hard "$remote"

# 1. build to a temp path
rm -f "$BIN.new"
if ! "$MISE" exec -- go build -o "$BIN.new" . ; then
  rm -f "$BIN.new"
  git reset --quiet --hard "$deployed"
  echo "$remote" > "$STATE/failed-sha"
  echo "FAILED: build of $(short "$remote") failed — still serving $(short "$deployed"); push a fix"
  exit 0
fi

# 2. swap (keep the running binary as .prev) and restart
[ -f "$BIN" ] && mv -f "$BIN" "$BIN.prev"
mv -f "$BIN.new" "$BIN"
systemctl --user restart "$SERVICE" 9>&-   # do not leak the deploy lock

# 3. health check, auto-rollback if the new version doesn't come up
if healthy; then
  echo "$deployed" > "$STATE/prev-sha"
  echo "$remote"   > "$STATE/deployed-sha"
  rm -f "$STATE/failed-sha" "$BIN.bad"
  echo "deployed $(short "$remote")"
else
  echo "FAILED: $(short "$remote") did not answer on $HEALTH_URL within ${HEALTH_WAIT}s — rolling back"
  restore "$deployed" "$remote"
fi
