#!/usr/bin/env bash
# T-08: Lighthouse scores for a few live pages, as a markdown table.
#   scripts/lighthouse.sh [BASE_URL]      (default https://cptx.ee)
# Needs Node (npx) and Chrome. Informational: always exits 0.
set -uo pipefail
base="${1:-https://cptx.ee}"
pages=("/" "/croquis/202610071/" "/categories/recuerdos/" "/archives/library/")
echo "| page | performance | accessibility | best practices | SEO | LCP | CLS |"
echo "|---|---|---|---|---|---|---|"
for p in "${pages[@]}"; do
  out=$(mktemp)
  if npx -y lighthouse@12 "$base$p" --quiet --output=json --output-path="$out" \
      --chrome-flags="--headless=new --no-sandbox" --only-categories=performance,accessibility,best-practices,seo >/dev/null 2>&1; then
    jq -r --arg p "$p" '[$p,
      (.categories.performance.score*100|round), (.categories.accessibility.score*100|round),
      (.categories["best-practices"].score*100|round), (.categories.seo.score*100|round),
      .audits["largest-contentful-paint"].displayValue, .audits["cumulative-layout-shift"].displayValue]
      | "| " + (map(tostring) | join(" | ")) + " |"' "$out"
  else
    echo "| $p | failed | | | | | |"
  fi
  rm -f "$out"
done
exit 0
