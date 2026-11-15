#!/usr/bin/env python3
"""Hand published issues to Buttondown, once each, in each language.

    scripts/newsletter-send.py [--dry-run]

Runs in CI after a verified deploy of main. For every published issue
(content/newsletter/NNN.md, draft: false) dated in the last 21 days, and its
Spanish version NNN.es.md when there is one, it creates one Buttondown email:

- subject  "cptx · issue 01: <title>"
- body     the issue's markdown (comments removed, links made absolute), with a
           line on top linking to the issue on the web
- audience subscribers tagged "en" or "es" (the language picked at sign-up)
- slug and metadata "cptx-issue-01-en" so the same issue is never created twice

Environment:
  BUTTONDOWN_API_KEY   the API key (GitHub secret). Unset: nothing happens.
  NEWSLETTER_MODE      "draft" (default): emails wait in Buttondown as drafts for
                       you to send. "send": they go out straight away.
Only the Python standard library is used.
"""
import argparse, datetime as dt, json, os, re, sys, urllib.error, urllib.request

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
NEWS = os.path.join(ROOT, "content", "newsletter")
SITE = "https://cptx.ee"
API = "https://api.buttondown.com/v1"
RECENT_DAYS = 21

TEXT = {
    "en": {"web": "read it on the web", "kiosk": "Every issue is kept on the [kiosk]({site}/newsletter/kiosk/).",
           "other": "[Leer en español]({url})"},
    "es": {"web": "léelo en la web", "kiosk": "Todos los números se guardan en el [quiosco]({site}/newsletter/kiosk/).",
           "other": "[Read in English]({url})"},
}


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
    return fm, t[end + 4:].lstrip("\n")


def issues(today):
    """(number, lang, frontmatter, body, web path, other-language path or None)."""
    out = []
    for f in sorted(os.listdir(NEWS)):
        m = re.fullmatch(r"(\d+)(\.es)?\.md", f)
        if not m:
            continue
        fm, body = frontmatter(open(os.path.join(NEWS, f), encoding="utf-8").read())
        if fm.get("draft", "").lower() in ("true", "yes"):
            continue
        n, lang = int(m.group(1)), "es" if m.group(2) else "en"
        date = dt.date.fromisoformat(fm.get("date", "1970-01-01")[:10])
        if (today - date).days > RECENT_DAYS or date > today:
            continue
        path = f"/newsletter/{n}/" + ("es/" if lang == "es" else "")
        other_file = os.path.join(NEWS, f"{n:03d}.md" if lang == "es" else f"{n:03d}.es.md")
        other = None
        if os.path.exists(other_file):
            ofm, _ = frontmatter(open(other_file, encoding="utf-8").read())
            if ofm.get("draft", "").lower() not in ("true", "yes"):
                other = f"/newsletter/{n}/" + ("" if lang == "es" else "es/")
        out.append((n, lang, fm, body, path, other))
    return out


def email_body(n, lang, fm, body, path, other):
    t = TEXT[lang]
    body = re.sub(r"<!--.*?-->", "", body, flags=re.S)
    body = re.sub(r"\]\(/", f"]({SITE}/", body)
    body = re.sub(r"\n{3,}", "\n\n", body).strip()
    date = dt.date.fromisoformat(fm["date"][:10])
    top = f"*issue {n:02d} · {date:%-d %b %Y} · [{t['web']}]({SITE}{path})*"
    foot = [t["kiosk"].format(site=SITE)]
    if other:
        foot.append(t["other"].format(url=SITE + other))
    return f"{top}\n\n{body}\n\n---\n\n" + " · ".join(foot) + "\n"


def payload(n, lang, fm, body, path, other, mode):
    key = f"cptx-issue-{n:02d}-{lang}"
    return key, {
        "subject": f"cptx · issue {n:02d}: {fm.get('title', '').strip()}",
        "body": email_body(n, lang, fm, body, path, other),
        "status": "about_to_send" if mode == "send" else "draft",
        "slug": key,
        "metadata": {"cptx_issue": key},
        "canonical_url": SITE + path,
        "filters": {"predicate": "and", "groups": [],
                    "filters": [{"operator": "contains", "field": "subscriber.tags", "value": lang}]},
    }


def call(method, url, key, data=None, extra=None):
    headers = {"Authorization": f"Token {key}", "X-API-Version": "2026-04-01",
               "Content-Type": "application/json", "User-Agent": "cptx.ee newsletter-send"}
    headers.update(extra or {})
    req = urllib.request.Request(url, data=json.dumps(data).encode() if data is not None else None,
                                 headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=30) as r:
        return json.loads(r.read() or b"{}")


def existing_keys(key):
    """Slugs/metadata of every email already in Buttondown, any status."""
    seen = set()
    url = API + "/emails?" + "&".join(f"status={s}" for s in
                                       ("draft", "scheduled", "about_to_send", "in_flight", "sent", "imported", "throttled"))
    while url:
        page = call("GET", url, key)
        for e in page.get("results", []):
            seen.add(e.get("slug") or "")
            seen.add((e.get("metadata") or {}).get("cptx_issue", ""))
        url = page.get("next")
    return seen


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true", help="print what would be created, call nothing")
    ap.add_argument("--today", help="YYYY-MM-DD (testing)")
    a = ap.parse_args()
    today = dt.date.fromisoformat(a.today) if a.today else dt.date.today()
    mode = os.environ.get("NEWSLETTER_MODE", "draft")
    key = os.environ.get("BUTTONDOWN_API_KEY", "")
    todo = [payload(*i, mode) for i in issues(today)]
    if a.dry_run:
        for k, p in todo:
            print(f"=== {k} ({p['status']})\nSubject: {p['subject']}\n\n{p['body']}")
        return
    if not key:
        print("newsletter: BUTTONDOWN_API_KEY not set, nothing to do")
        return
    if not todo:
        print("newsletter: no recent published issues")
        return
    seen = existing_keys(key)
    for k, p in todo:
        if k in seen:
            print(f"newsletter: {k} already in Buttondown")
            continue
        extra = {"X-Buttondown-Live-Dangerously": "true"} if p["status"] == "about_to_send" else None
        try:
            e = call("POST", API + "/emails", key, p, extra)
        except urllib.error.HTTPError as err:
            print(f"newsletter: {k} failed: {err.code} {err.read()[:300]!r}", file=sys.stderr)
            sys.exit(1)
        print(f"newsletter: {k} created as {e.get('status', p['status'])}")


if __name__ == "__main__":
    main()
