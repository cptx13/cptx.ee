# Deployment

cptx.ee is deployed on the `cptx` VM as systemd user services under the `exedev` user.

## Automatic deploys (normal path)

Just push to `main` on GitHub (web UI, phone, laptop — anywhere). The VM polls
GitHub itself, so nothing needs to reach into the VM:

- **`cptx-ee-deploy.timer`** — fires every minute
- **`cptx-ee-deploy.service`** — runs `deploy.sh`, which fetches `origin/main`
  and, if it moved since the last good deploy, deploys it safely (below).
  If nothing changed it exits immediately.

Watch a deploy: `journalctl --user -u cptx-ee-deploy -n 20`
(problems are logged with `FAILED:` — `journalctl --user -u cptx-ee-deploy | grep FAILED`)

### Safe deploy

1. `git reset --hard` to the new commit and build it with `mise exec -- go build`
   into `~/.local/bin/cptx.ee.new`. The live binary is not touched.
2. **Build fails** → the worktree goes back to the running commit, the site keeps
   serving the old version, and that commit is not retried. Push a fix.
3. **Build passes** → the live binary becomes `~/.local/bin/cptx.ee.prev`, the new
   one is moved into place, and `cptx-ee.service` is restarted.
4. **Health check**: `http://127.0.0.1:8080/` must answer within 90 s. If it doesn't,
   the previous binary and commit are restored automatically and the bad commit
   is not retried.

State lives in `~/.local/state/cptx-ee/` (`deployed-sha`, `prev-sha`, `failed-sha`).
`deploy.sh` no longer uses `cptx-ee-build.service`; it builds directly.

### Rollback

- **Instant (on the VM):** `~/cptx-v2/deploy.sh rollback` — restores the previous
  binary + commit and restarts. Main on GitHub is unchanged, so it stays rolled
  back only until the next push.
- **Permanent:** `git revert <sha> && git push origin main` — redeploys within ~2 min.

Note: `deploy.sh` hard-resets the worktree, so never leave uncommitted or
unpushed work on the VM — it will be discarded.

The `origin` remote uses a tokenless public HTTPS URL, so deploys need no
credentials. Pushing *from* the VM uses a GitHub token stored outside the repo
in `~/.git-credentials` (mode 600) via git's `credential.helper store`.
To set or rotate that token, run interactively on the VM:

```sh
~/set-github-token.sh
```

GitHub Actions cannot be used for deploys: hosted runners can't reach this VM.

## Manual deploy

## Prerequisites

- SSH access to `cptx` (configured in `~/.ssh/config`)
- [mise](https://mise.jdx.dev/) installed on the remote at `~/.local/bin/mise`
- The repository cloned to `~/cptx-v2` on the remote
- The `mise.toml` in the repo must be trusted (`mise trust ~/cptx-v2/mise.toml`)

## Systemd units

All four user units are kept in this repo and installed (copied) to
`~/.config/systemd/user/` on the VM:

| Unit | Role | Enabled |
|---|---|---|
| `cptx-ee.service` | Runs `~/.local/bin/cptx.ee` on :8080 from `~/cptx-v2`; restarts on failure | yes (`default.target`) |
| `cptx-ee-deploy.timer` | Fires `cptx-ee-deploy.service` every minute | yes (`timers.target`) |
| `cptx-ee-deploy.service` | Runs `deploy.sh` (safe build → swap → health check) | via the timer |
| `cptx-ee-build.service` | Manual bootstrap build for a fresh VM only | no |

The server does not depend on the build unit: `deploy.sh` builds and swaps the
binary, and the binary in `~/.local/bin` survives reboots, so a reboot simply
starts the last good build.

## Installing / updating unit files (on the VM)

Unit files are *not* applied by a deploy. After a change to any unit reaches
`main` (and the VM has pulled it), run on the VM:

```sh
export XDG_RUNTIME_DIR=/run/user/$(id -u)
cd ~/cptx-v2
cp cptx-ee.service cptx-ee-build.service cptx-ee-deploy.service cptx-ee-deploy.timer ~/.config/systemd/user/
systemctl --user daemon-reload
systemd-analyze --user verify ~/.config/systemd/user/cptx-ee*.service
```

Check that what is installed matches the repo:

```sh
for f in cptx-ee.service cptx-ee-build.service cptx-ee-deploy.service cptx-ee-deploy.timer; do
  diff -q ~/cptx-v2/$f ~/.config/systemd/user/$f || echo "DIFFERS: $f"; done
```

### Fresh VM bootstrap

```sh
mise trust ~/cptx-v2/mise.toml
cp ~/cptx-v2/cptx-ee*.service ~/cptx-v2/cptx-ee-deploy.timer ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user start cptx-ee-build.service      # first binary
systemctl --user enable --now cptx-ee.service cptx-ee-deploy.timer
```

## Restarting the server

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  systemctl --user restart cptx-ee.service'
```

## Checking status

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  systemctl --user status cptx-ee.service'
```

## Viewing logs

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  journalctl --user -eu cptx-ee.service --no-pager'
```

## Go version

The Go toolchain is pinned in `mise.toml` (currently `1.27.4`); `go.mod` only sets
the minimum (`1.26.1`). To upgrade: change the version in `mise.toml`, push, and
watch `journalctl --user -u cptx-ee-deploy -n 20`. `mise exec` installs the new
toolchain on the first build; if that build fails, the safe deploy keeps the
current version running.
