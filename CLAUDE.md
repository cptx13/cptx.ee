# CLAUDE.md

## Project overview

cptx.ee is a Go web server that serves a static site built from markdown content using htmlc, dispatch, and goldmark. It listens on port 8080.

## Build

```sh
go build -o cptx.ee .
```

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for the full deployment setup (systemd user units, SSH host `cptx`, mise for Go toolchain management).
