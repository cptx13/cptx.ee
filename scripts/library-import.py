#!/usr/bin/env python3
"""Build the /archives/library/ data from a Libib export.

    scripts/library-import.py OUTDIR [--offline] [--refetch]

Reads data/library/libib-books.csv (public fields of the Libib export),
data/library/shelves.tsv, data/library/overrides.tsv and data/library/own-books.tsv.
Writes OUTDIR/books.json and, unless --offline, OUTDIR/covers/<key>.jpg plus a cache of
the raw API answers in OUTDIR/raw/ (so a re-run only fetches new books).

Metadata: Google Books (title, authors, language, categories, pages) and Open Library
(subjects, first published year, cover). Covers: Open Library first, then Google Books.
Priority for every field: overrides.tsv > fetched > Libib.
"""
import csv, hashlib, json, os, re, sys, time, unicodedata, urllib.request, urllib.error

DATA = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "data", "library")
UA = "cptx.ee-library/0.1 (+https://cptx.ee/contact/)"
PLACEHOLDERS = set(os.environ.get("COVER_PLACEHOLDERS", "").split())  # md5s of "no cover" images
GOOGLE_KEY = os.environ.get("GOOGLE_BOOKS_KEY", "")
google_off = False  # set after a 429: Google's anonymous quota is spent, skip it for this run
OL_LANG = {"eng": "en", "spa": "es", "est": "et", "ben": "bn", "ger": "de", "cat": "ca", "fre": "fr", "ita": "it", "por": "pt", "dut": "nl"}


def isbn10to13(x):
    if len(x) != 10:
        return ""
    b = "978" + x[:9]
    s = sum((1 if i % 2 == 0 else 3) * int(c) for i, c in enumerate(b))
    return b + str((10 - s % 10) % 10)


