#!/usr/bin/env -S python3 -I
"""Fetch the Salt Palace text pages and the virtual tour script and print candidate places.yaml entries.

The output is a starting point that a human reviews before it goes into data/venue/places.yaml.
Facts the venue states in text are emitted with verified: true and the page URL as source.
Scene labels read from the virtual tour are emitted with verified: false because a label is not a statement.
"""
import argparse
import html
import json
import re
import sys
import urllib.request

BASE = "https://www.visitsaltlake.com/salt-palace-convention-center/"
PAGES = {
    "directory": BASE + "attend/venue-directory-map/",
    "parking": BASE + "attend/parking-information/",
}
TOUR = "https://www.imapenterprises.com/saltpalace/virtualtour/script.js"
UA = "Mozilla/5.0 (X11; Linux x86_64) WideDoor venue-extract/0.1"


def fetch(url):
    req = urllib.request.Request(url, headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=30) as r:
        return r.read().decode("utf-8", errors="replace")


def page_text(doc):
    doc = re.sub(r"(?is)<(script|style|noscript|svg|header|footer|nav)[^>]*>.*?</\1>", "", doc)
    m = re.search(r"(?is)<main[^>]*>(.*?)</main>", doc)
    body = m.group(1) if m else doc
    t = re.sub(r"(?s)<br\s*/?>|</p>|</li>|</h\d>|</tr>|</div>", "\n", body)
    t = html.unescape(re.sub(r"(?s)<[^>]+>", "", t))
    return [ln.strip() for ln in t.splitlines() if ln.strip()]


def sentences(lines):
    for ln in lines:
        for s in re.split(r"(?<=[.!?])\s+", ln):
            if s:
                yield s.strip()


RULES = [
    (re.compile(r"main entrances .*?(East Entrance) from (West Temple)", re.I), "entrance", "entrance-west-temple"),
    (re.compile(r"(South Entrance) from (200 South)", re.I), "entrance", "entrance-200-south"),
    (re.compile(r"entrances on the (North) and (West) sides", re.I), "entrance", "entrance-north-west"),
    (re.compile(r"(200 South Lot).*?Enter off of (.*?)\.", re.I), "parking", "parking-200-south-lot"),
    (re.compile(r"(300 West Lot).*?Enter off of (.*?)\.", re.I), "parking", "parking-300-west-lot"),
    (re.compile(r"Accessible parking available in both lots", re.I), "parking", "parking-accessible"),
    (re.compile(r"Meeting rooms (155, 255 and 355) are stacked", re.I), "landmark", "stack-155-255-355"),
    (re.compile(r"10 exhibit halls", re.I), "landmark", "exhibit-halls"),
    (re.compile(r"Meeting rooms are numbered by floor", re.I), "fact", "room-numbering"),
    (re.compile(r"Bicycles? racks", re.I), "fact", "bike-racks"),
]


def candidates_from_text(name, url, lines):
    out = []
    for s in sentences(lines):
        for rx, kind, pid in RULES:
            if rx.search(s):
                out.append({"id": pid, "kind": kind, "venue": "Salt Palace", "notes": s, "verified": True, "source": url, "page": name})
    return out


def candidates_from_tour(script):
    labels = sorted(set(re.findall(r'"label"\s*:\s*"([^"]{2,60})"', script)))
    skip = re.compile(r"^(Arrow|Polygon|Image|Circle|Info|Photo Album|Level\d|Leve\d)", re.I)
    out = []
    for lb in labels:
        if skip.match(lb):
            continue
        out.append({"id": "tour-" + re.sub(r"[^a-z0-9]+", "-", lb.lower()).strip("-"), "kind": "landmark", "venue": "Salt Palace", "name": lb, "verified": False, "source": TOUR, "notes": "Scene label in the venue's virtual tour; the position was read off a picture, not stated in text."})
    return out


def to_yaml(cands):
    lines = ["# candidate places, review by hand before merging into data/venue/places.yaml", "places:"]
    for c in cands:
        lines.append(f"  - id: {c['id']}")
        for k in ("kind", "name", "venue", "notes", "verified", "source", "page"):
            if k in c:
                v = c[k]
                v = str(v).lower() if isinstance(v, bool) else json.dumps(v)
                lines.append(f"    {k}: {v}")
    return "\n".join(lines) + "\n"


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--only", choices=["directory", "parking", "tour"], action="append", help="limit to one source (repeatable)")
    ap.add_argument("--raw", action="store_true", help="print the extracted page text instead of candidates")
    args = ap.parse_args()
    only = set(args.only or ["directory", "parking", "tour"])
    cands = []
    for name, url in PAGES.items():
        if name not in only:
            continue
        lines = page_text(fetch(url))
        if args.raw:
            print(f"# {url}")
            print("\n".join(lines))
            continue
        cands += candidates_from_text(name, url, lines)
    if "tour" in only and not args.raw:
        cands += candidates_from_tour(fetch(TOUR))
    if not args.raw:
        sys.stdout.write(to_yaml(cands))


if __name__ == "__main__":
    main()
