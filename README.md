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

Windows 交叉编译（需要 `gcc-mingw-w64-x86-64`、`p7zip-full`）：

```sh
make fetch-libmpv          # 从 SourceForge 下载官方 mpv-dev-x86_64，解到 third_party/mpv
make package-windows       # 交叉编译 + 拷贝 libmpv-2.dll 到 build/
makensis packaging/windows/installer.nsi
```

### Windows 本机（Docker，无需 mingw）

Windows 上只有 MSVC 也能出包——Go 的 cgo 不认 MSVC，但可以在官方 `golang` 镜像里交叉编译：

```powershell
./scripts/build-windows-docker.ps1              # build/mvplayer.exe + build/libmpv-2.dll
./scripts/build-windows-docker.ps1 -Installer   # 另出 build/mvPlayer-0.1.0-setup.exe
```

容器内自动装 `gcc-mingw-w64-x86-64`/`p7zip`/`nsis`，apt 与 `GOPROXY` 走国内镜像；
libmpv 默认从 GitHub release 取（`-MpvProxy` 可换加速镜像，`-MpvUrl` 可直接指定）。
源码会复制到容器内并统一成 LF——`core.autocrlf=true` 的 Windows 检出让 `Makefile`/`*.sh` 带 CRLF，
在 Linux 容器里会直接报错。加 `-SkipTests` 可跳过 `go vet`/`go test` 快速迭代。

`libmpv-2.dll`（约 116 MiB）不进仓库，由 `scripts/fetch-libmpv.sh` 按需下载；
安装包会把它和 `mvplayer.exe` 一起打包。手动下载地址：
<https://sourceforge.net/projects/mpv-player-windows/files/libmpv/>。

Linux 运行依赖系统 `libmpv2`（Debian/Ubuntu：`apt install libmpv2`），不需要额外下载 DLL。

## 体积

- 安装后：`mvplayer` 约 30 MiB；Windows 还需 `libmpv-2.dll`。
- 安装包：exe 经 LZMA 后约 9–11 MiB，libmpv 约 25–30 MiB，NSIS 安装包预计 38–45 MiB。


## CI

GitHub Actions：`.github/workflows/ci.yml`

- `linux` job：装 `libmpv-dev`/X11/GL 头，跑 `go vet`（两种 tag）、两套 `go test`、`go build -tags mpv`，上传二进制
- `windows-package` job：下载官方 libmpv → 交叉编译真实 exe → NSIS 打出安装包，产物 `mvplayer-windows`（含 exe + dll + setup）
- Linux 产物是动态链接系统 libmpv 的二进制，运行机器需要 `libmpv2`

本地也可用容器跑同一套检查：

```bash
docker build -f packaging/ci/Dockerfile -t mvplayer-ci .
```

测试不提交任何二进制媒体：单测用临时生成的占位 `.mkv`（只用到路径与扩展名），
真实解码测试需要显式给样本：

```bash
MVPLAYER_SAMPLE=/path/to/real.mkv go test -tags mpv ./...
```

CI 里用 `ffmpeg` 现场生成 `sample.mkv` 再跑真实解码测试。
