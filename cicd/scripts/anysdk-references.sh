#!/usr/bin/env bash
set -euo pipefail

STACKQL_DIR="${1:-.}"
PKG_ALIAS_REGEX="${2:-anysdk}"   # default alias name in imports
OUT_DIR="${3:-/tmp/anysdk-usage}"

mkdir -p "$OUT_DIR"

# 1) Find likely files that reference the alias (fast prefilter)
rg -n --hidden --follow --no-ignore-vcs \
  -g'!.git/**' -g'!vendor/**' -g'!**/*_test.go' \
  --type-add 'go:*.go' -tgo \
  "\b${PKG_ALIAS_REGEX}\." \
  "$STACKQL_DIR" \
  > "$OUT_DIR/anysdk_dot_all.txt" || true

# 2) Extract exact "anysdk.<member>" tokens (member includes chained selectors: Foo.Bar)
#    Examples captured: anysdk.NewClient, anysdk.Provider.Service, anysdk.HTTPInfo
awk '
  {
    line=$0
    while (match(line, /\banysdk\.[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*/, m)) {
      print m[0]
      line=substr(line, RSTART+RLENGTH)
    }
  }
' "$OUT_DIR/anysdk_dot_all.txt" \
  | sort \
  | uniq -c \
  | sort -nr \
  > "$OUT_DIR/anysdk_members_counts.txt" || true

# 3) Also provide a unique list without counts
cut -c9- "$OUT_DIR/anysdk_members_counts.txt" \
  | sed 's/^[[:space:]]*//' \
  > "$OUT_DIR/anysdk_members_unique.txt" || true

# 4) Group occurrences by file for quick browsing
cut -d: -f1 "$OUT_DIR/anysdk_dot_all.txt" \
  | sort | uniq -c | sort -nr \
  > "$OUT_DIR/anysdk_files_counts.txt" || true

echo "Wrote:"
echo "  $OUT_DIR/anysdk_dot_all.txt             (all matches: file:line:text)"
echo "  $OUT_DIR/anysdk_members_counts.txt      (unique anysdk.<member> with counts)"
echo "  $OUT_DIR/anysdk_members_unique.txt      (unique anysdk.<member>)"
echo "  $OUT_DIR/anysdk_files_counts.txt        (files ranked by number of matches)"
