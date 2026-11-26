#!/usr/bin/env python3
"""Films import (archives/films): data/library/films.csv -> OUTDIR/films.json + posters.

Run on the VM (the workspace can't reach OMDb):
    python3 scripts/films-import.py ~/cptx-data/library
    cp ~/cptx-data/library/films.json data/library/films.json

Film details and posters come from the Open Movie Database (https://www.omdbapi.com).
OMDb needs a free API key: it is read from $OMDB_API_KEY, else from the file
~/cptx-data/omdb.key (one line). The key never goes into git or into chat.

For each film it looks the title up on OMDb (by the `imdb` pin, else by title + year),
keeps the IMDb id, director, runtime, country and language, and downloads the poster
into OUTDIR/films/posters/<key>.jpg (OMDb's poster URL, asked for at 600 px wide).
A photo of our own copy in OUTDIR/films/photos/<key>.jpg always wins over the poster.
Results are cached in OUTDIR/films-cache.json, so a re-run only looks up new films
(--refresh looks everything up again; --refetch downloads every poster again).

Columns in films.csv: key, title, year, shelf (genre), format (dvd, blu-ray, 4k; "·"
between several), notes (about our copy), search_title (the title to look up when ours
differs from IMDb's), imdb (pin: an IMDb id like tt0080684 wins over the search; "none"
= don't match anything). Films found under a different year, or not found, are reported
and left without details; pin their imdb id to accept a match.
A hand-picked poster goes to OUTDIR/films/posters/<key>.jpg; add the key to the
`keep_poster` list in OUTDIR/films-cache.json (or pin it with a photo) so --refetch
doesn't replace it.
"""
import csv, json, os, re, sys, time, urllib.error, urllib.parse, urllib.request

UA = "cptx-library/0.2 ( https://cptx.ee )"
HERE = os.path.dirname(os.path.abspath(__file__))
CSV = os.path.join(HERE, "..", "data", "library", "films.csv")
FIELDS = ["key", "title", "year", "shelf", "format", "notes", "imdb", "director", "runtime", "country", "language", "poster", "poster_photo"]


def api_key():
    k = os.environ.get("OMDB_API_KEY", "").strip()
    if not k:
        p = os.path.expanduser("~/cptx-data/omdb.key")
        if os.path.exists(p):
            k = open(p).read().strip()
    if not k:
        sys.exit("no OMDb key: put it in ~/cptx-data/omdb.key (one line) or $OMDB_API_KEY")
    return k


def get(url, binary=False):
    for attempt in range(4):
        req = urllib.request.Request(url, headers={"User-Agent": UA})
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                return r.read()
        except urllib.error.HTTPError as e:
            if e.code in (429, 502, 503) and attempt < 3:
                time.sleep(3 * (attempt + 1))
                continue
            raise
        except urllib.error.URLError:
            if attempt < 3:
                time.sleep(2)
                continue
            raise


def omdb(key, **params):
    params["apikey"] = key
    data = json.loads(get("https://www.omdbapi.com/?" + urllib.parse.urlencode(params)))
    if data.get("Response") == "False":
        if "limit" in data.get("Error", "").lower():
            sys.exit("OMDb: " + data["Error"] + " (the free key allows 1,000 requests a day)")
        return None
    return data


def year_of(s):
    m = re.match(r"\d{4}", s or "")
    return int(m.group()) if m else 0


def lookup(key, row):
    kind = "series" if row["shelf"] == "series" and "series" in row["notes"].lower() else "movie"
    if row["imdb"]:
        return omdb(key, i=row["imdb"]), "pinned"
    title, year = row["search_title"] or row["title"], int(row["year"] or 0)
    for params in ({"t": title, "y": year, "type": kind}, {"t": title, "type": kind}):
        if not params.get("y"):
            params.pop("y", None)
        d = omdb(key, **params)
        if d and (not year or abs(year_of(d.get("Year")) - year) <= 1):
            return d, "title"
    found = omdb(key, s=title, type=kind)
    for hit in (found or {}).get("Search", []):
        if not year or abs(year_of(hit.get("Year")) - year) <= 1:
            return omdb(key, i=hit["imdbID"]), "search"
    return None, "no match"


