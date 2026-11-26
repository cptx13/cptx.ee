# cptx.ee

[![check](https://github.com/cptx13/cptx.ee/actions/workflows/check.yml/badge.svg)](https://github.com/cptx13/cptx.ee/actions/workflows/check.yml)

The source of **[cptx.ee](https://cptx.ee)**, a personal site of short prose, poems, daily memories, photos and the books on our shelves.

> Over the last decade, I have accumulated an obscene amount of digital notes. The official count hovers around **3,000**, reflecting only what I consider somewhat readable; the true number remains unknown to me, and yet, they somehow hold some value, they are true mathoms.

Read the rest in the **[manifesto](https://cptx.ee/manifesto/)**.

## What's on the site

| Section | What it is |
|---|---|
| [croquis](https://cptx.ee/categories/croquis/) | Short prose |
| [rimas](https://cptx.ee/rimas/) | Poems |
| [recuerdos](https://cptx.ee/categories/recuerdos/) | Daily memories, 2021–2024 (a closed series of 1,000+ entries) |
| [photos](https://cptx.ee/photos/) | Photographs by place and year |
| [addons](https://cptx.ee/addons/) | Playful views of the archive: a heatmap of every post, on this day, city lights, most used words, the articles I kept next to the latest from the blogs I follow, and readers |
| [pockets](https://cptx.ee/pockets/) | Things I carry around: books I'm reading, links, and the workbench, my projects with a guide each |
| [archives](https://cptx.ee/archives/) | The books I wrote, a library of the books Merzeede and I own, our vinyl records and our films |
| [newsletter](https://cptx.ee/newsletter/) | A weekly issue by email, in English and Spanish; every issue is kept on the [kiosk](https://cptx.ee/newsletter/kiosk/) |

New croquis and rimas are also published as an [RSS feed](https://cptx.ee/index.xml), and the site can be searched at [/search](https://cptx.ee/search/).

## How it works

A small Go server renders markdown into HTML pages. No JavaScript framework, no database and no trackers; the few scripts are small, plain JavaScript, and every page works without them. The site counts its own readers (daily totals only, no cookies): see [readers](https://cptx.ee/addons/readers/).

- **Stack:** Go, [htmlc](https://github.com/dhamidi/htmlc) (server-side `.vue` templates), [dispatch](https://github.com/dhamidi/dispatch) (routing), [goldmark](https://github.com/yuin/goldmark) (markdown) and [gofeed](https://github.com/mmcdole/gofeed) (reads the blogs I follow for [reads](https://cptx.ee/addons/reads/)).
- **Writing** happens in an [Obsidian](https://obsidian.md/) vault; published pieces are copied here as markdown under `content/`.
- **Hosting:** an [exe.dev](https://exe.dev) VM checks this repository every minute. When `main` moves, it builds the new version, swaps it in only if it starts and answers its health check, and keeps the previous one otherwise.
- **Checks:** every push runs [`scripts/check.sh`](scripts/check.sh) (build, tests, content lint, a snapshot of all ~3,000 URLs so none silently disappear, and a request to every page). On `main`, CI then waits for the live site to report the new commit and smoke-tests it.
- **Photos and book covers** live only on the server, not in git.
- **Newsletter:** issues are markdown files too (`content/newsletter/NNN.md`, Spanish `NNN.es.md`), drafted by [`scripts/newsletter-draft.py`](scripts/newsletter-draft.py). Once an issue is live, CI sends it through [Buttondown](https://buttondown.com) as one bilingual email ([`scripts/newsletter-send.py`](scripts/newsletter-send.py)); Buttondown handles sign-ups, confirmation and unsubscribes. Open and click tracking are off.

```
content/*.md ──► GitHub (main) ──► check ──► VM pulls, builds, swaps ──► https://cptx.ee
                                                                  └──► new issue? ──► Buttondown ──► inboxes
```

## Repository map

| Path | What lives there |
|---|---|
| `content/` | Posts (`posts/croquis` holds croquis and rimas, told apart by `categories`; `posts/recuerdos`), pockets (bedsidetable, cheatsheet), manifesto, legal docs, newsletter issues |
| `pages/`, `components/` | Page templates (`.vue`, rendered on the server) |
| `static/` | CSS, small JS files, fonts, icons |
| `data/library/` | The library: Libib export, shelves, overrides, generated `books.json`; the vinyl list `vinyl.csv` and generated `vinyl.json`; the film list `films.csv` and generated `films.json` |
| `data/reads.tsv` | Articles I kept (title, author, link), shown on `/addons/reads/` |
| `data/workbench/` | One markdown file per project: hardware, software, guide, notes |
| `main.go` | Startup and routes |
| `content.go` | Markdown loading, frontmatter, post IDs and URLs |
| `site_extras.go` | Headers, compression, RSS, sitemap, `/healthz`, redirects |
| `build.go` | Static build into `dist/` |
| `addons.go`, `places.go`, `words.go` | The addons |
| `library.go` | Archives and library pages |
| `vinyl.go` | The vinyl crate (`/archives/vinyl/`): records, square covers, listen links |
| `films.go` | The film shelf (`/archives/films/`): films, posters, where-to-watch links |
| `stats.go`, `geo.go` | Readers' counters: people's visits, countries and readers vs machines, without cookies or timers (`/addons/readers/`) |
| `newsletter.go` | Newsletter page, issues (`/newsletter/<n>/`, Spanish `/newsletter/<n>/es/`), kiosk covers |
| `search.go` | Site-wide search (`/search/`) |
| `legal.go` | `/legaldocs/` (privacy — including the newsletter —, cookies, how the site works) |
| `feeds.go`, `reads.go` | The blogs' feeds and `/addons/reads/` (kept + latest) |
| `workbench.go` | `/pockets/workbench/` and one guide page per project |
| `photos.go`, `pockets_photos.go` | Photos, pockets |
| `scripts/` | `check.sh`, URL snapshot, smoke test, link check, library import, newsletter draft/send, Lighthouse |
| `*.service`, `*.timer`, `deploy.sh` | How the VM runs and deploys the site |

## Posts and URLs

A post is a markdown file with a small frontmatter block:

```yaml
---
title: "Título"
author: cptx
date: 2026-11-15
draft: false
categories: ["croquis"]   # or ["rimas"] for a poem
showTags: true
---
```

Croquis, rimas and recuerdos get URLs like `/croquis/202611151/`: the date plus the post's position among same-day posts, ordered by filename. Adding `id: 202611151` pins that number, so a post keeps its URL if its date or filename changes. Old `/posts/<section>/<slug>/` addresses redirect to the new ones, and the poems' former `/croquis/<id>/` addresses redirect to `/rimas/<id>/`.

## Run it locally

You need [mise](https://mise.jdx.dev/) (it installs the pinned Go version) or Go 1.26+.

```sh
git clone https://github.com/cptx13/cptx.ee && cd cptx.ee
mise install
FEEDS_AUTO_REFRESH=0 mise exec -- go run .     # http://localhost:8080
```

`PORT` changes the port. Photos and book covers aren't in the repository, so locally the photo pages are empty and books show their generated stand-in covers. Before pushing a change, run `scripts/check.sh`.

Deployment details are in [DEPLOYMENT.md](DEPLOYMENT.md).

## Security

Please report security problems privately; see [SECURITY.md](SECURITY.md).

## Credits

The move from Hugo to Go and htmlc was started by [Dario Hamidi](https://github.com/dhamidi), who also wrote htmlc and dispatch. Book data comes from [Open Library](https://openlibrary.org) and Google Books; record data from [MusicBrainz](https://musicbrainz.org), [Wikidata](https://www.wikidata.org) and the [Cover Art Archive](https://coverartarchive.org).

All writing and photographs © cptx. Please ask before reusing them.
