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

There are two user-level systemd units, kept in this repo and installed to `~/.config/systemd/user/` on the remote:

- **`cptx-ee-build.service`** — oneshot unit that builds the Go binary using `mise exec` (so the Go version from `mise.toml` is used). The binary is placed at `~/.local/bin/cptx.ee`, outside the worktree.
- **`cptx-ee.service`** — runs the server on port 8080. Depends on the build unit. Restarts on failure.

## Deploying updated unit files

```sh
scp cptx-ee-build.service cptx-ee.service cptx:/tmp/
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  mv /tmp/cptx-ee-build.service /tmp/cptx-ee.service ~/.config/systemd/user/ && \
  systemctl --user daemon-reload'
```

## Building and starting the service

```sh
ssh cptx 'export XDG_RUNTIME_DIR=/run/user/$(id -u) && \
  systemctl --user start cptx-ee-build.service && \
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
