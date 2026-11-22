#!/usr/bin/env python3
"""Vinyl import (archives/vinyl): data/library/vinyl.csv -> OUTDIR/vinyl.json + covers.

Run on the VM (the workspace can't reach MusicBrainz):
    python3 scripts/vinyl-import.py ~/cptx-data/library
    cp ~/cptx-data/library/vinyl.json data/library/vinyl.json

For each record it finds the MusicBrainz release group (or uses the `mbid` column),
downloads the front cover from the Cover Art Archive into OUTDIR/vinyl/covers/<key>.jpg,
follows MusicBrainz -> Wikidata for the Spotify album and the YouTube Music playlist, and
looks up the artist's Wikipedia page (Spanish/Estonian first for those shelves). A photo
of our own copy in OUTDIR/vinyl/photos/<key>.jpg always wins over the archive's cover. Results are cached in OUTDIR/vinyl-cache.json, so a re-run only
looks up new records (--refresh looks everything up again).

Columns in vinyl.csv: key, artist, title, year, shelf (genre), lang (sung language, e.g. es), notes, search_artist, wiki_artist,
mbid, spotify, youtube_music, artist_wikipedia. wiki_artist names the artist to look up
on Wikipedia when the artist field won't do (a composer, a duo); the last four are pins
(mbid "none" = don't match an album at all): a value there wins
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


def sitelink(qid, prefer):
    url = "https://www.wikidata.org/w/api.php?" + urllib.parse.urlencode(
        {"action": "wbgetentities", "ids": qid, "props": "sitelinks/urls", "format": "json"})
    sl = json.loads(get(url))["entities"][qid].get("sitelinks", {})
    for lang in prefer + ["en", "es", "et", "de", "fr", "it", "ru"]:
        if sl.get(lang + "wiki"):
            return sl[lang + "wiki"]["url"]
    return ""


def artist_wikipedia(name, prefer):
    """Wikipedia page of an artist: MusicBrainz artist search -> Wikidata -> sitelink."""
    url = "https://musicbrainz.org/ws/2/artist/?" + urllib.parse.urlencode({"query": "artist:" + lucene(name), "fmt": "json", "limit": 1})
    found = json.loads(get(url, mb=True)).get("artists", [])
    if not found or int(found[0].get("score", 0)) < MIN_SCORE:
        return {"status": "no artist match"}
    a = json.loads(get(f"https://musicbrainz.org/ws/2/artist/{found[0]['id']}?inc=url-rels&fmt=json", mb=True))
    for rel in a.get("relations", []):
        if rel.get("type") == "wikidata":
            qid = rel["url"]["resource"].rstrip("/").rsplit("/", 1)[-1]
            return {"status": "ok", "artist_mbid": a["id"], "url": sitelink(qid, prefer)}
    return {"status": "ok", "artist_mbid": a["id"], "url": ""}


def lookup(row, outdir):
    prefer = [row["lang"]] if row.get("lang") else []
    if row.get("mbid") == "none":  # no album match wanted (the search found the wrong one)
        return {"status": "ok", "mbid": ""}
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
    if os.path.exists(os.path.join(outdir, "vinyl", "photos", row["key"] + ".jpg")):
        info["status"] = "ok"  # a photo of our copy wins over the archive's cover
        return info
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


def use_photo(outdir, key):
    """A photo of our copy (vinyl/photos/<key>.jpg) becomes the cover."""
    photo = os.path.join(outdir, "vinyl", "photos", key + ".jpg")
    dest = os.path.join(outdir, "vinyl", "covers", key + ".jpg")
    if not os.path.exists(photo):
        return
    data = open(photo, "rb").read()
    if os.path.exists(dest) and open(dest, "rb").read() == data:
        return
    with open(dest, "wb") as f:
        f.write(data)
    thumb = os.path.join(outdir, "vinyl", "thumbs", key + ".jpg")
    if os.path.exists(thumb):
        os.remove(thumb)
    os.makedirs(os.path.join(outdir, "vinyl", "cover-mbid"), exist_ok=True)
    with open(os.path.join(outdir, "vinyl", "cover-mbid", key), "w") as f:
        f.write("photo")


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
    artists = cache.setdefault("_artists", {})
    for row in rows:
        key = row["key"]
        use_photo(outdir, key)
        c = cache.get(key)
        if c is None or c.get("mbid", "") != (row.get("mbid") or c.get("mbid", "")) or c.get("status") != "ok":
            try:
                c = lookup(row, outdir)
            except Exception as e:
                c = {"status": f"error: {e}"}
            cache[key] = c
            print(f"{key}: {c.get('status')}", flush=True)
            json.dump(cache, open(cache_path, "w"), ensure_ascii=False, indent=1)  # keep progress if stopped
        cover = os.path.exists(os.path.join(outdir, "vinyl", "covers", key + ".jpg"))
        photo = os.path.exists(os.path.join(outdir, "vinyl", "photos", key + ".jpg"))
        name = (row.get("wiki_artist") or row["artist"].split(" / ")[0]).strip()
        aw = {}
        if name and name != "Various Artists":
            aw = artists.get(name) or {}
            if aw.get("status") != "ok":
                prefer = [row["lang"]] if row.get("lang") else []
                try:
                    aw = artist_wikipedia(name, prefer)
                except Exception as e:
                    aw = {"status": f"error: {e}"}
                artists[name] = aw
                print(f"  artist {name}: {aw.get('status')} {aw.get('url', '')}", flush=True)
            if aw.get("status") != "ok":
                problems.append(f"{key}: artist {name}: {aw.get('status')}")
        if c.get("status") != "ok":
            problems.append(f"{key}: {c.get('status')}")
        elif not cover and not photo:
            problems.append(f"{key}: no cover in the Cover Art Archive")
        year = int(row["year"]) if row["year"].strip().isdigit() else c.get("first_year", 0)
        rec = {
            "key": key, "artist": row["artist"], "title": row["title"], "year": year,
            "shelf": row["shelf"], "lang": row.get("lang", ""), "notes": row["notes"], "mbid": c.get("mbid", ""),
            "spotify": row.get("spotify") or c.get("spotify", ""),
            "youtube_music": row.get("youtube_music") or c.get("youtube_music", ""),
            "artist_wikipedia": row.get("artist_wikipedia") or aw.get("url", ""),
            "cover": "vinyl/covers/" + key + ".jpg" if cover else "",
            "cover_photo": photo,
        }
        records.append(rec)
    json.dump(cache, open(cache_path, "w"), ensure_ascii=False, indent=1)
    with open(os.path.join(outdir, "vinyl.json"), "w", encoding="utf-8") as f:
        json.dump(records, f, ensure_ascii=False, indent=1)
        f.write("\n")
    n = len(records)
    def has(k):
        return sum(1 for r in records if r[k])
    print(f"\n{n} records · covers {has('cover')} (photos {has('cover_photo')}) · artist Wikipedia {has('artist_wikipedia')} · Spotify {has('spotify')} · YouTube Music {has('youtube_music')}")
    if problems:
        print("\nto check:\n  " + "\n  ".join(problems))


if __name__ == "__main__":
    main()
