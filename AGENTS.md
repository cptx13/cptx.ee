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
| `main.go` | Startup and routes (dispatch matches the path only, via `pathOnly` in `site_extras.go`, so query strings never 404) |
| `content.go` | Markdown loading, frontmatter, post IDs/URLs (`id:` in frontmatter pins a post's URL ID). Poems are `categories: ["rimas"]` inside `content/posts/croquis/`; `isRima` + `/rimas/` live in `site_extras.go` |
| `site_extras.go` | Headers, gzip, RSS, sitemap, `/healthz`, retired URLs |
| `build.go` | Static build into `dist/` |
| `feeds.go` | Cheatsheet RSS reader (feed cache, refresh limits) |
| `reads.go` | `/addons/reads/`: kept articles (`data/reads.tsv`) + latest from the feed |
| `photos.go`, `pockets_photos.go` | Photo folders, thumbnails, pockets/photos page data |
| `addons.go`, `places.go`, `words.go` | `/addons/` pages |
| `library.go` | `/archives/` and the library |
| `stats.go`, `geo.go`, `static/js/hello.js` | Readers' counters (people via the `/_hello` beacon on first scroll/tap/click/key, no timers; a visit to a post is a read; countries; raw HTML loads shown as "machines"; daily totals in `~/cptx-data/stats/`) and `/addons/readers/` |
| `newsletter.go` + `content/newsletter/` | Newsletter = every email; the weekly ones are issues (`NNN.md` → `/newsletter/<n>/`, Spanish `NNN.es.md` → `/newsletter/<n>/es/`; "issue 01", read-counted, generated cover art) kept on the kiosk `/newsletter/kiosk/`; how it works is the `#newsletter` section of the privacy policy; `scripts/newsletter-draft.py` drafts the weekly issue (en + es); `scripts/newsletter-send.py` (CI job `newsletter`, after deploy-verify) sends each new issue through Buttondown as one bilingual email (repo variable `NEWSLETTER_MODE=send`; `draft` keeps it waiting); `buttondownUser` (`cptx`) in `newsletter.go` opens the sign-up form. Details: DEPLOYMENT.md → Newsletter |
| `search.go` | `/search/?q=`: full text of posts + library, accent-insensitive, server-side |
| `legal.go` + `content/legaldocs/` | `/legaldocs/`: privacy, cookies, how the site works |

Before pushing: `scripts/check.sh` (runs in CI on every push).
