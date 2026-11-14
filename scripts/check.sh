#!/usr/bin/env bash
# Pre-push check for cptx.ee (T-01). Run from a checkout that is NOT the live one:
#   scripts/check.sh                 build + lint + URL diff + HTTP smoke
#   scripts/check.sh --update-urls   same, then rewrite urls.txt (after an intended URL change)
# On the VM use the pinned Go:  ~/.local/bin/mise exec -- scripts/check.sh
# Exit 0 = OK to push. Exit 1 = something would break the site or a URL.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
if [ "$PWD" = "$HOME/cptx-v2" ]; then
  echo "refusing to run in the live checkout ($PWD) — use a scratch clone"; exit 2
fi
UPDATE=0; [ "${1:-}" = "--update-urls" ] && UPDATE=1
PORT="${CHECK_PORT:-18080}"
OUT=.check; rm -rf "$OUT"; mkdir -p "$OUT"
fail=0
bad() { echo "  FAIL: $*"; fail=1; }

echo "== 1/4 build"
if ! go build -o "$OUT/cptx.ee" . 2> "$OUT/build.log"; then
  cat "$OUT/build.log"; echo "FAILED: build"; exit 1
fi
echo "  ok ($(go version | awk '{print $3}'))"

echo "== 2/4 content lint"
python3 - <<'PY' || fail=1
import pathlib, re, sys
VAULT = {"type", "blog", "status", "collection", "source", "draft_note"}
KNOWN = {"title", "date", "draft", "categories", "tags", "showTags", "author", "toc", "description"}
errors = warns = 0
for p in sorted(pathlib.Path("content").rglob("*.md")):
    t = p.read_text(encoding="utf-8")
    post = p.parts[1] == "posts"
    if not t.startswith("---"):
        if post: print(f"  WARN: no frontmatter: {p}"); warns += 1
        continue
    keys = set(re.findall(r"^([A-Za-z_]+):", t.split("---", 2)[1], re.M))
    if keys & VAULT:
        print(f"  FAIL: vault-only field(s) {sorted(keys & VAULT)}: {p}"); errors += 1
    if post and not {"title", "date"} <= keys:
        print(f"  WARN: missing title/date: {p}"); warns += 1
    for k in sorted(keys - KNOWN - VAULT):
        print(f"  WARN: unknown field '{k}': {p}"); warns += 1
print(f"  {errors} error(s), {warns} warning(s)")
sys.exit(1 if errors else 0)
PY

echo "== 3/4 start a test server on :$PORT (builds dist/)"
FEEDS_AUTO_REFRESH=0 PORT="$PORT" "$OUT/cptx.ee" > "$OUT/server.log" 2>&1 &
pid=$!
trap 'kill $pid 2>/dev/null || true' EXIT
for _ in $(seq 1 120); do
  curl -fsS -o /dev/null "http://127.0.0.1:$PORT/" 2>/dev/null && break
  kill -0 "$pid" 2>/dev/null || { tail -5 "$OUT/server.log"; echo "FAILED: server exited"; exit 1; }
  sleep 1
done
curl -fsS -o /dev/null "http://127.0.0.1:$PORT/" || { echo "FAILED: server not answering"; exit 1; }

echo "== URL diff vs urls.txt"
scripts/urls.sh dist > "$OUT/urls.new"
comm -23 urls.txt "$OUT/urls.new" > "$OUT/removed"
comm -13 urls.txt "$OUT/urls.new" > "$OUT/added"
echo "  $(wc -l < "$OUT/urls.new") URLs; $(wc -l < "$OUT/added") added, $(wc -l < "$OUT/removed") removed"
sed 's/^/  + /' "$OUT/added" | head -20
if [ -s "$OUT/removed" ]; then
  sed 's/^/  - /' "$OUT/removed" | head -20
  if [ "$UPDATE" = 1 ]; then echo "  (removals accepted: --update-urls)"
  else bad "URLs would disappear or change (same-day posts shift IDs). If intended: --update-urls"; fi
fi

echo "== 4/4 HTTP smoke (every URL)"
python3 - "$PORT" "$OUT/urls.new" <<'PY' || fail=1
import http.client, sys, urllib.parse
port, path = int(sys.argv[1]), sys.argv[2]
urls = [l.strip() for l in open(path, encoding="utf-8") if l.strip()]
known = set(urls); bad = []
def get(u):
    h = http.client.HTTPConnection("127.0.0.1", port, timeout=15)
    h.request("GET", urllib.parse.quote(u, safe="/")); r = h.getresponse(); r.read(); h.close()
    return r.status, r.getheader("Location")
for u in urls:
    s, loc = get(u)
    if s == 301 and urllib.parse.unquote(loc or "") in known: continue
    if s != 200: bad.append(f"{s} {u}" + (f" -> {loc}" if loc else ""))
for u, want in (("/shuffle-pool.json", 200), ("/no-such-page/", 404)):
    s, _ = get(u)
    if s != want: bad.append(f"{s} {u} (want {want})")
for b in bad[:20]: print("  FAIL:", b)
print(f"  {len(urls) - len(bad)}/{len(urls)} ok" if not bad else f"  {len(bad)} failing")
sys.exit(1 if bad else 0)
PY

if [ "$UPDATE" = 1 ] && [ "$fail" = 0 ]; then cp "$OUT/urls.new" urls.txt; echo "urls.txt updated"; fi
if [ "$fail" = 0 ]; then echo "CHECK OK"; else echo "CHECK FAILED"; exit 1; fi
