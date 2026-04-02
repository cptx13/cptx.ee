# Deployment

cptx.ee is deployed to the `cptx` SSH host as systemd user services under the `exedev` user.

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
