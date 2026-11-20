#!/usr/bin/env python3
"""Draft the next weekly issue of the newsletter ("issue 01", "issue 02"…).

    scripts/newsletter-draft.py [--since YYYY-MM-DD] [--until YYYY-MM-DD] [--dry-run]

Writes content/newsletter/NNN.md and its Spanish version NNN.es.md, both with
draft: true (hidden until you publish them). Sections, in order:

- what changed            the site's own changes (commit subjects, content commits
                          left out): raw material to rewrite for readers
- what was written        croquis and rimas dated in the week, with their URLs
- this week, years ago    croquis and recuerdos written on the same days in
                          earlier years, newest year first, one post per line,
                          opening with a few lines of the longest one
- what is being written   yours
- articles I enjoyed      articles kept in data/reads.tsv during the week (not
                          in an earlier issue); if none, up to 5 posts of the week from the
                          blogs I follow (feed-cache.json, on the VM), else yours
- reading                 new entries in the bedsidetable pocket, else yours
- listening, watching     yours

A section left empty stays in the file as an HTML comment, so it shows nowhere
(not on the site, not in the email) until you fill it in.
The week starts the day after the last issue (or 6 days before --until) and
ends on --until (default today). Nothing is sent: sending happens after the
issue is published (see scripts/newsletter-send.py).
"""
import argparse, datetime as dt, os, re, subprocess, sys

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
NEWS = os.path.join(ROOT, "content", "newsletter")
POSTS = os.path.join(ROOT, "content", "posts")
POCKETS = os.path.join(ROOT, "content", "pockets")
READS = os.path.join(ROOT, "data", "reads.tsv")
FEED_CACHE = os.environ.get("FEED_CACHE", os.path.join(ROOT, "feed-cache.json"))
SITE = "https://cptx.ee"

HEAD = {
    "en": {"changed": "what changed", "written": "what was written", "past": "this week, years ago",
           "writing": "what is being written", "articles": "articles I enjoyed", "reading": "reading",
           "listening": "listening", "watching": "watching", "title": "This week on cptx.ee",
           "none_written": "Nothing new written this week.", "more": "and {n} more",
           "fresh": "Fresh from the blogs I follow:"},
    "es": {"changed": "qué ha cambiado", "written": "qué se ha escrito", "past": "esta semana, otros años",
           "writing": "qué se está escribiendo", "articles": "artículos que me gustaron", "reading": "leyendo",
           "listening": "escuchando", "watching": "viendo", "title": "Esta semana en cptx.ee",
           "none_written": "Nada nuevo escrito esta semana.", "more": "y {n} más",
           "fresh": "Lo último de los blogs que sigo:"},
}


def read(path):
    return open(path, encoding="utf-8").read()


def frontmatter(text):
    t = text.lstrip()
    if not t.startswith("---"):
        return {}, text
    end = t.find("\n---", 3)
    fm = {}
    for line in t[3:end].splitlines():
        if ":" in line:
            k, v = line.split(":", 1)
            fm[k.strip()] = v.strip().strip("\"'")
    return fm, t[end + 4:]


def day(s):
    return dt.date.fromisoformat(s[:10])


def posts():
    """(date, title, url, body) for every published croquis and recuerdo.
    URL rule as in content.go: /<category>/YYYYMMDD<n>/, n = position among
    same-category, same-day files sorted by name; frontmatter id: pins it."""
    groups, out = {}, []
    for section in ("croquis", "recuerdos"):
        d = os.path.join(POSTS, section)
        for f in sorted(os.listdir(d)):
            if not f.endswith(".md"):
                continue
            fm, body = frontmatter(read(os.path.join(d, f)))
            if fm.get("draft", "").lower() in ("true", "yes") or "date" not in fm:
                continue
            cat = (re.findall(r"[\w-]+", fm.get("categories", "")) or [section])[0]
            item = [day(fm["date"]), fm.get("title", f[:-3]), None, body, f]
            if fm.get("id"):
                item[2] = f"/{cat}/{fm['id']}/"
            else:
                groups.setdefault((cat, item[0]), []).append(item)
            out.append(item)
    for (cat, d), items in groups.items():
        for i, item in enumerate(sorted(items, key=lambda x: x[4]), 1):
            item[2] = f"/{cat}/{d:%Y%m%d}{i}/"
    return [tuple(i[:4]) for i in out]


