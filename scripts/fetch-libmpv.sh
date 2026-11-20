#!/usr/bin/env bash
# Download the Windows libmpv development package and unpack it into
# third_party/mpv (gitignored). Used by `make package-windows`.
#
# Why not commit it: libmpv-2.dll is ~116 MiB of binary; the repo stays text-only.
# Source: official mpv Windows builds (SourceForge, shinchiro/mpv-winbuild-cmake).
set -euo pipefail

dest="${1:-third_party/mpv}"
arch="${MPV_ARCH:-x86_64}"

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

echo "resolving latest mpv-dev-${arch} ..."
rss=$(curl -fsSL "https://sourceforge.net/projects/mpv-player-windows/rss?path=/libmpv")
# -m1 stops at the first match; a `| head -1` here trips pipefail with SIGPIPE.
url=$(printf '%s' "$rss" \
    | grep -m1 -oE "https://[^\"<]*mpv-dev-${arch}-[0-9]+-git-[0-9a-f]+\.7z/download" || true)
if [ -z "$url" ]; then
    echo "could not find mpv-dev-${arch} in the SourceForge feed" >&2
    echo "download manually from https://sourceforge.net/projects/mpv-player-windows/files/libmpv/" >&2
    exit 1
fi
echo "downloading $url"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fL --retry 3 -o "$tmp/mpv-dev.7z" "$url"
mkdir -p "$dest"
if command -v 7z >/dev/null 2>&1; then
    7z x -y -o"$dest" "$tmp/mpv-dev.7z" >/dev/null
else
    7za x -y -o"$dest" "$tmp/mpv-dev.7z" >/dev/null
fi
echo "libmpv unpacked into $dest:"
ls -lh "$dest"/libmpv-2.dll "$dest"/libmpv.dll.a "$dest"/include/mpv/client.h
