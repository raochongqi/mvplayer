# 多路 mkv 播放器：立项与技术栈调研

> 调研日期 2026-11-20。任务：同时播放 1～4 个 mkv、多布局、每路独立倍速/跳转、全局控制、历史记忆、可拖分割线，Windows/Linux（Win7 可选），应用体积 ≤50MiB，候选技术栈 纯 web/HTML / Go / Python，原生优先、WebView 其次。
> 证据等级只有四种：**实测 / 文档 / 推断 / 未核实**。实测均在本机（Ubuntu 24.04 aarch64 容器，8 核，headless）用 Chromium 155（Playwright 内置）与本仓库 `.dsh-tmp/` 下的脚本完成，脚本与产物保留在 `.dsh-tmp/ff/`。

## 结论

1. **≤50MiB 与「自带解码器」互斥，且差距不是一点点。** 实测：mpv 官方 Windows 构建单 `mpv.exe` 就有 **116.2 MiB**；PySide6-Essentials 安装后 **208.8 MiB**（Addons 另 100.1 MiB）；Electron v45 的 win32-x64 包 **151.0 MiB**。这三条路在 50MiB 门口直接出局，没有压缩空间可谈。
2. **能进 50MiB 的只剩「不捆解码器、用系统/浏览器解码」的壳**：纯 web（0 MiB）、Go + Wails（~10MB 二进制 + 系统 WebView）、Rust + Tauri（~6MiB + 系统 WebView）。但它们同时继承浏览器的编码墙。
3. **编码墙是实测出来的，不是传闻**：同一台机器、同一 Chromium，`.mkv` 容器本身能解——H.264/AAC 正常播（49 帧/1.5s、音频有解码字节）、VP9/Opus 正常、10-bit H.264 正常；但 **HEVC 是 0 帧黑屏**（`videoWidth=0`、`totalVideoFrames=0`、`videoBytes=0`，音频照放），**AC3 音轨静音**（`audioBytes=0`，画面正常）。真实 mkv 若以 HEVC/AC3/DTS 为主，纯 web / WebView 两条路都不可用。
4. **功能面在 Chromium 里实测可行**：4 路同时播放（640×360，5 秒各 154–155 帧，掉帧 0–10）；`playbackRate` 0.5/1/2/4/8/12/16 全部生效（实测有效倍率 0.48/0.99/1.94/3.88/7.74/11.54/15，最后一档被 30 秒片长截断），`preservesPitch` 默认 true，8x 时音频仍在解码；File System Access API（选择文件 + 句柄存 IndexedDB）在本机 Chromium 可用。
5. **所以真正的决策点不是「选哪个框架」，而是两个事实**：(a) 用户的 mkv 里 HEVC/AC3 占多少；(b) 50MiB 是硬上限还是优先目标。这两个答案决定走「WebView 壳（≤50MiB，编码受限）」还是「自带 mpv（全编码，≥116MiB）」。**在你回答之前不动手写代码。**

## 证据