def opening(body, limit=260):
    """First sentences of a post's text, up to about `limit` characters."""
    text = " ".join(l.strip() for l in body.splitlines() if l.strip() and not l.lstrip().startswith(("#", "!", "<", ">", "|", "-")))
    text = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", text)
    text = re.sub(r"[*_`]", "", text)
    out = ""
    for s in re.split(r"(?<=[.!?…])\s+", text):
        if out and len(out) + len(s) > limit:
            break
        out = (out + " " + s).strip()
    return out


def site_changes(since, until):
    out = subprocess.run(["git", "-C", ROOT, "log", f"--since={since.isoformat()}", f"--until={(until + dt.timedelta(days=1)).isoformat()}",
                          "--format=%s", "origin/main"], capture_output=True, text=True).stdout.splitlines()
    skip = re.compile(r"^(croquis|edit|unpublish|pockets|newsletter|recuerdos|urls\.txt)\b|^Merge ", re.I)
    return [s for s in out if s and not skip.match(s)]


def pocket_entries(slug):
    """List items of a pocket, newest month first."""
    path = os.path.join(POCKETS, slug + ".md")
    if not os.path.exists(path):
        return []
    return [l.strip()[2:] for l in read(path).splitlines() if l.strip().startswith(("* ", "- "))]


def in_earlier_issues():
    seen = ""
    for f in os.listdir(NEWS):
        if re.fullmatch(r"\d+(\.es)?\.md", f):
            seen += read(os.path.join(NEWS, f))
    return seen


def new_pocket_entries(slug, seen):
    out = []
    for e in pocket_entries(slug):
        m = re.search(r"\]\(([^)]+)\)", e)
        key = m.group(1) if m else e
        if key not in seen:
            out.append(e)
    return out


def kept_reads(seen, since, until):
    """Articles kept in data/reads.tsv during the week whose link is in no earlier issue."""
    out = []
    if not os.path.exists(READS):
        return out
    for line in read(READS).splitlines():
        if not line.strip() or line.startswith("#"):
            continue
        f = (line.split("\t") + [""] * 6)[:6]
        kept, title, author, url = f[0].strip(), f[1].strip(), f[2].strip(), f[3].strip()
        kept = kept if len(kept) == 10 else kept + "-01"
        if url and url not in seen and since.isoformat() <= kept <= until.isoformat():
            out.append(f"[{title}]({url})" + (f" by {author}" if author else ""))
    return out


def fresh_from_blogs(since, until, limit=5):
    """Fallback: posts of the week from the blogs I follow (AD-11)."""
    import json
    try:
        cache = json.loads(read(FEED_CACHE))
    except (OSError, ValueError):
        return []
    out = []
    for it in cache.get("items", []):
        d = (it.get("date") or "")[:10]
        if d and since.isoformat() <= d <= until.isoformat():
            out.append(f"[{it['title']}]({it['link']}) by {it['source']}")
        if len(out) == limit:
            break
    return out


def section(title, lines, placeholder=""):
    """A section; empty ones are kept as a comment so they show nowhere."""
    if lines:
        return [f"## {title}", ""] + lines + [""]
    return [f"<!-- ## {title}", "", placeholder or "(fill in, or leave this comment: it shows nowhere)", "-->", ""]


MONTHS = {"en": "Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec".split(),
          "es": "ene feb mar abr may jun jul ago sep oct nov dic".split()}


def short_title(t):
    return t.strip().rstrip(".")


