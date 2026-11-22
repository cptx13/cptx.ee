#!/usr/bin/env python3
"""Vinyl import (archives/vinyl): data/library/vinyl.csv -> OUTDIR/vinyl.json + covers.

Run on the VM (the workspace can't reach MusicBrainz):
    python3 scripts/vinyl-import.py ~/cptx-data/library
    cp ~/cptx-data/library/vinyl.json data/library/vinyl.json

For each record it finds the MusicBrainz release group (or uses the `mbid` column),
downloads the front cover from the Cover Art Archive into OUTDIR/vinyl/covers/<key>.jpg,
and follows MusicBrainz -> Wikidata for the Wikipedia article, the Spotify album and the
YouTube Music playlist. Results are cached in OUTDIR/vinyl-cache.json, so a re-run only
looks up new records (--refresh looks everything up again).

Columns in vinyl.csv: key, artist, title, year, shelf, notes, search_artist, mbid,
spotify, youtube_music, wikipedia. The last four are optional pins: a value there wins
over whatever the lookup finds. A hand-made cover goes to OUTDIR/vinyl/covers/<key>.jpg
(also delete OUTDIR/vinyl/thumbs/<key>.jpg when replacing one).
Matches scoring under 90 are reported and left out; pin their mbid to accept them.
"""
import csv, json, os, sys, time, urllib.error, urllib.parse, urllib.request

UA = "cptx-library/0.2 ( https://cptx.ee )"
MIN_SCORE = 90
HERE = os.path.dirname(os.path.abspath(__file__))
CSV = os.path.join(HERE, "..", "data", "library", "vinyl.csv")
_last = [0.0]


def get(url, mb=False):
    for attempt in range(5):
        if mb:  # MusicBrainz: at most one request per second (a bit slower to be safe)
            wait = 1.5 - (time.time() - _last[0])
            if wait > 0:
                time.sleep(wait)
            _last[0] = time.time()
        req = urllib.request.Request(url, headers={"User-Agent": UA, "Accept": "application/json"})
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                return r.read()
        except urllib.error.HTTPError as e:
            if e.code in (429, 502, 503) and attempt < 4:  # throttled: back off and retry
                time.sleep(3 * (attempt + 1))
                continue
            raise


def lucene(s):
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'


def search(title, artist, single=False):
    q = "releasegroup:" + lucene(title)
    if artist:
        q += " AND artist:" + lucene(artist)
    # albums only (a title search otherwise finds the hit single first), unless the
    # record itself is a single
    q += " AND primarytype:" + ("single" if single else "album")
    url = "https://musicbrainz.org/ws/2/release-group/?" + urllib.parse.urlencode({"query": q, "fmt": "json", "limit": 1})
    groups = json.loads(get(url, mb=True)).get("release-groups", [])
    return groups[0] if groups else None


def wikidata_of(mbid):
    url = f"https://musicbrainz.org/ws/2/release-group/{mbid}?inc=url-rels&fmt=json"
    rg = json.loads(get(url, mb=True))
    for rel in rg.get("relations", []):
        if rel.get("type") == "wikidata":
            return rg, rel["url"]["resource"].rstrip("/").rsplit("/", 1)[-1]
    return rg, ""


def wikidata_links(qid, prefer):
    url = "https://www.wikidata.org/w/api.php?" + urllib.parse.urlencode(
        {"action": "wbgetentities", "ids": qid, "props": "sitelinks/urls|claims", "format": "json"})
    ent = json.loads(get(url))["entities"][qid]
    out = {}
    sl = ent.get("sitelinks", {})
    for lang in prefer + ["en", "es", "et", "de", "fr", "it", "ru"]:
        s = sl.get(lang + "wiki")
        if s:
            out["wikipedia"], out["wikipedia_lang"] = s["url"], lang
            break

    def claim(p):
        for c in ent.get("claims", {}).get(p, []):
            v = c.get("mainsnak", {}).get("datavalue", {}).get("value")
            if isinstance(v, str):
                return v
        return ""

    sp = claim("P2205")  # Spotify album ID
    if sp:
        out["spotify"] = "https://open.spotify.com/album/" + sp
    yt = claim("P4300")  # YouTube playlist ID (YouTube Music albums are playlists)
    if yt:
        out["youtube_music"] = "https://music.youtube.com/playlist?list=" + yt
    return out


