# CI image for mvPlayer: Go + Fyne (cgo) + libmpv + X11/GL headers.
# The build itself runs vet/tests, so a red test fails `docker build`.
FROM golang:1.27-bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        libmpv-dev \
        libgl1-mesa-dev \
        xorg-dev \
        libegl1 \
        libxkbcommon-x11-0 \
        libwayland-dev \
        wayland-protocols \
        libxkbcommon-dev \
        ca-certificates \
        git \
    && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=off

WORKDIR /src

# cache modules separately from sources
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go vet ./... && go vet -tags mpv ./...
RUN go test ./... && go test -tags mpv ./...
RUN go build -tags mpv -ldflags "-s -w" -o /out/mvplayer ./cmd/mvplayer