def read_tsv(name):
    rows = []
    with open(os.path.join(DATA, name), encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")
            if line and not line.startswith("#"):
                rows.append(line.split("\t"))
    return rows


# --- cleaning Libib's (Amazon-flavoured) titles and names -------------------------------

PAREN_TAIL = re.compile(r"\s*\(([^()]*)\)\s*$")


def clean_title(t):
    series = ""
    while True:
        m = PAREN_TAIL.search(t)
        if not m:
            break
        inner = m.group(1)
        if not series and re.search(r"book \d|#\d|saga|trilogy|series|throne of|court of|inheritance|song of ice", inner, re.I):
            series = inner
        t = t[: m.start()]
    # bilingual catalogue titles: "Elogio de la locura / Praise of Folly"
    t = re.split(r"\s*/\s+|\s+/\s*", t)[0]
    t = t.strip(" .:-")
    if t.isupper():
        t = t.capitalize()
    return t, series


def clean_name(n):
    n = re.sub(r"\s*\([^)]*\)", "", n).strip(" .")
    n = n.split(" / ")[0]
    if n.isupper():
        n = " ".join(w.capitalize() if len(w) > 2 else w for w in n.split())
    return n


def first_author(creators):
    for c in creators.split(", "):
        c = clean_name(c)
        if c and not re.search(r"S\.L\.|^author$|^unknown", c, re.I):
            return c
    return ""


# --- fetching ----------------------------------------------------------------------------

def get_json(url):
    global google_off
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Accept": "application/json"})
    for attempt in range(3):
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                return json.load(r)
        except urllib.error.HTTPError as e:
            if e.code == 404:
                return None
            if e.code == 429 and "googleapis" in url:
                google_off = True
                print("Google Books quota spent (429): continuing with Open Library only", flush=True)
                return None
            time.sleep(2 + attempt * 3)
        except Exception:
            time.sleep(2 + attempt * 3)
    return None


BROWSER_UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"


def amazon_asin(link):
    """Follow an amzn.eu short link to the product page and read its ASIN."""
    try:
        req = urllib.request.Request(link, headers={"User-Agent": BROWSER_UA})
        with urllib.request.urlopen(req, timeout=30) as r:
            m = re.search(r"/(?:dp|gp/product)/([A-Z0-9]{10})", r.geturl())
            return m.group(1) if m else ""
    except urllib.error.HTTPError as e:
        m = re.search(r"/(?:dp|gp/product)/([A-Z0-9]{10})", e.geturl() or "")
        return m.group(1) if m else ""
    except Exception:
        return ""


def get_bytes(url):
    req = urllib.request.Request(url, headers={"User-Agent": BROWSER_UA if "amazon" in url else UA})
    try:
        with urllib.request.urlopen(req, timeout=15) as r:
            if r.status == 200 and r.headers.get("Content-Type", "").startswith("image/"):
                return r.read()
    except Exception:
        pass
    return None


def fetch_raw(key, isbn, rawdir, refetch):
    path = os.path.join(rawdir, key + ".json")
    raw = None
    if os.path.exists(path) and not refetch:
        with open(path, encoding="utf-8") as f:
            raw = json.load(f)
        if raw.get("google_tried", True) or google_off or not isbn:
            return raw
    if raw is None:
        raw = {"google": None, "openlibrary": None, "google_tried": False}
    if isbn and not google_off:
        g = get_json("https://www.googleapis.com/books/v1/volumes?q=isbn:" + isbn + ("&key=" + GOOGLE_KEY if GOOGLE_KEY else ""))
        if not google_off:
            raw["google_tried"] = True
            if g and g.get("items"):
                raw["google"] = g["items"][0]
        time.sleep(0.4)
    if isbn and raw.get("openlibrary") is None:
        o = get_json("https://openlibrary.org/search.json?isbn=" + isbn +
                     "&fields=key,title,author_name,first_publish_year,subject,cover_i,number_of_pages_median,language")
        if o and o.get("docs"):
            raw["openlibrary"] = o["docs"][0]
        time.sleep(0.4)
    with open(path, "w", encoding="utf-8") as f:
        json.dump(raw, f, ensure_ascii=False)
    return raw


def fetch_cover(key, isbn, raw, coverdir, refetch, isbn10="", asin=""):
    path = os.path.join(coverdir, key + ".jpg")
    if os.path.exists(path) and not refetch:
        return "cached"
    tries = []
    if isbn:
        tries.append(("openlibrary", "https://covers.openlibrary.org/b/isbn/%s-L.jpg?default=false" % isbn))
    ol = raw.get("openlibrary") or {}
    if ol.get("cover_i"):
        tries.append(("openlibrary", "https://covers.openlibrary.org/b/id/%s-L.jpg" % ol["cover_i"]))
    links = ((raw.get("google") or {}).get("volumeInfo") or {}).get("imageLinks") or {}
    g = links.get("thumbnail") or links.get("smallThumbnail")
    if g:
        g = g.replace("http://", "https://").replace("&edge=curl", "")
        tries.append(("google", g + "&fife=w600"))
        tries.append(("google", g))
    for a in [x for x in (asin, isbn10) if x]:
        tries.append(("amazon", "https://images-na.ssl-images-amazon.com/images/P/%s.01._SCLZZZZZZZ_.jpg" % a))
        tries.append(("amazon", "https://m.media-amazon.com/images/P/%s.01._SCLZZZZZZZ_.jpg" % a))
    if isbn.startswith("97884"):  # Casa del Libro, good for Spanish editions
        tries.append(("casadellibro", "https://imagessl.casadellibro.com/a/l/t0/%s/%s.jpg" % (isbn[-2:], isbn)))
    if isbn:  # Google's public cover image by ISBN (no API quota)
        tries.append(("google", "https://books.google.com/books/content?vid=ISBN%s&printsec=frontcover&img=1&zoom=1" % isbn))
    for src, url in tries:
        b = get_bytes(url)
        time.sleep(0.3)
        if b and len(b) > 2000 and hashlib.md5(b).hexdigest() not in PLACEHOLDERS:
            with open(path, "wb") as f:
                f.write(b)
            return src
    return ""


# --- language from ISBN when nothing better is known --------------------------------------

def lang_from_isbn(isbn):
    for prefix, lang in (("97884", "es"), ("978950", "es"), ("9789985", "et"), ("9789949", "et"),
                         ("978984", "bn"), ("9783", "de"), ("97890", "nl"), ("9780", "en"),
                         ("9781", "en"), ("9798", "en")):
        if isbn.startswith(prefix):
            return lang
    return ""


def year_of(s):
    m = re.match(r"(\d{4})", s or "")
    return int(m.group(1)) if m else 0


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    offline, refetch = "--offline" in sys.argv, "--refetch" in sys.argv
    if not args:
        sys.exit(__doc__)
    out = args[0]
    rawdir, coverdir = os.path.join(out, "raw"), os.path.join(out, "covers")
    os.makedirs(rawdir, exist_ok=True)
    os.makedirs(coverdir, exist_ok=True)

    shelves = {k: v for k, v in read_tsv("shelves.tsv")}
    own = {k: v for k, v in read_tsv("own-books.tsv")}
    over = {}
    for k, field, value in read_tsv("overrides.tsv"):
        over.setdefault(k, {})[field] = value

    books = []
    with open(os.path.join(DATA, "libib-books.csv"), encoding="utf-8") as f:
        rows = list(csv.DictReader(f))
    for i, r in enumerate(rows, 1):
        isbn = r["isbn13"] or isbn10to13(r["isbn10"])
        key = isbn or r["libib_id"]
        title, series = clean_title(r["title"])
        b = {
            "key": key, "isbn13": isbn, "isbn10": r["isbn10"],
            "title": title, "subtitle": "", "series": series,
            "author": first_author(r["creators"]),
            "publisher": clean_name(r["publisher"]) if r["publisher"].isupper() else r["publisher"],
            "year": year_of(r["publish_date"]), "first_year": 0,
            "pages": int(r["pages"]) if r["pages"].isdigit() else 0,
            "lang": lang_from_isbn(isbn), "subjects": [], "categories": [],
            "shelf": shelves.get(key, ""), "own": key in own, "buy": own.get(key, ""),
            "cover": "", "openlibrary": "", "google": "",
        }
        if not offline:
            raw = fetch_raw(key, isbn, rawdir, refetch)
            g = (raw.get("google") or {}).get("volumeInfo") or {}
            o = raw.get("openlibrary") or {}
            if g.get("title"):
                b["title"], b["subtitle"] = g["title"], g.get("subtitle", "")
            if g.get("authors"):
                b["author"] = " & ".join(g["authors"][:2])
            elif o.get("author_name"):
                b["author"] = " & ".join(o["author_name"][:2])
            ol_langs = [OL_LANG.get(l, "") for l in o.get("language", [])]
            if not b["lang"] and len(ol_langs) == 1:
                b["lang"] = ol_langs[0]
            b["lang"] = g.get("language") or b["lang"]
            b["pages"] = g.get("pageCount") or b["pages"] or o.get("number_of_pages_median") or 0
            b["year"] = b["year"] or year_of(g.get("publishedDate"))
            b["first_year"] = o.get("first_publish_year") or 0
            b["categories"] = g.get("categories", [])
            b["subjects"] = [s for s in o.get("subject", []) if len(s) < 40][:12]
            if o.get("key"):
                b["openlibrary"] = "https://openlibrary.org" + o["key"]
            if raw.get("google"):
                b["google"] = (raw["google"].get("volumeInfo") or {}).get("infoLink", "")
            asin = amazon_asin(own[key]) if key in own else ""
            b["cover"] = fetch_cover(key, isbn, raw, coverdir, refetch, r["isbn10"], asin)
            if b["cover"] == "cached":
                b["cover"] = "yes"
            print("%3d/%d %-14s %-8s %s" % (i, len(rows), key, b["cover"] or "-", b["title"][:50]), flush=True)
        elif os.path.exists(os.path.join(coverdir, key + ".jpg")):
            b["cover"] = "yes"
        for field, value in over.get(key, {}).items():
            if field in ("year", "first_year", "pages"):
                b[field] = int(value)
            elif field == "cover":
                b["cover"] = "" if value == "none" else value
            else:
                b[field] = "" if value == "-" else value
        b["title"] = unicodedata.normalize("NFC", b["title"])
        books.append(b)

    with open(os.path.join(out, "books.json"), "w", encoding="utf-8") as f:
        json.dump(books, f, ensure_ascii=False, indent=1)
    have = sum(1 for b in books if b["cover"])
    print("%d books, %d with covers, %d without shelf" % (len(books), have, sum(1 for b in books if not b["shelf"])))


if __name__ == "__main__":
    main()
