#!/bin/sh
# Writes the license and NOTICE text of every third-party module linked into
# the gibidify binary to the file named by $1 (default THIRD_PARTY_NOTICES).
#
# The MIT and BSD licenses of the dependencies require their notices in binary
# distributions, and yaml.v3's Apache-2.0 part requires its NOTICE file. The
# list comes from `go list -deps` on the main package, so it is the set
# actually compiled in and cannot drift from go.mod.
set -eu

out="${1:-THIRD_PARTY_NOTICES}"

{
  printf 'gibidify includes the following third-party software.\n'
  go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}} {{.Dir}}{{end}}{{end}}' . |
    sort -u |
    while read -r path version dir; do
      printf '\n================================================================\n'
      printf '%s %s\n' "$path" "$version"
      printf '================================================================\n'
      found=0
      for f in "$dir"/LICENSE* "$dir"/LICENCE* "$dir"/COPYING* "$dir"/NOTICE*; do
        [ -f "$f" ] || continue
        found=1
        printf '\n--- %s ---\n\n' "$(basename "$f")"
        cat "$f"
      done
      if [ "$found" -eq 0 ]; then
        echo "no license file found for $path" >&2
        exit 1
      fi
    done
} >"$out"
