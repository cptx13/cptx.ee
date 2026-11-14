#!/usr/bin/env bash
# Live smoke test (T-04):  scripts/smoke.sh [BASE_URL] [EXPECTED_COMMIT]
# Checks /healthz (and that it runs EXPECTED_COMMIT, if given), feed/sitemap/robots,
# a legacy accented redirect, the 404 page and a random sample of 60 URLs from urls.txt.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
BASE="${1:-https://cptx.ee}"
EXPECT="${2:-}"
python3 - "$BASE" "$EXPECT" <<'PY'
import json, random, sys, urllib.parse, urllib.request, urllib.error
base, expect = sys.argv[1].rstrip("/"), sys.argv[2][:7]
class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *a, **k): return None
opener = urllib.request.build_opener(NoRedirect)
def get(path):
    req = urllib.request.Request(base + urllib.parse.quote(path, safe="/"), headers={"User-Agent": "cptx-smoke"})
    try:
        r = opener.open(req, timeout=20); return r.status, r.headers, r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.headers, b""
fails = []
s, _, body = get("/healthz")
h = json.loads(body or b"{}") if s == 200 else {}
print(f"healthz {s} {h}")
if s != 200 or h.get("status") != "ok": fails.append("healthz")
if expect and not h.get("commit", "").startswith(expect): fails.append(f"commit {h.get('commit')} != {expect}")
for path, want in [("/index.xml", 200), ("/sitemap.xml", 200), ("/robots.txt", 200),
                   ("/posts/recuerdos/761---césped-húmedo/", 301), ("/no-such-page/", 404)]:
    s, _, _ = get(path)
    if s != want: fails.append(f"{path} {s} (want {want})")
urls = [l.strip() for l in open("urls.txt", encoding="utf-8") if l.strip()]
for u in random.sample(urls, min(60, len(urls))):
    s, hd, _ = get(u)
    if s not in (200, 301): fails.append(f"{u} {s}")
print(f"{len(fails)} failure(s)"); [print("  FAIL:", f) for f in fails]
sys.exit(1 if fails else 0)
PY
