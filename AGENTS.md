# AGENTS.md

## Project overview

cptx.ee is a Go web server that serves a static site built from markdown content using htmlc, dispatch, and goldmark. It listens on port 8080.

## Build

```sh
go build -o cptx.ee .
```

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for the full deployment setup (systemd user units, SSH host `cptx`, mise for Go toolchain management).

## Code map

| File | What lives there |
|---|---|
| `main.go` | Startup and routes |
| `content.go` | Markdown loading, frontmatter, post IDs/URLs (`id:` in frontmatter pins a post's URL ID) |
| `site_extras.go` | Headers, gzip, RSS, sitemap, `/healthz`, retired URLs |
| `build.go` | Static build into `dist/` |
| `feeds.go` | Cheatsheet RSS reader |
| `photos.go`, `pockets_photos.go` | Photo folders, thumbnails, pockets/photos page data |
| `addons.go`, `places.go`, `words.go` | `/addons/` pages |
| `library.go` | `/archives/` and the library |
| `stats.go`, `geo.go`, `static/js/read.js` | Readers' counters (visits, reads, countries; daily totals in `~/cptx-data/stats/`) and `/addons/readers/` |
| `legal.go` + `content/legaldocs/` | `/legaldocs/`: privacy, cookies, how it works, newsletter |

Before pushing: `scripts/check.sh` (runs in CI on every push).
