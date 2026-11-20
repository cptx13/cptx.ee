---
title: cptx.ee
icon: globe
status: running
order: 3
summary: This site: a small Go server, written in Obsidian, published from a chat.
link: https://cptx.ee/
kit: Go, htmlc, goldmark, GitHub, an exe.dev VM, Obsidian, Claude
---
## how it's set up

1. I write in Obsidian, on the computer or the phone.
2. Claude turns a note into a post, shows me a preview and waits for my yes.
3. The post goes to GitHub; checks run on every change.
4. A small VM polls GitHub every minute, rebuilds the Go server and swaps it in only if it starts.

The long version is in [how this site works](/legaldocs/how-it-works/), and every line of code is on [GitHub](https://github.com/cptx13/cptx.ee).

## resources

- [Libib](https://www.libib.com/), where the books in the [library](/archives/library/) are scanned