def clean(v):
    return "" if not v or v == "N/A" else v.strip()


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    if len(args) != 1:
        sys.exit(__doc__)
    out = os.path.expanduser(args[0])
    refresh, refetch = "--refresh" in sys.argv, "--refetch" in sys.argv
    key = api_key()
    posters = os.path.join(out, "films", "posters")
    photos = os.path.join(out, "films", "photos")
    os.makedirs(posters, exist_ok=True)
    cache_path = os.path.join(out, "films-cache.json")
    cache = json.load(open(cache_path)) if os.path.exists(cache_path) else {}
    keep = set(cache.get("keep_poster", []))
    films = cache.setdefault("films", {})
    rows = list(csv.DictReader(open(CSV, encoding="utf-8")))
    result, missing = [], []
    for row in rows:
        k = row["key"]
        c = films.get(k)
        if row["imdb"] == "none":
            c = {}
        elif c is None or not c.get("imdb") or refresh or (row["imdb"] and c.get("imdb") != row["imdb"]):
            d, how = lookup(key, row)
            if d:
                c = {"imdb": d["imdbID"], "found_title": d.get("Title", ""), "found_year": d.get("Year", ""),
                     "director": clean(d.get("Director")), "runtime": clean(d.get("Runtime")),
                     "country": clean(d.get("Country")), "language": clean(d.get("Language")),
                     "poster_url": clean(d.get("Poster")), "how": how}
            else:
                c = {}
            films[k] = c
            json.dump(cache, open(cache_path, "w"), ensure_ascii=False, indent=1)
            time.sleep(0.2)
        if not c.get("imdb") and row["imdb"] != "none":
            missing.append(f"{k}: no match for {row['search_title'] or row['title']} ({row['year']})")
        elif c.get("how") in ("search",) or (c.get("found_title") and c["found_title"].lower() != (row["search_title"] or row["title"]).lower()):
            missing.append(f"{k}: check — matched {c.get('found_title')} ({c.get('found_year')}) {c.get('imdb')}")

        # poster: our own photo wins, then OMDb's poster (600 px wide)
        dest = os.path.join(posters, k + ".jpg")
        photo = os.path.join(photos, k + ".jpg")
        is_photo = os.path.exists(photo)
        if is_photo:
            if not os.path.exists(dest) or os.path.getmtime(photo) > os.path.getmtime(dest):
                with open(photo, "rb") as f, open(dest, "wb") as g:
                    g.write(f.read())
        elif c.get("poster_url") and k not in keep and (refetch or not os.path.exists(dest)):
            url = re.sub(r"_SX\d+(\.jpg)$", r"_SX600\1", c["poster_url"])
            img = None
            for u in (url, c["poster_url"]):
                try:
                    img = get(u, binary=True)
                    break
                except Exception as e:
                    print(f"  {k}: poster {e}")
            if img and img[:3] == b"\xff\xd8\xff":
                with open(dest, "wb") as g:
                    g.write(img)
        has = os.path.exists(dest)
        if not has and row["imdb"] != "none":
            missing.append(f"{k}: no poster")
        result.append({
            "key": k, "title": row["title"], "year": int(row["year"] or 0), "shelf": row["shelf"],
            "format": row["format"], "notes": row["notes"], "imdb": c.get("imdb", ""),
            "director": c.get("director", ""), "runtime": c.get("runtime", ""),
            "country": c.get("country", ""), "language": c.get("language", ""),
            "poster": "films/posters/" + k + ".jpg" if has else "", "poster_photo": is_photo,
        })
    cache["keep_poster"] = sorted(keep)
    json.dump(cache, open(cache_path, "w"), ensure_ascii=False, indent=1)
    with open(os.path.join(out, "films.json"), "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=1)
        f.write("\n")
    print(f"{len(result)} films, {sum(1 for r in result if r['poster'])} posters, {sum(1 for r in result if r['imdb'])} matched")
    for m in missing:
        print("  " + m)


if __name__ == "__main__":
    main()
