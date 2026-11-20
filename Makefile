# mvPlayer build entry points.
# Linux builds need libmpv-dev + X11/GL headers; cgo must use the system gcc
# (the container ships termux clang as `gcc`, which cannot see Ubuntu headers).
GO        ?= go
CC_LINUX  ?= /usr/bin/gcc-13
BUILD     ?= build
MPV_DEV   ?= $(HOME)/mpv-dev-x86_64

.PHONY: test build build-mpv run package-windows clean

test:
	CC=$(CC_LINUX) $(GO) test ./...

build:
	CC=$(CC_LINUX) $(GO) build -o $(BUILD)/mvplayer ./cmd/mvplayer

build-mpv:
	CC=$(CC_LINUX) $(GO) build -tags mpv -o $(BUILD)/mvplayer ./cmd/mvplayer

run: build-mpv
	./$(BUILD)/mvplayer

# Cross-compiled Windows build. MPV_DEV must point at the extracted
# mpv-dev-x86_64 package (include/ + libmpv.dll.a + libmpv-2.dll).
package-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
	CGO_CFLAGS="-I$(MPV_DEV)/include" CGO_LDFLAGS="-L$(MPV_DEV) -lmpv" \
	$(GO) build -tags mpv -ldflags "-H windowsgui -s -w" \
	-o $(BUILD)/mvplayer.exe ./cmd/mvplayer
	cp $(MPV_DEV)/libmpv-2.dll $(BUILD)/
	@ls -lh $(BUILD)/mvplayer.exe $(BUILD)/libmpv-2.dll

package-linux: build-mpv
	@echo "build/ binary ready; see packaging/linux/README.md for the AppImage step"

clean:
	rm -rf $(BUILD)
