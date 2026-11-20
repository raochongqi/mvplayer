# mvPlayer

同时播放 1～4 个 mkv 的小工具：多种布局、每路独立倍速与跳转、全局控制、历史记忆、可拖分割线。
技术栈：Go + Fyne 2.8 + libmpv。

## 构建

依赖（Linux）：`libmpv-dev`、`libgl1-mesa-dev`、`xorg-dev`、`gcc`。
注意：cgo 必须用系统 gcc（本容器里 `gcc` 是 termux 的 clang，需显式指定）。

```sh
make test          # 纯 Go 单测，不需要 libmpv
make build         # 无 mpv 的内存桩版本（UI 调试用）
make build-mpv     # 链接 libmpv 的正式版本
```

Windows 交叉编译：

```sh
# 解压 mpv-dev-x86_64 到 ~/mpv-dev-x86_64（含 include/、libmpv.dll.a、libmpv-2.dll）
make package-windows
makensis packaging/windows/installer.nsi
```

## 体积

- 安装后：`mvplayer` 约 30 MiB；Windows 还需 `libmpv-2.dll`。
- 安装包：exe 经 LZMA 后约 9–11 MiB，libmpv 约 25–30 MiB，NSIS 安装包预计 38–45 MiB。


## CI

GitHub Actions：`.github/workflows/ci.yml`

- `linux` job：装 `libmpv-dev`/X11/GL 头，跑 `go vet`（两种 tag）、两套 `go test`、`go build -tags mpv`，上传二进制
- `windows-cross` job：mingw 交叉编译 Windows exe（不含 libmpv 链接，仅验证编译）

本地也可用容器跑同一套检查：

```bash
docker build -f packaging/ci/Dockerfile -t mvplayer-ci .
```
