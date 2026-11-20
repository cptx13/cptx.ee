---
title: cptx.ee
icon: globe
status: running
order: 6
summary: This site: a small Go server, written in Obsidian, published from a chat.
link: https://cptx.ee/
hardware: an exe.dev VM
software: Go, htmlc, goldmark, GitHub, Obsidian, Claude
---
## guide

1. **Write in Obsidian**, on the computer or the phone.
2. **Claude turns a note into a post**, shows a preview and waits for my yes.
3. **The post goes to GitHub**; checks run on every change.
4. **A small VM polls GitHub** every minute, rebuilds the Go server and swaps it in only if it starts.

## notes

- The long version: [how this site works](/legaldocs/how-it-works/). Every line of code is on [GitHub](https://github.com/cptx13/cptx.ee).
- The books in the [library](/archives/library/) are scanned with [Libib](https://www.libib.com/).
