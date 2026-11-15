#!/usr/bin/env python3
"""Hand published issues to Buttondown, once each.

    scripts/newsletter-send.py [--dry-run]

Runs in CI after a verified deploy of main. For every published issue
(content/newsletter/NNN.md, draft: false) dated in the last 21 days it creates
one Buttondown email for everyone:

- subject  "cptx · issue 01: <title> / <título>"
- body     the English issue, then the Spanish one (NNN.es.md, when published)
           under a divider; comments removed, links made absolute, each half
           opening with a link to its page on the web
- slug and metadata "cptx-issue-01", so the same issue is never created twice

(One bilingual email because Buttondown's tags, which could split readers by
language, need a paid plan.)

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

WEB = {"en": "read it on the web", "es": "léelo en la web"}
KIOSK = "Every issue is kept on the [kiosk]({site}/newsletter/kiosk/) · Todos los números se guardan en el [quiosco]({site}/newsletter/kiosk/)."


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


def published(path):
    if not os.path.exists(path):
        return None
    fm, body = frontmatter(open(path, encoding="utf-8").read())
    if fm.get("draft", "").lower() in ("true", "yes"):
        return None
    return fm, body


def issues(today):
    """(number, english (fm, body), spanish (fm, body) or None) for recent published issues."""
    out = []
    for f in sorted(os.listdir(NEWS)):
        m = re.fullmatch(r"(\d+)\.md", f)
        if not m:
            continue
        en = published(os.path.join(NEWS, f))
        if not en:
            continue
        date = dt.date.fromisoformat(en[0].get("date", "1970-01-01")[:10])
        if (today - date).days > RECENT_DAYS or date > today:
            continue
        n = int(m.group(1))
        out.append((n, en, published(os.path.join(NEWS, f"{n:03d}.es.md"))))
    return out


def clean(body):
    body = re.sub(r"<!--.*?-->", "", body, flags=re.S)
    body = re.sub(r"\]\(/", f"]({SITE}/", body)
    return re.sub(r"\n{3,}", "\n\n", body).strip()


def email_body(n, en, es):
    date = dt.date.fromisoformat(en[0]["date"][:10])
    parts = [f"*issue {n:02d} · {date:%-d %b %Y} · [{WEB['en']}]({SITE}/newsletter/{n}/)*", clean(en[1])]
    if es:
        parts += ["---", f"# {es[0].get('title', '').strip()}",
                  f"*issue {n:02d} · en español · [{WEB['es']}]({SITE}/newsletter/{n}/es/)*", clean(es[1])]
    parts += ["---", KIOSK.format(site=SITE)]
    return "\n\n".join(parts) + "\n"


def payload(n, en, es, mode):
    key = f"cptx-issue-{n:02d}"
    subject = f"cptx · issue {n:02d}: {en[0].get('title', '').strip()}"
    if es:
        subject += f" / {es[0].get('title', '').strip()}"
    return key, {
        "subject": subject,
        "body": email_body(n, en, es),
        "status": "about_to_send" if mode == "send" else "draft",
        "slug": key,
        "metadata": {"cptx_issue": key},
        "canonical_url": f"{SITE}/newsletter/{n}/",
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
