---
title: "How this site works"
author: cptx
date: 2026-11-15
draft: false
lang: en
summary: "Where the writing comes from, what runs the site, and what is (and isn't) loaded in your browser."
---
## From notebook to page

Everything starts as a note in [Obsidian](https://obsidian.md/), on my own computer. When a piece is ready, it is copied as a plain markdown file into the site's public repository on [GitHub](https://github.com/cptx13/cptx.ee). Every change is checked automatically (does it build, do all the old addresses still work), and a small server picks up the new version within a couple of minutes.

## What runs it

- A small program written in [Go](https://go.dev/) turns markdown into pages. It runs on a virtual server rented from [exe.dev](https://exe.dev).
- Pages are plain HTML and CSS. The few scripts are tiny and only add conveniences (switching the theme, searching the list, sharing a view, browsing photos); every page works without them.
- The font is [JetBrains Mono](https://www.jetbrains.com/lp/mono/), served from this site.
- Photos and book covers are stored on the server, not in the repository.

## What your browser loads

Only files from cptx.ee: the page, one stylesheet, one font in two or four styles, a few small scripts and the images on that page. No ads, no embedded videos or social buttons, no fonts or scripts from other companies, and no analytics service: the [readers](/addons/readers/) counts are made by the site itself, without cookies (see the [privacy policy](/legaldocs/privacy/#counting-readers)).

## Where some content comes from

- **The library** uses book details and covers from [Open Library](https://openlibrary.org) and Google Books. They were fetched once by the server; your browser doesn't contact them.
- **The cheatsheet's reading list** is collected by the server from the blogs' public RSS feeds a few times a day.
- **City lights** finds place names in the posts themselves.
- **Readers' countries** come from [DB-IP](https://db-ip.com)'s free country list (CC BY 4.0), a file the server downloads once a month and checks locally.

## Feeds and archives

New croquis are published as an [RSS feed](/index.xml), and there is a [sitemap](/sitemap.xml) for search engines. The full source code, including every change to these pages, is public on [GitHub](https://github.com/cptx13/cptx.ee).

## Security

Problems can be reported privately: see the [security policy](https://github.com/cptx13/cptx.ee/security/policy).
