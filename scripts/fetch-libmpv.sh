#!/usr/bin/env bash
# Download the Windows libmpv development package and unpack it into
# third_party/mpv (gitignored). Used by `make package-windows`.
#
# Why not commit it: libmpv-2.dll is ~116 MiB of binary; the repo stays text-only.
#
# Sources, in order of preference:
#   MPV_URL   explicit .7z url — used as-is, no resolution
#   MPV_SOURCE=github     latest asset of MPV_GH_REPO (default repo below)
#   MPV_SOURCE=sourceforge the official SourceForge libmpv feed (default)
# MPV_PROXY prefixes the resolved url, for networks where GitHub is blocked:
#   MPV_PROXY=https://gh-proxy.com/ make fetch-libmpv
set -euo pipefail

dest="${1:-third_party/mpv}"
arch="${MPV_ARCH:-x86_64}"
gh_repo="${MPV_GH_REPO:-shinchiro/mpv-winbuild-cmake}"

if [ -f "$dest/libmpv-2.dll" ] && [ -f "$dest/include/mpv/client.h" ]; then
    echo "libmpv already present in $dest"
    exit 0
fi

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing $1; install it first ($2)" >&2; exit 1; }; }
need curl "apt install curl"
# 7z is only needed for extraction
if ! command -v 7z >/dev/null 2>&1 && ! command -v 7za >/dev/null 2>&1; then
    echo "missing 7z; install p7zip-full (Debian/Ubuntu) or p7zip (macOS)" >&2
    exit 1
fi

# Scratch dir up front: feeds are written to files, not piped into `grep -m1`
# (grep exiting early would SIGPIPE the writer and, under pipefail, poison the
# pipeline — it used to print "printf: write error: Broken pipe" in CI).
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

url="${MPV_URL:-}"
if [ -z "$url" ]; then
    case "${MPV_SOURCE:-sourceforge}" in
        github)
            echo "resolving latest mpv-dev-${arch} from ${gh_repo} ..."
            curl -fsSL "https://api.github.com/repos/${gh_repo}/releases/latest" -o "$tmp/rel.json"
            url=$(grep -m1 -oE "https://[^\"]*mpv-dev-${arch}-[0-9]+-git-[0-9a-f]+\.7z" "$tmp/rel.json" || true)
            if [ -z "$url" ]; then
                echo "could not find mpv-dev-${arch} in the latest ${gh_repo} release" >&2
                exit 1
            fi
            ;;
        *)
            echo "resolving latest mpv-dev-${arch} ..."
            curl -fsSL "https://sourceforge.net/projects/mpv-player-windows/rss?path=/libmpv" -o "$tmp/rss.xml"
            url=$(grep -m1 -oE "https://[^\"<]*mpv-dev-${arch}-[0-9]+-git-[0-9a-f]+\.7z/download" "$tmp/rss.xml" || true)
            if [ -z "$url" ]; then
                echo "could not find mpv-dev-${arch} in the SourceForge feed" >&2
                echo "download manually from https://sourceforge.net/projects/mpv-player-windows/files/libmpv/" >&2
                exit 1
            fi
            ;;
    esac
fi
# Optional mirror prefix (e.g. https://gh-proxy.com/) for restricted networks.
if [ -n "${MPV_PROXY:-}" ] && [ "${url#"$MPV_PROXY"}" = "$url" ]; then
    url="${MPV_PROXY}${url}"
fi

echo "downloading $url"
curl -fsSL --retry 5 --retry-all-errors --retry-delay 3 -o "$tmp/mpv-dev.7z" "$url"

size=$(wc -c <"$tmp/mpv-dev.7z")
if [ "$size" -lt 10000000 ]; then
    echo "download looks wrong (${size} bytes); refusing to unpack" >&2
    exit 1
fi
echo "downloaded $((size / 1048576)) MiB"

mkdir -p "$dest"
if command -v 7z >/dev/null 2>&1; then
    7z x -y -o"$dest" "$tmp/mpv-dev.7z" >/dev/null
else
    7za x -y -o"$dest" "$tmp/mpv-dev.7z" >/dev/null
fi
echo "libmpv unpacked into $dest:"
ls -lh "$dest"/libmpv-2.dll "$dest"/libmpv.dll.a "$dest"/include/mpv/client.h
