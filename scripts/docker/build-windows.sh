#!/usr/bin/env bash
# Runs INSIDE the golang:<ver>-bookworm container (repo bind-mounted at /src).
# Installs mingw-w64 + p7zip, then cross-compiles mvplayer.exe against the
# official Windows libmpv dev package. Host entry point:
#   scripts/build-windows-docker.ps1
#
# Why build from a copy instead of /src directly: the repo is checked out on
# Windows with core.autocrlf=true, so Makefile/*.sh carry CRLF and `make` /
# `bash` choke on them. We stage an LF-normalized copy in /work, keep the 116 MiB
# libmpv cache on the host via a symlink, and copy only the artefacts back.
set -euo pipefail

SRC="${SRC:-/src}"
WORK="${WORK:-/work}"
APT_MIRROR="${APT_MIRROR:-https://mirrors.aliyun.com/debian}"
APT_MIRROR_SEC="${APT_MIRROR_SEC:-https://mirrors.aliyun.com/debian-security}"
WANT_INSTALLER="${WANT_INSTALLER:-0}"
MPV_DEV="third_party/mpv"

step() { printf '\n==> %s\n' "$*"; }

step "go / os"
go version
. /etc/os-release && echo "$PRETTY_NAME $(uname -m) / $(nproc) cpus"

step "stage source into $WORK (LF-normalized)"
rm -rf "$WORK"; mkdir -p "$WORK"
tar -C "$SRC" \
    --exclude=./.git --exclude=./build --exclude=./third_party --exclude=./.dsh-tmp \
    -cf - . | tar -C "$WORK" -xf -
# CRLF -> LF for everything the toolchain parses as text.
find "$WORK" -type f \
    \( -name '*.sh' -o -name 'Makefile' -o -name '*.mk' -o -name '*.nsi' \
       -o -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) \
    -exec sed -i 's/\r$//' {} +
echo "files staged: $(find "$WORK" -type f | wc -l)"
mkdir -p "$SRC/third_party"
ln -sfn "$SRC/third_party" "$WORK/third_party"   # libmpv cache stays on the host
cd "$WORK"

step "apt mirror -> $APT_MIRROR"
for f in /etc/apt/sources.list /etc/apt/sources.list.d/*.sources /etc/apt/sources.list.d/*.list; do
    [ -f "$f" ] || continue
    sed -i \
        -e "s|https\?://deb.debian.org/debian|${APT_MIRROR}|g" \
        -e "s|https\?://security.debian.org/debian-security|${APT_MIRROR_SEC}|g" \
        "$f"
done
grep -rhoE 'https?://[^ ]+' /etc/apt/sources.list /etc/apt/sources.list.d/ 2>/dev/null | sort -u || true

step "apt packages"
export DEBIAN_FRONTEND=noninteractive
# mingw for the cross-build. The X11/Wayland/GL/mpv dev headers are only needed by
# the linux-side `go vet` / `go test` (Fyne + go-gl + libmpv), same set CI installs.
# With -SkipTests they are skipped on purpose: that reproduces the CI packaging job,
# which has no libmpv-dev and must still cross-compile.
pkgs="gcc-mingw-w64-x86-64 p7zip-full make curl ca-certificates file"
if [ "${SKIP_TESTS:-0}" != "1" ]; then
    pkgs="$pkgs libgl1-mesa-dev xorg-dev libegl1 libxkbcommon-x11-0 libxkbcommon-dev libwayland-dev libmpv-dev"
fi
[ "$WANT_INSTALLER" = "1" ] && pkgs="$pkgs nsis"
apt-get update -qq
apt-get install -y -qq --no-install-recommends $pkgs

x86_64-w64-mingw32-gcc --version | head -1
7z i >/dev/null 2>&1 && echo "7z ok"
[ "$WANT_INSTALLER" = "1" ] && makensis -VERSION

step "go env (CN mirrors)"
go env -w GOPROXY="${GOPROXY:-https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct}"
go env -w GOSUMDB=off            # go.sum hashes are still verified
go env -w GOTOOLCHAIN=local
go env GOPROXY GOSUMDB GOTOOLCHAIN CGO_ENABLED

step "fetch official libmpv (cached on the host, gitignored)"
if [ -f "$MPV_DEV/libmpv-2.dll" ] && [ -f "$MPV_DEV/include/mpv/client.h" ]; then
    echo "already cached: $(du -h "$MPV_DEV/libmpv-2.dll" | cut -f1)"
else
    bash scripts/fetch-libmpv.sh "$MPV_DEV"
fi

step "cross-compile windows exe"
make package-windows

step "verify artefact"
file build/mvplayer.exe
x86_64-w64-mingw32-objdump -p build/mvplayer.exe | grep 'DLL Name' | sort || true

# Publish the exe before the linux-side tests: the artefact is the point here.
step "copy artefacts back to $SRC/build"
mkdir -p "$SRC/build"
cp -f build/mvplayer.exe build/libmpv-2.dll "$SRC/build/"
ls -lh "$SRC/build"

# Tests run after the cross-build on purpose: the exe lands even if the linux
# side tests fail, which is the useful signal when debugging the packaging path.
step "go vet (both tag sets) + unit tests"
if [ "${SKIP_TESTS:-0}" = "1" ]; then
    echo "skipped (-SkipTests)"
else
    go vet ./...
    go vet -tags mpv ./...
    go test ./...
fi

if [ "$WANT_INSTALLER" = "1" ]; then
    step "nsis installer"
    makensis packaging/windows/installer.nsi
    cp -f build/*setup.exe "$SRC/build/"
    ls -lh "$SRC/build"
fi
