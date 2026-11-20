# mvPlayer build entry points.
# Linux builds need libmpv-dev + X11/GL headers; cgo must use the system gcc
# (the container ships termux clang as `gcc`, which cannot see Ubuntu headers).
GO        ?= go
CC_LINUX  ?= /usr/bin/gcc-13
BUILD     ?= build
MPV_DEV     ?= third_party/mpv
MPV_DEV_ABS := $(abspath $(MPV_DEV))

.PHONY: test build build-mpv run fetch-libmpv package-windows clean

test:
	CC=$(CC_LINUX) $(GO) test ./...

build:
	CC=$(CC_LINUX) $(GO) build -o $(BUILD)/mvplayer ./cmd/mvplayer

build-mpv:
	CC=$(CC_LINUX) $(GO) build -tags mpv -o $(BUILD)/mvplayer ./cmd/mvplayer

run: build-mpv
	./$(BUILD)/mvplayer

# Downloads the official Windows libmpv dev package into third_party/mpv.
# The repo does not commit the 116 MiB DLL; the script fetches it on demand.
fetch-libmpv:
	bash scripts/fetch-libmpv.sh $(MPV_DEV)

# Cross-compiled Windows build. Needs mingw-w64 + the libmpv dev package.
package-windows: fetch-libmpv
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
	CGO_CFLAGS="-I$(MPV_DEV_ABS)/include" CGO_LDFLAGS="-L$(MPV_DEV_ABS) -lmpv" \
	$(GO) build -tags mpv -ldflags "-H windowsgui -s -w" \
	-o $(BUILD)/mvplayer.exe ./cmd/mvplayer
	cp $(MPV_DEV)/libmpv-2.dll $(BUILD)/
	@ls -lh $(BUILD)/mvplayer.exe $(BUILD)/libmpv-2.dll

package-linux: build-mpv
	@echo "build/ binary ready; see packaging/linux/README.md for the AppImage step"

clean:
	rm -rf $(BUILD)
