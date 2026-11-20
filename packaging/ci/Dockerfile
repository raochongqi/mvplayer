# CI image for mvPlayer: Go + Fyne (cgo) + libmpv + X11/GL headers.
# Tests run in the `ci` stage; the final scratch stage keeps the pushed image
# small (~30 MB) so Gitee Go's push does not time out.
FROM ubuntu:24.04 AS ci

ENV DEBIAN_FRONTEND=noninteractive

# Use a China mirror for apt (archive.ubuntu.com is often slow from CI).
RUN if [ -f /etc/apt/sources.list.d/ubuntu.sources ]; then \
        sed -i 's|http://archive.ubuntu.com|https://mirrors.aliyun.com|g; s|http://security.ubuntu.com|https://mirrors.aliyun.com|g' /etc/apt/sources.list.d/ubuntu.sources; \
    elif [ -f /etc/apt/sources.list ]; then \
        sed -i 's|http://archive.ubuntu.com|https://mirrors.aliyun.com|g; s|http://security.ubuntu.com|https://mirrors.aliyun.com|g' /etc/apt/sources.list; \
    fi

RUN apt-get update -o Acquire::Retries=3 \
    && apt-get install -y --no-install-recommends \
        ca-certificates curl git build-essential pkg-config \
        libmpv-dev libgl1-mesa-dev xorg-dev libegl1 \
        libxkbcommon-x11-0 libxkbcommon-dev libwayland-dev wayland-protocols \
    && rm -rf /var/lib/apt/lists/*

ARG GO_VERSION=1.27.4
RUN curl -fsSL "https://mirrors.aliyun.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
    || curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
RUN tar -C /usr/local -xzf /tmp/go.tgz && rm -f /tmp/go.tgz

ENV PATH=/usr/local/go/bin:$PATH \
    CGO_ENABLED=1 \
    GOPROXY=https://goproxy.cn,direct \
    GOSUMDB=off

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go vet ./... && go vet -tags mpv ./...
RUN go test ./... && go test -tags mpv ./...
RUN go build -tags mpv -ldflags "-s -w" -o /out/mvplayer ./cmd/mvplayer

# Final image: just the binary, so pushing to the registry is fast.
FROM scratch
COPY --from=ci /out/mvplayer /mvplayer
