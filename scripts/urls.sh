#!/usr/bin/env bash
# List every page URL the static build produced (T-03).
#   scripts/urls.sh [DIST_DIR]   (default: dist)  -> sorted URLs on stdout
# The server writes dist/ at startup; each page is dist/<path>/index.html.
# /photos/ is left out: photos live only on the VM, not in git.
set -euo pipefail
dist="${1:-dist}"
[ -f "$dist/index.html" ] || { echo "no $dist/index.html — start the server once to build dist/" >&2; exit 1; }
find "$dist" -name index.html -printf '/%P\n' \
  | sed 's#index\.html$##' \
  | grep -v '^/photos/' \
  | LC_ALL=C sort
