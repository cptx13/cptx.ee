---
title: "manifesto"
author: cptx
date: 2026-11-14
draft: false
categories: ["manifesto"]
tags: [manifesto]
showTags: false
toc: false
---
Over the last decade, I have accumulated an obscene amount of digital notes. The official count hovers around **3,000**, reflecting only what I consider somewhat readable; the true number remains unknown to me, and yet, they somehow hold some value, they are true mathoms.

A mathom is a term coined by J.R.R. Tolkien to describe an object that Hobbits have no immediate use for but are unwilling to discard. Often considered a trinket or knick-knack, this idea has become integral to my recent efforts concerning my notes and writings.

This site emerged out of necessity and stubbornness. Instead of addressing my hoarding tendencies, I sought various organizational systems from different platforms, however, I became tired of fighting and losing data due to their monopolistic practices.

Since creating this site, I have developed various ways to organize my existing data. Thanks to [Obsidian](https://obsidian.md/) and its philosophy of prioritizing "file over app," I have regained my digital freedom and creative ownership.

---

### the principles

- **file over app**: everything starts as a plain markdown file that I own and can open with anything, today and in twenty years.
- **one source, many shapes**: the same files become pages, a feed, maps and heatmaps, without ever being rewritten for a platform.
- **quiet by design**: no ads, no trackers, no cookies, no accounts. Pages are rendered on the server and read fine without JavaScript.
- **nothing disappears**: old links keep working; removed pages redirect to where their content went.

---

### how it works

**Writing.** Everything is written in my [Obsidian](https://obsidian.md/) vault, *Mathom*. Drafts stay there; when a piece is ready, it is copied into the site's repository as markdown, keeping the vault-only notes behind.

**Code.** cptx.ee is a small [Go](https://go.dev/) application. [goldmark](https://github.com/yuin/goldmark) turns markdown into HTML, [htmlc](https://github.com/dhamidi/htmlc) renders the templates on the server and [dispatch](https://github.com/dhamidi/dispatch) handles the URLs. The source code and every published text live on [GitHub](https://github.com/cptx13/cptx.ee).

**Publishing.** The site runs on a small [exe.dev](https://exe.dev) virtual machine that checks GitHub every minute. When something new arrives, it builds the site, starts it, checks that it answers, and only then replaces the running version; if anything fails, the previous version stays up. Every change is also tested on GitHub before and after it goes live: the build, that no existing URL breaks, and the live pages.

**Extras.** On every start the server reads all the posts and builds the [addons](/addons/) (the archive map, the cities, the words) and the [library](/archives/library/): book data comes from my Libib scans, with covers and details from [Open Library](https://openlibrary.org/) and other public catalogues. Photos and book covers live only on the server. There is an [RSS feed](/index.xml) for the croquis and a sitemap for everything else.

---

### the sections

My works are organized and distributed using the following system:

- **[croquis](/categories/croquis/)** are quick, often rough notes made to capture the essence or form of a subject that may lead to more detailed work.
  - **[retales](/addons/map/?tag=retales)**: This refers to "re-tales," emphasizing the act of retelling and rewriting. In Spanish, "retales" refers to leftover scraps that can be repurposed, representing forgotten stories—fragments of thoughts that evoke nostalgia and inspire me to return to writing.
  - **[rimas](/addons/map/?tag=rimas)**: The Spanish term for "rhyme," often used in my earlier poems to distinguish the quality and immediacy of some works.

- **[recuerdos](/categories/recuerdos/)** are daily memories written from 2021 to 2024, a project that began as a naive attempt to document lived experiences but evolved into something far more demanding. What started as voluntary writing gradually transformed into a disciplined obligation, eventually stirring feelings of guilt with each passing day. Over 1,000 entries reflect the tension between the desire to preserve memory and the burden of maintaining the practice.

- **[photos](/photos/)** are collections of photographs from various places and years, served directly from the server.

- **[addons](/addons/)** are my playground, where I try new features that make the blog more immersive: an [archive map](/addons/map/) of every day I wrote, [what I wrote on this day](/addons/today/) in other years, the [cities that light up](/addons/places/) across my posts and the [words](/addons/words/) I keep coming back to.

- **[pockets](/pockets/)** comprise my daily resources, everything I would carry around if I could.
  - **[bedsidetable](/pockets/bedsidetable/)**: Some books I have enjoyed recently.
  - **[cheatsheet](/pockets/cheatsheet/)**: Some links I like to have at hand.
  - **[workbench](/pockets/workbench/)**: Some of my projects.
  - **[weekendreads](/pockets/weekendreads/)**: Articles from other people I have recently enjoyed.

- **[archives](/archives/)** are the things we keep: mathoms of a more physical kind.
  - **[written by me](/archives/written-by-me/)**: published books—manuscripts that have been formatted, reviewed, and are considered complete.
  - **[library](/archives/library/)**: some of the books Merzeede and I own, shown as covers or spines and shelved by genre, colour, author, decade, language or size. Not all of them have been scanned yet. Records, films and other collections will follow.
