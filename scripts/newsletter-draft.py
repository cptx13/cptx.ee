#!/usr/bin/env python3
"""Draft the next weekly newsletter issue (batch K; issues are "issue 01", "issue 02"…).

    scripts/newsletter-draft.py [--since YYYY-MM-DD] [--dry-run]

Writes content/newsletter/NNN.md and its Spanish version NNN.es.md, both with
draft: true (hidden until you publish them):
- "what was written": croquis dated in the period, with their live URLs;
- "what changed": the site's own changes from the commit history (content commits left out);
- "what is being written" and "other stuff": left for you to write.
The period starts the day after the last issue (or 7 days ago) and ends today.
Nothing is sent; sending happens after you publish the issue.
"""
import argparse, datetime as dt, os, re, subprocess, sys

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
NEWS = os.path.join(ROOT, "content", "newsletter")
CROQUIS = os.path.join(ROOT, "content", "posts", "croquis")


def frontmatter(path):
    t = open(path, encoding="utf-8").read().lstrip()
    if not t.startswith("---"):
        return {}
    fm = {}
    for line in t[3:t.find("---", 3)].splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            fm[k.strip()] = v.strip().strip("\"'")
    return fm


def day(s):
    return dt.date.fromisoformat(s[:10])


def croquis_urls():
    """Same rule as content.go: /croquis/YYYYMMDD<n>/, n = position among same-day files by name."""
    groups = {}
    for f in sorted(os.listdir(CROQUIS)):
        if not f.endswith(".md"):
            continue
        fm = frontmatter(os.path.join(CROQUIS, f))
        if fm.get("draft", "").lower() in ("true", "yes") or "date" not in fm:
            continue
        cats = re.findall(r'[\w-]+', fm.get("categories", "")) or ["croquis"]
        d = day(fm["date"])
        if fm.get("id"):
            yield d, fm.get("title", f[:-3]), f"/{cats[0]}/{fm['id']}/"
            continue
        groups.setdefault((cats[0], d), []).append((f, fm.get("title", f[:-3])))
    for (cat, d), items in groups.items():
        for i, (f, title) in enumerate(sorted(items), 1):
            yield d, title, f"/{cat}/{d:%Y%m%d}{i}/"


def site_changes(since):
    out = subprocess.run(["git", "-C", ROOT, "log", f"--since={since.isoformat()}", "--format=%s", "origin/main"],
                         capture_output=True, text=True).stdout.splitlines()
    skip = re.compile(r"^(croquis|edit|unpublish|pockets|newsletter|recuerdos)\b|^Merge ", re.I)
    return [s for s in out if s and not skip.match(s)]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--since")
    ap.add_argument("--dry-run", action="store_true")
    a = ap.parse_args()
    os.makedirs(NEWS, exist_ok=True)
    issues = sorted(int(f[:-3]) for f in os.listdir(NEWS) if re.fullmatch(r"\d+\.md", f))
    n = (issues[-1] + 1) if issues else 1
    today = dt.date.today()
    if a.since:
        since = dt.date.fromisoformat(a.since)
    elif issues:
        since = day(frontmatter(os.path.join(NEWS, f"{issues[-1]:03d}.md"))["date"]) + dt.timedelta(days=1)
    else:
        since = today - dt.timedelta(days=7)

    written = sorted((d, t, u) for d, t, u in croquis_urls() if since <= d <= today)
    changes = site_changes(since)
    lines = ["---", 'title: "This week on cptx.ee"', "author: cptx", f"date: {today.isoformat()}", "draft: true",
             f'summary: "{since:%-d %b} – {today:%-d %b %Y}"', "---", "",
             "<!-- Give the issue a title above (it becomes the cover headline; the cover already says 'issue NN'), edit freely,",
             "     then set draft: false and push. Nothing is sent until the issue is published. -->", "",
             "## what changed", ""]
    lines += [f"- {c}" for c in changes] or ["- (nothing new on the site this week)"]
    lines += ["", "## what was written", ""]
    lines += [f"- [{t}](https://cptx.ee{u}) · {d:%-d %b}" for d, t, u in written] or ["- (no new croquis this week)"]
    lines += ["", "## what is being written", "", "<!-- drafts in progress, ideas, what might come next -->", "",
              "## other stuff", "", "<!-- weekend reads, music, things worth passing on -->", ""]
    text = "\n".join(lines)
    # the Spanish version: same lists, Spanish headings, to be translated
    es = ["---", 'title: "Esta semana en cptx.ee"', "author: cptx", f"date: {today.isoformat()}", "draft: true", "lang: es",
          f'summary: "{since:%-d %b} – {today:%-d %b %Y}"', "---", "",
          "<!-- Versión en español de este número (/newsletter/N/es/). Traduce y publica junto con la versión en inglés. -->", "",
          "## qué ha cambiado", ""]
    es += [f"- {c}" for c in changes] or ["- (nada nuevo en la web esta semana)"]
    es += ["", "## qué se ha escrito", ""]
    es += [f"- [{t}](https://cptx.ee{u}) · {d:%-d %b}" for d, t, u in written] or ["- (ningún croquis nuevo esta semana)"]
    es += ["", "## qué se está escribiendo", "", "## otras cosas", ""]
    files = {os.path.join(NEWS, f"{n:03d}.md"): text, os.path.join(NEWS, f"{n:03d}.es.md"): "\n".join(es)}
    if a.dry_run:
        for path, t in files.items():
            print(f"--- {path}\n{t}")
        return
    for path in files:
        if os.path.exists(path):
            sys.exit(f"{path} already exists")
    for path, t in files.items():
        open(path, "w", encoding="utf-8").write(t)
    print(f"drafted issue {n:02d} (en + es): {len(written)} croquis, {len(changes)} site changes since {since}")


if __name__ == "__main__":
    main()