def past_years(all_posts, since, until, lang):
    """What was written on the same days in earlier years: newest year first,
    one post per line with its date; opens with a few lines from the longest."""
    span = [since + dt.timedelta(days=i) for i in range((until - since).days + 1)]
    by_day = {d.strftime("%m-%d"): [] for d in span}
    for d, t, u, body in all_posts:
        k = d.strftime("%m-%d")
        if d.year < until.year and k in by_day:
            by_day[k].append((d, t, u, body))
    found = [p for ps in by_day.values() for p in ps]
    if not found:
        return []
    d, t, u, body = max(found, key=lambda p: len(p[3]))
    lines = [f"> {opening(body)}", ">",
             f"> — *[{short_title(t)}]({SITE}{u})*, {d.day} {MONTHS[lang][d.month - 1]} {d.year}", ""]
    years = sorted({p[0].year for p in found}, reverse=True)
    for y in years:
        lines += [f"**{y}**", ""]
        for p in sorted((p for p in found if p[0].year == y), key=lambda p: p[0]):
            lines.append(f"- {p[0].day} {MONTHS[lang][p[0].month - 1]} · [{short_title(p[1])}]({SITE}{p[2]})")
        lines.append("")
    return lines


def build(lang, n, since, until, all_posts, changes, seen):
    h = HEAD[lang]
    written = sorted((d, t, u) for d, t, u, _ in all_posts if since <= d <= until and u.startswith(("/croquis/", "/rimas/")))
    out = ["---", f'title: "{h["title"]}"', "author: cptx", f"date: {until.isoformat()}", "draft: true"]
    if lang == "es":
        out.append("lang: es")
    out += [f'summary: "{since:%-d %b} – {until:%-d %b %Y}"', "---", "",
            "<!-- Give the issue a title above (the cover headline; the cover already says 'issue NN'),",
            "     rewrite 'what changed' for readers, fill in what you like, then set draft: false",
            "     in both files and push. Nothing is sent until the issue is published. -->", ""]
    out += section(h["changed"], [f"- {c}" for c in changes])
    out += section(h["written"], [f"- [{t}]({SITE}{u}) · {d:%-d %b}" for d, t, u in written] or [h["none_written"]])
    out += section(h["past"], past_years(all_posts, since, until, lang))
    out += section(h["writing"], [])
    articles = kept_reads(seen, since, until)
    if not articles:
        fresh = fresh_from_blogs(since, until)
        if fresh:
            articles = [h["fresh"], ""] + fresh
    out += section(h["articles"], [e if e in ("", h["fresh"]) else f"- {e}" for e in articles])
    out += section(h["reading"], [f"- {e}" for e in new_pocket_entries("bedsidetable", seen)])
    out += section(h["listening"], [])
    out += section(h["watching"], [])
    return "\n".join(out)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--since")
    ap.add_argument("--until")
    ap.add_argument("--dry-run", action="store_true")
    a = ap.parse_args()
    os.makedirs(NEWS, exist_ok=True)
    issues = sorted(int(f[:-3]) for f in os.listdir(NEWS) if re.fullmatch(r"\d+\.md", f))
    n = (issues[-1] + 1) if issues else 1
    until = dt.date.fromisoformat(a.until) if a.until else dt.date.today()
    if a.since:
        since = dt.date.fromisoformat(a.since)
    elif issues:
        since = day(frontmatter(read(os.path.join(NEWS, f"{issues[-1]:03d}.md")))[0]["date"]) + dt.timedelta(days=1)
    else:
        since = until - dt.timedelta(days=6)
    all_posts, changes, seen = posts(), site_changes(since, until), in_earlier_issues()
    files = {os.path.join(NEWS, f"{n:03d}.md"): build("en", n, since, until, all_posts, changes, seen),
             os.path.join(NEWS, f"{n:03d}.es.md"): build("es", n, since, until, all_posts, changes, seen)}
    if a.dry_run:
        for path, t in files.items():
            print(f"--- {os.path.relpath(path, ROOT)}\n{t}")
        return
    for path in files:
        if os.path.exists(path):
            sys.exit(f"{path} already exists")
    for path, t in files.items():
        open(path, "w", encoding="utf-8").write(t)
    print(f"drafted issue {n:02d} (en + es) for {since} – {until}")


if __name__ == "__main__":
    main()