| # | 结论 | 等级 | 证据 |
|---|---|---|---|
| 1 | mpv 官方 Windows 构建单 exe 116.2 MiB | 实测 | `mpv-x86_64-20261111-git-bcde6b93cc.7z`（32.6 MiB 压缩包）解包：`mpv.exe` = 121,794,560 B；`d3dcompiler_43.dll` 4.3 MiB |
| 2 | PySide6 Qt 运行时 208.8 MiB | 实测 | PyPI `PySide6-Essentials` 6.12.0 win_amd64 文件清单 2483 个文件求和；`PySide6-Addons` 另 100.1 MiB（含 avcodec-63.dll 13.5 MiB） |
| 3 | Electron 包 151.0 MiB（win32-x64）/ 118.2 MiB（linux-x64） | 实测 | GitHub Releases API `electron/electron` v45.4.1 资产大小 |
| 4 | Chromium 能解 mkv 容器，但 HEVC 黑屏、AC3 静音 | 实测 | 本机 Chromium 155 + 自造 mkv（PyAV/lavc 编码）：`h264_aac.mkv` 640×360 播 49 帧、audioBytes 27,246；`hevc_aac.mkv` `videoWidth=0`、`totalVideoFrames=0`、`videoBytes=0`；`h264_ac3.mkv` 画面 49 帧、`audioBytes=0` |
| 5 | `canPlayType` 同向：HEVC/AC3/EAC3/DTS 全空串 | 实测 | `video/x-matroska; codecs="hvc1…"` → `""`；`…, ac-3` → `""`；`…, ec-3` → `""`；`…, dtsc` → `""`；`avc1.42E01E` → `"probably"` |
| 6 | 4 路同播在 Chromium 可行（360p） | 实测 | 4 个 `<video>` 同源循环播放 5 秒：currentTime 5.01/4.98/4.99/4.97，`getVideoPlaybackQuality` 帧数 155/155/155/154，掉帧 10/1/0/0 |
| 7 | 0.5–16x 倍速生效、音调默认保持、8x 音频不静音 | 实测 | `playbackRate` 设 0.5→16：有效倍率 0.48/0.99/1.94/3.88/7.74/11.54/15；`preservesPitch=true`；音频解码字节随倍率线性增长（8x 时 255,632/2s） |
| 8 | 浏览器本地文件能力够用（仅 Chromium 系） | 实测 | `showOpenFilePicker`/`showSaveFilePicker`/`showDirectoryPicker` 均存在，`indexedDB` 可用，`navigator.storage.persist` 存在；句柄可持久化到 IndexedDB（API 文档级，本机未做重启验证） |
| 9 | mpv 原生就支持本工具需要的全部播放语义 | 文档 | mpv 手册（本地副本 `.dsh-tmp/research/mpv-options.rst`）：`--speed=<0.01-100>`（L262）、`--audio-pitch-correction` 默认开启（L265）、`--hr-seek=<no\|absolute\|yes\|default>`（L395） |
| 10 | Qt6 Multimedia 有 playbackRate 与音调补偿 | 文档 | Qt 6.12 `QMediaPlayer`：`playbackRate`、`playbackOptions`、`PitchCompensationAvailability`（6.10 起）；Linux 走 GStreamer 时需 `gstreamer1.0-libav` 才有好编码覆盖 |
| 11 | Wails 产物约 10MB 量级（不捆 Chromium） | 文档/线索 | Wails 使用系统 WebView（WebView2/WebKitGTK），二进制约 10MB（[Wails 实践文](https://blog.john-pfeiffer.com/building-a-desktop-app-with-golang-and-wails/)）；官方未给基准数字 |
| 12 | Tauri v2 空应用约 3–6.2 MiB | 文档/线索 | [tauri#12820](https://github.com/tauri-apps/tauri/issues/12820) 报告 v1→v2 体积 3MiB→6.2MiB |
| 13 | Chrome 的 HEVC 只在「平台已有解码能力」时可用 | 文档 | [streamingmedia 报道](https://www.streamingmedia.com/Articles/ReadArticle.aspx?ArticleID=155572)、[caniuse HEVC](https://caniuse.com/hevc)：Chrome 1–106 不支持，之后依赖底层平台硬解/系统扩展 |
| 14 | 1080p×4 的实际解码开销 | 未核实 | 本机只测 640×360×4（ARM 容器）；1080p×4 在目标 Windows 机器上的 CPU/硬解表现没测，测法：目标机跑 4 个 1080p 样本 60s，记掉帧与 CPU |
| 15 | libmpv-2.dll 单独体积 | 未核实 | `mpv-dev-x86_64` 包下载被容器中断（4.3/30.3 MiB）；按 mpv.exe 静态链接 116.2 MiB 推断 dll 同量级，需补测 |
| 16 | WebKitGTK（Linux Wails/WebView）对 mkv/HEVC 的实际支持 | 未核实 | 取决于发行版 GStreamer 插件（`libav`/`bad`/`vaapi`）；Ubuntu 24.04 桌面默认装 `plugins-good`（含 matroska），HEVC 需额外插件；未实测 |
| 17 | WebView2 上 HEVC（系统 HEVC 扩展）与 AC3 的行为 | 未核实 | WebView2 是 Chromium 内核，推断同 Chromium 墙；Windows 装 HEVC 扩展后可能可解 HEVC，AC3 仍无解；需真机验证 |

## 方案对比

体积列：**实测**为本机量的官方发行物，**文档**为官方/项目自述，**推断**写明推理链。≤50MiB 一列按「安装后占用」判。

| 方案 | 做法 | 体积 | ≤50MiB | mkv 全编码 | 4 路 + 倍速 | Win7 | 推荐度 |
|---|---|---|---|---|---|---|---|
| **纯 web/HTML** | 静态页/PWA，浏览器里播 `<video>` | 0 MiB（实测：静态资源 <1 MiB） | ✅ | ❌ HEVC 黑屏、AC3 静音（实测） | ✅ 实测（360p×4、0.5–16x） | 取决于浏览器（Chrome 最后支持 Win7 的版本已旧） | ★★★（若编码可接受） |
| **Go + Wails v2** | Go 后端 + 系统 WebView 前端 | ~10MB 二进制（文档/线索）+ 系统 WebView | ✅ | Windows 同 Chromium 墙；Linux 取决于 GStreamer（未核实） | ✅（前端同 web） | ❌ WebView2 已停止 Win7 支持 | ★★★（原生窗口 + 小体积的最佳折中） |
| **Rust + Tauri 2** | 同上，Rust 后端 | ~3–6.2 MiB（文档/线索） | ✅ | 同上 | ✅（前端同 web） | ❌ | ★★（体积最小，但团队要会 Rust） |
| **Python + PySide6 + QtMultimedia（FFmpeg 后端）** | Qt 原生控件 + QMediaPlayer | 208.8 MiB 安装后（实测）；打包只会更大 | ❌ | ✅ FFmpeg 后端可解 mkv/HEVC/AC3（文档） | ✅ playbackRate + 音调补偿（文档） | ❌ Qt6 不支持 Win7 | ★（功能能对，体积必死） |
| **C++/Qt + libmpv** | Qt 窗口 + libmpv 渲染 | mpv 单 exe 116.2 MiB（实测）+ Qt | ❌ | ✅ 最强（mpv 手册） | ✅ 4 路各自独立实例 | Qt5 可，Qt6 不行 | ★★（品质上限最高，但破体积） |
| **Flutter + media_kit** | Flutter 桌面 + libmpv 内核 | Flutter 空应用 ~20MB + libmpv（推断 ≥80MB） | ❌ | ✅（libmpv） | ✅ | ❌ | ★ |
| **Electron** | 内置 Chromium + Node | 151.0 MiB 包（实测） | ❌ | 同 Chromium 墙 | ✅ | 旧版可 | ★ |
| **系统解码原生壳（备选）** | Windows Media Foundation + Linux GStreamer，自绘 4 宫格 | 几 MB（推断） | ✅ | 取决于系统编解码器（Windows 需 HEVC 扩展，Linux 需插件） | 需自己实现播放器 UI/同步 | 可（MF） | ★（体积与编码都赌系统，开发量最大） |

## 关键取舍（给用户看的版本）

- 想要 **≤50MiB**：只能用系统解码，mkv 里 **H.264/AAC、VP9/Opus、10-bit H.264** 没问题；**HEVC、AC3/EAC3/DTS 不行**（实测黑屏/静音，且不报错）。
- 想要 **什么 mkv 都能播**：必须自带解码器，最小也是 mpv（**116.2 MiB**，实测），50MiB 做不到。
- 折中：应用本体 ≤50MiB，但**要求系统已装 mpv/libmpv**（Linux 常见、Windows 不常见），或用 mpv 作为可选后端——这会把「装完即用」变成「先装依赖」。

## 未解决的问题

1. **用户 mkv 的编码分布**（最重要）：有没有 HEVC / AC3 / DTS / 10-bit？拿 3–5 个真实样本 `ffprobe` 即可定。它直接决定 WebView 路线是否成立。
2. **50MiB 的口径**：安装包还是安装后？是否把系统 WebView 运行时算进去（Win10+ 自带 WebView2，不应计入）？
3. **Win7 是不是硬需求**：所有现代 WebView 方案（WebView2 已停 Win7）与 Qt6 都不支持；若要 Win7，只能 Qt5 + mpv 或 MF 老 API，且体积破 50MiB。
4. **1080p×4 的真实开销**：本机只测 360p×4；需在目标 Windows 机器上用真实样本复测。
5. **"关联视频进度"的语义**：是多路之间联动同步，还是每路各记各的进度？两种实现差别很大（同步需要主时钟与漂移校正）。
6. **WebView2 的 HEVC 行为**：Windows 装 HEVC 扩展后 WebView2 能否解 HEVC、AC3 是否仍无解——需真机验证。

---

## 补充：体积口径修正 + 2026 框架实测（2026-11-20T16:00，用户答复后）

用户答复：① 编码分布不能确定；② 50MiB 指**安装包**；③ Win7 不是硬需求；④ 分控与总控同时要（如原需求）。

**修正上一节的判断**：上一节的体积列是「安装后占用」，按安装包口径结论会变——`mpv.exe` 116.2 MiB 经 LZMA 压缩后其官方 7z 包只有 32.6 MiB。所以「自带解码器一定超 50MiB」**不成立**，必须重新按安装包算：

| 组成 | 安装后 | 压缩后（LZMA/7z 量级） | 依据 |
|---|---|---|---|
| mpv.exe（全解码内核） | 116.2 MiB | **≈31–33 MiB** | 实测：官方 7z 包 32.6 MiB（含 exe + d3dcompiler 4.3 MiB + 文档） |
| Qt6 Core+Gui+Widgets+icu+平台插件 | 39.7 MiB | ≈14–18 MiB | 实测：PySide6-Essentials 文件清单筛选求和；LZMA 比例按 mpv 的 0.27–0.35 估 |
| Qt Multimedia + FFmpeg 解码 DLL（Windows） | ≈21.4 MiB | ≈8–9 MiB | 实测：Addons 清单中 avcodec 13.5 + avformat 2.7 + avutil 1.1 + swscale 1.2 + swresample 0.2 + Qt6Multimedia 1.3 + ffmpeg 插件 0.7 |
| Python 运行时（PyInstaller 打进包） | ≈20 MiB | ≈8–9 MiB | 推断（CPython 3.12 标准库裁剪 + PyInstaller 引导） |

### 2026 年框架体积实测（第三方同一台 Windows 11 机器，2026-10-23）

| 框架 | 安装包 | 落地磁盘 | 说明 | 来源 |
|---|---|---|---|---|
| Tauri 2.12 | **1.05 MB**（NSIS） | 3.00 MB | 依赖系统 WebView2 | https://dawloom.com/blog/tauri-vs-electron-vs-wpf-benchmark/ |
| Wails 2.16（Go） | 未打安装包 | **4.50 MB**（zip）/ 11.19 MB 落地 | 依赖系统 WebView2 | 同上 |
| Electron 44.4 | 106.1 MB | 367.4 MB | 自带 Chromium + Node | 同上 |
| WPF .NET10 自包含 | 56–60 MB | — | 仅 Windows | 同上 |
| Avalonia 12.1 | 10.54 MB（依赖 .NET）/ 13.74 MB（AOT） | 26.86 / 34.68 MB | 需 .NET 运行时或 AOT | 同上 |
| Qt（Widgets/Quick） | — | 15–80 MB | 2026 框架对比表 | https://prod0.dev/compare/desktop-frameworks |
| Flutter | — | 18–80 MB | 同上 | 同上 |
| Electron | — | 80–200 MB | 同上 | 同上 |

**平台与现状（2026-11）**：Wails v3.0.0 已稳定（2026-10-26 发布，v3.1.0 于 2026-11-01）——https://wails.io/status/ ；Tauri 2 官方体积说明 https://v2.tauri.app/concept/size/ ；media_kit（Flutter，libmpv 内核）支持 Windows/Linux 视频——https://pub.dev/packages/media_kit ；Fyne 2.8 新增 `canvas.Shader`，已有 fyne-mpv-video 用 libmpv 软件渲染 API 在 Fyne 窗口内播视频，**X11/Wayland/Windows/macOS 同一份代码**（不走 `--wid`）——https://github.com/uidbz/fyne-mpv-video 。

### 修正后的方案表（安装包口径）

| 方案 | 安装包估算 | ≤50MiB 安装包 | mkv 全编码 | 4 路/8x | 备注 |
|---|---|---|---|---|---|
| 纯 web（PWA/本地页） | 0 | ✅ | ❌ HEVC/AC3（实测） | ✅ 实测 | 无安装包，但依赖浏览器 |
| Go + Wails v3（WebView） | ≈5–12 MB | ✅ | ❌ 同上 | ✅ 前端同 web | UI 最好做；编码是硬伤 |
| Rust + Tauri 2（WebView） | ≈1–5 MB | ✅ | ❌ 同上 | ✅ | 体积最小 |
| Python + PySide6 + QtMultimedia（FFmpeg 后端） | **≈30–35 MB** | ✅（估算） | ✅ FFmpeg 解码（Windows 内置 DLL；Linux 依赖系统 FFmpeg） | 待 spike | 开发最快、Qt 原生 UI；4 实例性能未验 |
| Go + Fyne 2.8 + libmpv（软件渲染） | ≈40–45 MB | ✅（估算） | ✅ | 软件渲染 4×1080p CPU 开销待验 | 单窗口、Wayland 可用（有现成 demo） |
| C++/Qt6 + libmpv（QOpenGLWidget） | ≈43–48 MB | ✅（擦线） | ✅ | ✅ 零拷贝 GPU | 性能/UI 最稳，C++ 成本高 |
| Python + PySide6 + libmpv | ≈80–100 MB | ❌ | ✅ | — | PySide6 基础包 78.3 MB 压缩就已超标 |
| Electron | 106 MB | ❌ | ❌ 同 Chromium | ✅ | 直接排除 |
| Flutter + media_kit | 60 MB+（估） | ❌（估） | ✅ | ✅ | 体积擦不上 |

### 结论（修正后）

1. **安装包口径下，「全编码 + ≤50MiB」不再是不可能**：Python+QtMultimedia（≈30–35MB）、Go+Fyne+libmpv（≈40–45MB）、C++/Qt+libmpv（≈43–48MB）都在门口。
2. **WebView 路线（Tauri/Wails/纯 web）体积最漂亮（1–12MB），但 HEVC 黑屏、AC3 静音的实测墙还在**；用户又无法确认编码分布，所以这条路必须配「编码探测 + 明确告知/转码」的兜底，否则可能整工具不可用。
3. **开发成本与风险排序**：Python+QtMultimedia 最快但要验 4 实例与 Linux FFmpeg 依赖；Go+Fyne+libmpv 居中（有 demo 可抄）但要验 4×1080p 软件渲染；C++/Qt+libmpv 最稳但最慢。
4. **建议的验证顺序**：先用用户 3–5 个真实样本 `ffprobe` 定编码；若全是 H.264/AAC，直接 Wails/Tauri（5–12MB）最划算；只要出现 HEVC/AC3，就在 Python+QtMultimedia 与 Go+Fyne+libmpv 之间用 spike 二选一。