def lookup(row, outdir):
    prefer = {"espanol": ["es"], "eesti": ["et"]}.get(row["shelf"], [])
    info = {"mbid": row.get("mbid", ""), "score": 100 if row.get("mbid") else 0}
    if not info["mbid"]:
        rg = search(row["title"], row.get("search_artist", ""), "single" in row.get("notes", "").lower())
        if not rg:
            return {"status": "no match"}
        info["score"] = int(rg.get("score", 0))
        info["match"] = rg["title"] + " — " + (rg.get("artist-credit") or [{}])[0].get("name", "")
        if info["score"] < MIN_SCORE:
            info["status"] = "low score: " + info["match"]
            return info
        info["mbid"] = rg["id"]
    rg, qid = wikidata_of(info["mbid"])
    info["first_year"] = int((rg.get("first-release-date") or "0")[:4] or 0)
    if qid:
        try:
            info.update(wikidata_links(qid, prefer))
        except Exception as e:
            print("  wikidata:", qid, e)
    dest = os.path.join(outdir, "vinyl", "covers", row["key"] + ".jpg")
    old = os.path.join(outdir, "vinyl", "cover-mbid", row["key"])
    if os.path.exists(dest) and (not os.path.exists(old) or open(old).read() != info["mbid"]):
        os.remove(dest)  # the match changed: fetch the new cover, drop the old thumbnail
        thumb = os.path.join(outdir, "vinyl", "thumbs", row["key"] + ".jpg")
        if os.path.exists(thumb):
            os.remove(thumb)
    if not os.path.exists(dest):
        try:
            data = get(f"https://coverartarchive.org/release-group/{info['mbid']}/front-500")
            with open(dest, "wb") as f:
                f.write(data)
            os.makedirs(os.path.dirname(old), exist_ok=True)
            with open(old, "w") as f:
                f.write(info["mbid"])
        except Exception as e:
            info["cover_error"] = str(e)
    info["status"] = "ok"
    return info


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    if len(args) != 1:
        sys.exit(__doc__)
    outdir = os.path.expanduser(args[0])
    os.makedirs(os.path.join(outdir, "vinyl", "covers"), exist_ok=True)
    cache_path = os.path.join(outdir, "vinyl-cache.json")
    cache = {} if "--refresh" in sys.argv or not os.path.exists(cache_path) else json.load(open(cache_path))
    rows = list(csv.DictReader(open(CSV, encoding="utf-8")))
    records, problems = [], []
    for row in rows:
        key = row["key"]
        c = cache.get(key)
        if c is None or c.get("mbid", "") != (row.get("mbid") or c.get("mbid", "")) or c.get("status") != "ok":
            try:
                c = lookup(row, outdir)
            except Exception as e:
                c = {"status": f"error: {e}"}
            cache[key] = c
            print(f"{key}: {c.get('status')}")
        cover = os.path.exists(os.path.join(outdir, "vinyl", "covers", key + ".jpg"))
        if c.get("status") != "ok":
            problems.append(f"{key}: {c.get('status')}")
        elif not cover:
            problems.append(f"{key}: no cover in the Cover Art Archive")
        year = int(row["year"]) if row["year"].strip().isdigit() else c.get("first_year", 0)
        rec = {
            "key": key, "artist": row["artist"], "title": row["title"], "year": year,
            "shelf": row["shelf"], "notes": row["notes"], "mbid": c.get("mbid", ""),
            "spotify": row.get("spotify") or c.get("spotify", ""),
            "youtube_music": row.get("youtube_music") or c.get("youtube_music", ""),
            "wikipedia": row.get("wikipedia") or c.get("wikipedia", ""),
            "cover": "vinyl/covers/" + key + ".jpg" if cover else "",
        }
        records.append(rec)
    json.dump(cache, open(cache_path, "w"), ensure_ascii=False, indent=1)
    with open(os.path.join(outdir, "vinyl.json"), "w", encoding="utf-8") as f:
        json.dump(records, f, ensure_ascii=False, indent=1)
        f.write("\n")
    n = len(records)
    def has(k):
        return sum(1 for r in records if r[k])
    print(f"\n{n} records · covers {has('cover')} · Wikipedia {has('wikipedia')} · Spotify {has('spotify')} · YouTube Music {has('youtube_music')}")
    if problems:
        print("\nto check:\n  " + "\n  ".join(problems))


if __name__ == "__main__":
    main()
