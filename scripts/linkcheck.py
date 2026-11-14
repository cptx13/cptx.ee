#!/usr/bin/env python3
"""Monthly external link check (T-05): python3 scripts/linkcheck.py
Finds http(s) links in content/ and pages/, checks each once, and fails on links
that are gone (404/410/DNS errors). 401/403/429 and timeouts are reported as
warnings only, since many sites block bots."""
import concurrent.futures as cf, pathlib, re, sys, urllib.error, urllib.request

LINK = re.compile(r"https?://[^\s)\]\"'<>`]+")
links = {}
for p in list(pathlib.Path("content").rglob("*.md")) + list(pathlib.Path("pages").glob("*.vue")) + list(pathlib.Path("components").glob("*.vue")):
    for m in LINK.findall(p.read_text(encoding="utf-8", errors="ignore")):
        u = m.rstrip(".,;:")
        if "cptx.ee" not in u:
            links.setdefault(u, set()).add(str(p))

def check(u):
    for method in ("HEAD", "GET"):
        req = urllib.request.Request(u, method=method, headers={"User-Agent": "Mozilla/5.0 (cptx-linkcheck)"})
        try:
            with urllib.request.urlopen(req, timeout=20) as r:
                return u, r.status
        except urllib.error.HTTPError as e:
            if method == "HEAD" and e.code in (403, 405, 400, 501):
                continue
            return u, e.code
        except Exception as e:
            if method == "HEAD":
                continue
            return u, type(e).__name__
    return u, "?"

broken, warn = [], []
with cf.ThreadPoolExecutor(16) as ex:
    for u, s in ex.map(check, sorted(links)):
        if s in (404, 410) or s in ("URLError", "gaierror"):
            broken.append((u, s))
        elif not (isinstance(s, int) and s < 400):
            warn.append((u, s))
print(f"{len(links)} external links, {len(broken)} broken, {len(warn)} warnings")
for u, s in broken:
    print(f"  BROKEN {s} {u}  <- {', '.join(sorted(links[u]))[:150]}")
for u, s in warn:
    print(f"  warn   {s} {u}")
sys.exit(1 if broken else 0)
