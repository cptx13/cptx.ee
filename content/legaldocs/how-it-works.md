---
title: "How this site works"
author: cptx
date: 2026-11-22
draft: false
lang: en
summary: "Where the writing comes from, what runs the site, and what is (and isn't) loaded in your browser."
---
## From notebook to page

Everything starts as a note in [Obsidian](https://obsidian.md/), on my own computer. When a piece is ready, it is copied as a plain markdown file into the site's public repository on [GitHub](https://github.com/cptx13/cptx.ee). Every change is checked automatically (does it build, do all the old addresses still work), and a small server picks up the new version within a couple of minutes.

## What runs it

- A small program written in [Go](https://go.dev/) turns markdown into pages. It runs on a virtual server rented from [exe.dev](https://exe.dev).
- Pages are plain HTML and CSS. The few scripts are tiny and only add conveniences (switching the theme, sharing a view, browsing photos, telling the [readers](/addons/readers/) counter that a person is on the page); every page works without them.
- The font is [JetBrains Mono](https://www.jetbrains.com/lp/mono/), served from this site.
- Photos, book covers and record covers are stored on the server, not in the repository.

## What your browser loads

Only files from cptx.ee: the page, one stylesheet, one font in two or four styles, a few small scripts and the images on that page. No ads, no embedded videos or social buttons, no fonts or scripts from other companies, and no analytics service (the newsletter form is the one thing that talks to another company, Buttondown, and only when you press subscribe): the [readers](/addons/readers/) counts are made by the site itself, without cookies (see the [privacy policy](/legaldocs/privacy/#counting-readers)).

## Where some content comes from

- **The library** uses book details and covers from [Open Library](https://openlibrary.org) and Google Books. They were fetched once by the server; your browser doesn't contact them.
- **The vinyl crate** uses record details from [MusicBrainz](https://musicbrainz.org) and [Wikidata](https://www.wikidata.org), and covers from the [Cover Art Archive](https://coverartarchive.org) or photos of our own copies, all fetched once by the server. The Spotify, YouTube Music and Wikipedia links are plain links: nothing loads from those sites until you click one.
- **The latest posts on [reads](/addons/reads/?view=latest)** are collected by the server from the cheatsheet blogs' public RSS feeds a few times a day; your browser never contacts those blogs until you click a link.
- **City lights** finds place names in the posts themselves.
- **Readers' countries** come from [DB-IP](https://db-ip.com)'s free country list (CC BY 4.0), a file the server downloads once a month and checks locally.

## Feeds and archives

New croquis are published as an [RSS feed](/index.xml), and there is a [sitemap](/sitemap.xml) for search engines. The weekly issues of the [newsletter](/newsletter/) are pages on this site too; [Buttondown](https://buttondown.com) only delivers them by email (see [the newsletter](/legaldocs/privacy/#newsletter) in the privacy policy). The full source code, including every change to these pages, is public on [GitHub](https://github.com/cptx13/cptx.ee).

## Security

Problems can be reported privately: see the [security policy](https://github.com/cptx13/cptx.ee/security/policy).
