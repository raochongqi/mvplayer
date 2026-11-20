# 任务文档：tech-stack-research

状态：进行中
分支：`task/tech-stack-research`
worktree：`/root/wt/tech-stack-research`

---

## §0 原始材料

<!-- parent 贴：用户原话逐字 + 时间 + 作品件路径。不写需求、不写方案、不做事实核对。
     owner 只读这一节，别改它。 -->

```
用户原话（2026-11-20T13:11）：
「打算编写一个小工具，目标：支持同时播放1～4个mkv视频文件，且可关联视频进度。指标：1.支持一次选择多个视频播放，2.支持不同播放布局，1视频为独占界面，2视频支持左右/上下布局，3视频支持左加右上/右下、右+左上左下、四宫格平分但空一格，4视频支持四宫格布局；3.每个视频支持独立的0.5～8倍速，（可选的，支持小数倍速和自定义倍速）；4.每个视频支持独立前进/后退10/30/60/120秒（可选的，支持自定义前进/后退秒数）5.支持Windows/Linux（可选的，支持Windows7）；6.支持全局暂停，支持全局前进/后退/叠加倍速，7.支持历史播放记忆，8.支持用户拖动分割线调整布局，视频同时自适应调整分辨率，9.界面美观简洁，减少ai感。现在进行立项与技术栈调研。请派子代理进行调研，潜在的技术栈：纯web/HTML，go，Python；原生跨平台优先，WebView方案其次。请进行搜索、调查、比对」

作品件：无
```

用户什么都没说就只写一行：`任务：tech-stack-research`

## §1 任务书与规划

<!-- owner 写。这一节同时是任务书：需求带引用，闸门查的就是它。 -->

### 1.0 定级

定级：T3 —— 立项/架构选型：全新技术栈、多方案取舍、后续实施必然 >8 文件且跨层，先调研设计，定案后分期走 T2。

### 1.1 需求（每条带来源）

- 同时播放 1～4 个 .mkv 视频文件，且可关联视频进度 [→ 用户原话 2026-11-20T13:11]
- 支持一次选择多个视频文件并播放 [→ 用户原话 2026-11-20T13:11]
- 1 路视频时独占界面 [→ 用户原话 2026-11-20T13:11]
- 2 路视频支持左右、上下两种布局 [→ 用户原话 2026-11-20T13:11]
- 3 路视频支持「左 + 右上」「右 + 左上」「四宫格平分但空一格」三种布局 [→ 用户原话 2026-11-20T13:11]
- 4 路视频支持四宫格布局 [→ 用户原话 2026-11-20T13:11]
- 每个视频支持独立的 0.5～8 倍速 [→ 用户原话 2026-11-20T13:11]
- （可选）支持小数倍速与自定义倍速 [→ 用户原话 2026-11-20T13:11]
- 每个视频支持独立前进/后退 10/30/60/120 秒 [→ 用户原话 2026-11-20T13:11]
- （可选）支持自定义前进/后退秒数 [→ 用户原话 2026-11-20T13:11]
- 支持 Windows 与 Linux 两个平台 [→ 用户原话 2026-11-20T13:11]
- Windows 7 不是硬需求，能支持更好 [→ 用户原话 2026-11-20T15:36]
- 支持全局暂停 [→ 用户原话 2026-11-20T13:11]
- 支持全局前进/后退与全局叠加倍速 [→ 用户原话 2026-11-20T13:11]
- 每路独立控制（分控）与全局控制（总控）同时具备 [→ 用户原话 2026-11-20T15:36]
- 快捷键全局控制：空格 = 全局暂停，←/→ = 全局前进/后退 10 秒（固定 10s，不跟随档位下拉） [→ 用户原话 2026-11-20T17:10]
- 支持历史播放记忆 [→ 用户原话 2026-11-20T13:11]
- 支持用户拖动分割线调整布局，视频同时自适应调整分辨率 [→ 用户原话 2026-11-20T13:11]
- 界面美观简洁、减少 AI 感 [→ 用户原话 2026-11-20T13:11]
- 技术栈候选限定为纯 web/HTML、Go、Python，原生跨平台优先、WebView 方案其次 [→ 用户原话 2026-11-20T13:11]
- 应用体积不超过 50MiB，且按安装包体积计 [→ 用户原话 2026-11-20T13:30 + 2026-11-20T15:36 口径澄清]
- 立项阶段先完成技术栈调研与比对，再进入实施 [→ 用户原话 2026-11-20T13:11]

外观：批准件 .docs/design/tech-stack-research/approved.png

```yaml
design: .docs/design/tech-stack-research/approved.png
sha256: c84f937833128751656d8d5d9b185303bfaa065f5efb7284173e9886b4e33163
viewport: 1440x900
dpr: 1
theme: dark-neutral (#17181A bg / #1A1C1F chrome / #1E2023 pane / #2A2D31 line / #D7D9DD text / #8A8F98 muted / #5B8CFF accent)
font: system-ui / Noto Sans CJK SC
fixture: 4-up grid, 4 mkv, toolbar with all controls visible
approved_at: 2026-11-20T17:04
approved_by: user
approved_quote: 「1可以，2装，要能过编译」
```

### 1.1b 用户已裁定（2026-11-20T15:36，不属于需求）

1. 真实 mkv 的编码分布**不能确定**——因此路线必须能覆盖常见 mkv 编码，或提供编码探测与明确告知/转码兜底。
2. 50MiB 指**安装包**体积（不是安装后占用）。
3. Windows 7 **不是硬需求**，能支持更好。
4. 「关联视频进度」= **分控 + 总控同时具备**，如需求所述。

### 1.2 事实核对

调研全文与证据分级见 `.docs/investigations/tech-stack-research.md`；以下是 §1 直接引用的证据：

1. **体积口径会反转结论（实测）**：`mpv.exe` 安装后 116.2 MiB，但官方 7z 包（含 exe + d3dcompiler + 文档）只有 **32.6 MiB**；PySide6 的 Qt6 Core/Gui/Widgets + icu + 平台插件安装后 **39.7 MiB**（Qt Multimedia 与 FFmpeg DLL 另 **21.4 MiB**）。按安装包口径，自带解码器的方案可以进 50MiB。
2. **2026 年框架体积实测（同一台 Win11 机器，2026-10-23）**：Tauri 2.12 安装包 **1.05 MB**、落地 3.00 MB；Wails 2.16 zip **4.50 MB**、落地 11.19 MB；Electron 44.4 安装包 **106.1 MB**、落地 367.4 MB；Avalonia 12.1 依赖 .NET 时 10.54 MB。来源：`https://dawloom.com/blog/tauri-vs-electron-vs-wpf-benchmark/`（第三方实测，贴了测量方法）。
3. **框架现状（2026-11）**：Wails **v3.0.0 已稳定**（2026-10-26 发布，v3.1.0 于 2026-11-01）——`https://wails.io/status/`；Tauri 2 官方体积说明 `https://v2.tauri.app/concept/size/`；Fyne 2.8 新增 `canvas.Shader`，已有 **fyne-mpv-video** 用 libmpv 软件渲染 API 在 Fyne 窗口内播视频，X11/Wayland/Windows/macOS 同一份代码（`https://github.com/uidbz/fyne-mpv-video`）。
4. **Chromium 的编码墙（实测）**：本机 Chromium 155 加载自造 mkv——H.264/AAC 正常（640×360、49 帧/1.5s、audioBytes 27,246）；**HEVC 0 帧黑屏**（videoWidth=0、totalVideoFrames=0、videoBytes=0，音频照放、不报错）；**AC3 音轨静音**（画面正常、audioBytes=0、不报错）；VP9/Opus、10-bit H.264 正常。`canPlayType` 对 hvc1/ac-3/ec-3/dtsc 全为空串。
5. **4 路同播与倍速（实测，Chromium 360p）**：4 路 5 秒各 154–155 帧、掉帧 0–10；`playbackRate` 0.5–16 全部生效（有效倍率 0.48/0.99/1.94/3.88/7.74/11.54/15），`preservesPitch=true`，8x 音频仍解码。1080p×4 **未测**。
6. **mpv 原生覆盖播放语义（文档）**：`.dsh-tmp/research/mpv-options.rst` → `--speed=<0.01-100>`（L262）、`--audio-pitch-correction` 默认开启（L265）、`--hr-seek`（L395）。
7. **Qt6 的格式与倍速支持（文档）**：Qt 6.12 `QMediaFormat::FileFormat` 含 `Matroska`、`VideoCodec` 含 `H265`；`QMediaPlayer` 有 `playbackRate`、`playbackOptions`、`PitchCompensationAvailability`（6.10 起）；Linux 走 GStreamer 需 `gstreamer1.0-libav`。
8. **目标视频编码已确认（用户输入）**：用户确认现有目标视频**全部是 H.264** [→ 用户原话 2026-11-20T16:16]。这条降低了编码风险，但不改变「仍要探测编码并提示」的要求——因为用户只看了现有素材。
9. **本机 QtMultimedia 播放测试未跑通（未核实）**：容器缺 `libEGL.so.1`，`import PySide6.QtMultimedia` 直接失败；因此在真机上跑 4 实例/8x/seek 的 spike 是必需的，不能拿文档当结论。

### 1.3 任务环境

- 仓库现状：greenfield。`main` 只有 `AGENTS.md` 与 `.docs/` 骨架，无源码、无依赖、无构建配置。
- 本任务 worktree：`/root/wt/tech-stack-research`（分支 `task/tech-stack-research`），起点 `a1a4cf0`。
- 目标平台：Windows 10/11 x86_64 + Linux x86_64（Win7 可选，非硬需求）。
- 本机实测环境：Ubuntu 24.04 aarch64 容器、8 核、headless、Chromium 155（Playwright）、PyAV 19.0.2、PySide6 6.12（缺 EGL 未能跑多媒体）；只作可行性证据，不等于目标机性能。
- 依赖约束：包管理只用 `uv` 与 `npm`；临时文件只放 `.dsh-tmp/`。
- **Go 工具链（已装，用户 2026-11-20T16:37 授权）**：`go1.27.4 linux/arm64`，解压在 `.dsh-tmp/goenv/go`（未动系统 PATH）；冒烟构建通过，产物 2,342,397 B。
- **仍缺（未授权安装）**：`libmpv` 头文件与库、`libEGL.so.1`、X11/GL 开发头（`xorg-dev`）。因此本容器**能 `go build` 纯 Go 包、能跑 `go test`，但不能编译 cgo 的 Fyne/libmpv，也开不了窗口**。apt 有 `libmpv-dev 0.37.0`、`libgl1-mesa-dev`、`xorg-dev`，装不装等用户另行授权。

### 1.4 执行步骤

1. **取样本定编码**：用户给 3–5 个真实 .mkv，`ffprobe` 出容器/视频/音频编码清单——验证点：清单落盘，明确是否存在 HEVC/AC3/DTS/10-bit。
2. **路线已由用户初定（2026-11-20T16:16）：Go + Fyne 2.8 + libmpv**（用户原话「目标视频都是 h264。考虑到兼容性我觉得首选那个很不错，确定初步就用它了」）——验证点：本条与 §1.7 文件所有权一致；若 spike 证明 4×1080p 软件渲染不可用，按「偏离申报」回用户重批，不自行换栈。
3. **外观定稿（已批准 2026-11-20T17:04）**：变体 A 落 `approved.png`，10 字段 `DESIGN.lock` 已写；工具/控件语义见 `.docs/design/tech-stack-research/UI-SPEC.md`——验证点：`check-alignment.sh --design-lock --design` exit 0。
4. **播放内核 spike**：所选栈上 4 路 1080p 同播、每路 0.5–8x、±10/30/60/120s、全局控制——验证点：4 路连续 60s 不断流；掉帧与 CPU 记录；8x 音频音调记录。
5. **布局与分割线**：6 种布局 + 拖动分割线 + 视频自适应——验证点：每种布局截图 + 拖动前后视频尺寸记录。
6. **分控/总控与历史记忆**：每路独立 + 全局叠加；多选文件；进度落盘与恢复——验证点：重启后恢复（或按降级路径重选，需用户确认）。
7. **打包与安装包门禁**：Windows 安装包 + Linux 包，按安装包体积实测——验证点：安装包 ≤50MiB，记录测量命令与产物路径。
8. **双平台验收**：按 §1.6 的 AC 在 Windows 与 Linux 各过一遍。

### 1.5 容易错忘漏的地方（≥3）

1. **体积口径会骗人**：安装后 116.2 MiB 的 mpv.exe 压进 7z 只有 32.6 MiB。按「安装后」算会把可行方案枪毙，按「安装包」算又可能放过一个解包后 300MB 的怪物。门禁必须量**安装包**，并在 §2 同时记录安装后占用。
2. **HEVC 不是报错，是黑屏**：实测 `videoWidth=0`、`totalVideoFrames=0`、`videoBytes=0`，音频照放、`error` 为 null。必须探测视频编码并明示「此文件 HEVC 无法播放」。
3. **AC3/EAC3/DTS 是静音，也不是报错**：实测画面正常、`audioBytes=0`；要探测音轨编码并提示「无声音」。
4. **QtMultimedia 在 Linux 依赖系统 GL/EGL 与 FFmpeg**：本机缺 `libEGL.so.1` 直接 import 失败；打包必须带依赖或声明，不能假设「pip 装了就能跑」。
5. **Fyne + libmpv 是软件渲染**：每帧过内存，4×1080p 的 CPU 开销没测过；demo 作者自己注明「4K/HDR 在弱硬件上可能不行」。spike 必须用 4 路 1080p 实测。
6. **1080p×4 未测**：本机只测 640×360×4；硬解实例数上限、软解 CPU 都可能在真实 1080p 上翻车。
7. **倍速音频依后端而异**：Chromium 8x 仍解码、`preservesPitch=true`；Qt/FFmpeg 与 mpv 的实现不同，必须在目标栈上实测音调与是否静音。
8. **浏览器/WebView 没有 hr-seek**：±10s 在无 index 的 mkv 上可能落到关键帧；mpv 有 `--hr-seek`（手册 L395）。
9. **文件句柄持久化只在 Chromium 系**：`showOpenFilePicker` + IndexedDB 是本机 Chromium 实测可用；其他引擎无此 API，历史记忆要有「重新选择文件」的降级路径。

### 1.6 验收标准 → 测试

| AC | 测试名 / 截图 |
|---|---|
| AC1 多选 4 个 mkv 同播 | `spike/four-up`：4 路 1080p 60s 不断流，记录每路 currentTime 与帧数 |
| AC2 6 种布局 + 拖分割线 | 截图 `.docs/evidence/tech-stack-research/layout-*.png` + 拖动前后视频尺寸记录 |
| AC3 每路 0.5～8x 独立倍速 | `spike/rate`：0.5/1/2/4/8 各 2s，实测 currentTime 倍率与音调 |
| AC4 每路 ±10/30/60/120s | `spike/seek`：记录 seek 后落点与目标误差 |
| AC5 全局暂停/跳转/叠加倍速 | `spike/global`：4 路同时暂停/跳转/倍率叠加后的一致性 |
| AC6 历史播放记忆 | `spike/history`：重启后恢复进度与文件（或记录降级行为） |
| AC7 安装包 ≤50MiB | `package/size`：Windows 安装包与 Linux 包实测，记录安装后占用 |
| AC8 Windows + Linux | 两平台冒烟记录（截图 + 版本） |
| AC9 快捷键 | `spike/keys`：空格暂停/恢复全部；←/→ 各全局 ±10s，4 路 currentTime 同步位移 ±10s |

### 1.7 范围

- 文件所有权（本阶段）：`.docs/investigations/tech-stack-research.md`、`.docs/tasks/tech-stack-research.md` §1、`.docs/design/tech-stack-research/{UI-SPEC.md,candidate-a-*.png,approved.*,DESIGN.lock}`、`.dsh-tmp/ff/**` 与 `.dsh-tmp/mockups/**`（实测脚本、样本与变体，过程证据，合并前清）。
- 文件所有权（实施阶段，Go + Fyne + libmpv）：
  - `go.mod` / `go.sum`：Go 模块与依赖（Fyne 2.8、go-mpv 或自写 cgo 绑定）
  - `main.go`：入口、窗口、生命周期
  - `internal/player/**`：libmpv 封装（4 实例、0.5–8x、±10/30/60/120s、音调、seek）
  - `internal/layout/**`：6 种布局、拖分割线、视频自适应
  - `internal/history/**`：历史播放记忆（落盘位置与格式）
  - `internal/ui/**`：工具栏、全局控制、每路 HUD、主题
  - `packaging/**`：Windows 安装包（NSIS/Inno）+ Linux 包（AppImage/deb）
  - `.docs/evidence/tech-stack-research/**`：截图与 spike 证据
- 禁止触碰：§0、`AGENTS.md`、`.docs/00-STATUS.md`（parent 维护）、`main` 分支。
- 外观锁：本阶段无；实施任务必须有 `approved.*` + `DESIGN.lock` 才能开工。
- 预计时长：调研 1 天（已完成）；实施 3～5 天，分期。

---

## §2 编写结果

> 进行中（2026-11-20T18:55）。已提交 `2d19c09`（骨架），后续继续追加。

### 2.1 改动清单

| 文件:行 | 为什么 |
|---|---|
| `go.mod` / `go.sum` | Go 模块；Fyne v2.8.2；`-tags mpv` 时链接 libmpv |
| `cmd/mvplayer/main.go` | 入口：主题、窗口 1280×800、历史文件路径 |
| `internal/layout/layout.go` | 6 种布局的矩形计算、分割线命中/拖拽、比例钳制 |
| `internal/history/history.go` | 按「路径+大小+mtime」记进度，原子落盘、裁剪 |
| `internal/player/player.go` | Player 接口 + 全局控制语义（全部暂停/跳转/倍速叠加、0.5–8 钳制） |
| `internal/player/engine_mpv.go` | libmpv cgo 实现（`-tags mpv`）：loadfile/speed/seek/pause/time-pos |
| `internal/player/engine_stub.go` | 无 mpv 标签时的内存桩，保证 UI 可编译可测 |
| `internal/ui/theme.go` | 批准件配色（#17181A/#1A1C1F/#1E2023/#2A2D31/#D7D9DD/#8A8F98/#5B8CFF） |
| `internal/ui/pane.go` | 单路画格：文件名、倍速徽标、HUD（时间/进度/档位/±/播放/倍速/移除） |
| `internal/ui/stage.go` | 画格排布、5px 分割线拖动、布局切换 |
| `internal/ui/app.go` | 工具栏（打开/历史/7 布局/全局控制/倍速/叠加/设置）、开文件、历史落盘、200ms 刷新 |
| `internal/ui/ui_test.go` | 布局切换/按钮启用/移除重排/分割线拖动 + 无头截图 |
| `cmd/spike/main.go` | 4 路 libmpv 实测（倍速、seek、CPU） |

### 2.2 新增/修改的测试及其守住的行为

| 测试 | 守住什么 | 变异证据（改坏后红在哪条） |
|---|---|---|
| `TestComputeFillsAreaWithoutOverlap` | 6 种布局铺满、不重叠、空位（3 路四宫格）只留一格 | 去掉 gap 计算 → 面积断言红 |
| `TestHitTest` / `TestDragClamps` | 分割线命中、拖到边界钳在 10%/90% | 去掉 clamp → `TestDragClamps` 红 |
| `TestSeekAllClampsPerPane` | 全局 ± 对每路独立钳制（含未知时长） | 去掉 duration 钳制 → 短片区断言红 |
| `TestToggleAllPausesWhenAnyPlaying` | 任一路在播 → 全停；再按 → 全播 | 改成「全都在播才停」→ 红 |
| `TestSetBaseAllAndNudgeClamp` | 全局基线与叠加倍速钳在 0.5–8 | 去掉 ClampRate → 红 |
| `TestKeyIncludesSizeAndMtime` | 文件被替换后不复用旧进度 | 去掉 size/mtime → 红 |
| `TestAddFourFilesSwitchesToQuadLayout` | 开满 4 路自动四宫格 | 不调用 relayoutForCount → 红 |
| `TestLayoutButtonsFollowPaneCount` | 布局按钮按路数启用/禁用 | 去掉 refreshLayoutButtons → 红 |
| `TestRemovePaneRelayouts` | 移除一路后自动换到 3 路布局 | 不重排 → 红 |
| `TestSplitterDragChangesRatio` | 拖动改变比例且不越界 | 去掉 Drag → 红 |

### 2.3 门禁输出

```
$ go test ./...
ok  mvplayer/internal/history
ok  mvplayer/internal/layout
ok  mvplayer/internal/player
ok  mvplayer/internal/ui
exit 0
```

```
$ go build -tags mpv -o build/mvplayer_mpv ./cmd/mvplayer
-rwxr-xr-x 30M build/mvplayer_mpv
exit 0
```

4 路 libmpv spike（`./build/spike -file long_h264_aac.mkv -n 4`，640×360）：

```
playing pane=0..3 pos=5500/5500/5500/5467ms rate=1.00 paused=false
rate pane=0 set=0.50 effective=0.56
rate pane=1 set=1.00 effective=1.00
rate pane=2 set=2.00 effective=1.88
rate pane=3 set=4.00 effective=3.62
seeked pane=0..3 pos=10433/18833/21866/28067ms（目标 +10s，误差 ≤300ms）
cpu avg=105.1% wall=9.5s panes=4
```

无头 UI 截图：`.docs/evidence/tech-stack-research/impl-4up.png`（2 路/3 路同目录）。

### 2.4 设计一致性（有设计稿时）

| 项 | 设计稿 | 实现 | 状态 |
|---|---|---|---|
| 配色 | #17181A/#1A1C1F/#1E2023/#2A2D31/#D7D9DD/#8A8F98/#5B8CFF | `theme.go` 同值（测试也套用同一主题） | ✅ |
| 工具栏分组 | 打开/历史/7 布局 · 全局 · 倍速/叠加/设置 | 左/中/右三组，6px 内边距，1px 分组线 | ✅（分组线偏细） |
| 控件形态 | 1px 描边 chip | `chip.go`：1px stroke + 4px 圆角 + 悬停/选中强调底 | ✅ |
| 每路 HUD | 时间/进度/档位/±/播放/倍速/移除，单行 | 单行 Border：时间、进度、档位、−、Ⅱ、＋、倍速、✕ | ✅ |
| 分割线 5px | 5px 拖动区 | `layout.Gap=5`，10%–90% 钳制 | ✅ |
| 布局图标 | 7 个矢量图标 | 暂用字符「1 ∥ ＝ ◧ ◨ ▣ ▦」 | ⚠️ 待换成矢量 |
| 底部状态 | 基线列表 + 叠加后列表 + 快捷键 | 「N 路 · 状态 · 叠加 ±」+ 快捷键 | ⚠️ 简化 |

并排图：`.docs/evidence/tech-stack-research/compare.png`（左批准件，右实现）。

### 2.5 自审（T2/T3 七条）

1. 变异：见 2.2 表（每条都写了改坏后红在哪条）。
2. 接线：`relayoutForCount` 一行删掉 → `TestAddFourFilesSwitchesToQuadLayout` 红。
3. 用例名 vs 断言：测试名都写行为，断言看值（如 pos=1000 钳制）。
4. 注释 vs 事实：`engine_stub.go` 注释写明「无 mpv 标签时的桩」，`go build` 两种标签都验证。
5. 静默路径：打开失败弹 dialog；满 4 路提示；编码不支持后续要探测。
6. 最薄弱的一处：mpv 画面尚未渲染进画格（当前只有音视频解码与 HUD），这是下一步。
7. 边界：暂无 reach-in；未重构；无格式三件。

### 2.6 顺带发现（不在本支修）

- 容器里 `gcc` 是 termux 的 clang（Android sysroot），cgo 必须显式 `CC=/usr/bin/gcc-13`，否则 `net` 与 glfw 都编不过；已装 `gcc-13`。
- Fyne 在 Xvfb + llvmpipe 下 SIGILL（GL 栈崩溃），GUI 只能靠 `fyne.io/fyne/v2/test` 的软件画布验证；真机需要用户跑。
- `go.dev`/`proxy.golang.org` 不可达，模块走 `GOPROXY=https://goproxy.cn,direct`。

### 2.8 追加：渲染、交互、设置与打包（2026-11-20T19:20）

**新增能力**
| 能力 | 实现 | 证据 |
|---|---|---|
| mpv 画面进画格 | `engine_mpv.go` 软件渲染 API（`rgb0` → `image.RGBA`，alpha 补 255，双缓冲） | `TestRenderFrameProducesPicture`：320×180 帧方差 1087.5、alpha=255；`frame-check.png` |
| 4 路画面绘制 | `app.renderLoop`（33ms）+ `Pane.SetFrame` | `TestFramesPaintIntoPanes` → `impl-frames.png`（四格四色渐变） |
| 一次多选文件 | `picker.go` 自建目录浏览器（目录/视频过滤、☐/☑、上一级） | `TestMultiPickerListsVideos` |
| 快捷键 | 空格=全局暂停，←/→=全局 ±10s，1–7=切布局 | `TestShortcutSpacePausesAll`、`TestShortcutArrowsSeekTenSeconds` |
| 历史恢复 | 打开文件命中历史 → `pendingSeek` 待时长就绪后 seek | `TestHistoryRestoresPosition`（恢复到 120s） |
| 基线 + 叠加倍速 | `baseRate` 每路 + 全局 `overlay`，有效值 = clamp(base+overlay, 0.5, 8) | 状态栏「基线 … · 叠加后 …」 |
| 设置页 | 倍速列表、档位列表、历史开关，落 `settings.json` | `settings.go`，写入 config 目录 |
| 矢量布局图标 | `icons.go` 按布局画 1–4 个矩形 | `impl-frames.png` 工具栏 |
| 打包脚手架 | `Makefile`、`packaging/windows/installer.nsi`、`packaging/linux/*`、`README.md` | 见 §2.9 体积 |

**新增测试**（全部通过）
`TestRenderFrameProducesPicture`、`TestFramesPaintIntoPanes`、`TestShortcutSpacePausesAll`、
`TestShortcutArrowsSeekTenSeconds`、`TestMultiPickerListsVideos`、`TestHistoryRestoresPosition`。

**门禁**
```
$ go build -tags mpv -o build/mvplayer_mpv ./cmd/mvplayer   → 30M, exit 0
$ go test ./...                                            → 4 包全绿, exit 0
```

### 2.9 体积（实测 + 推算，待 Windows 实机复核）

| 组成 | 实测/推算 | 依据 |
|---|---|---|
| Linux 二进制（Fyne + mpv） | **30 MiB** | `build/mvplayer_mpv` |
| Windows exe（Fyne，无 mpv） | **25 MiB**；xz -9 后 **9.2 MiB** | `x86_64-w64-mingw32-gcc` 交叉编译（`-H windowsgui -s -w`）实测 |
| Windows exe（链接 libmpv 后增量） | 忽略不计（只多导入表） | 推断 |
| libmpv-2.dll（Windows，20261108 构建） | **116 MiB 未压缩 / 34 MiB xz -6** | SourceForge `libmpv/mpv-dev-x86_64-20261108`，实测 |
| NSIS 安装包 | **42.5 MiB（44,588,554 B，实测）** | `makensis` 3.09 实打包：安装后 147,304,968 B，压缩率 30.2%，≤50MiB（余量 ≈7.5 MiB） |

未完成：Windows 交叉编译（容器缺 `libmpv.dll.a`，GitHub 下载被截断）；真机运行验证（容器 GL 栈 SIGFAIL）。

### 2.10 1080p×4 实测与绑定修复（2026-11-20T19:35）

**1080p×4（1920×1080 H.264/AAC，10s 样本，4 实例同播）**
```
load pane=0..3 dur=10021ms pos≈1267–1432ms
playing pane=0..3 pos=5433/5367/5333/5266ms rate=1.00
render 640x360 frames=60 avg=0.83ms/frame
cpu avg=90.9% wall=9.4s panes=4
```
结论：4 路 1080p 解码 + 每帧软件渲染在本机（8 核 ARM）平均不到 1 个核；渲染单帧 0.83ms（640×360）。
注意：10s 片长在 2x/4x 时提前 EOF，倍速有效性以 640×360/30s 样本的实测为准（0.56/1.00/1.88/3.62）。

**修复：画格↔播放器绑定**
- 原实现用「画格下标 = players 下标」，移除中间一路后错位。改为每个 `Pane` 持有自己的 `player.Player`，所有控制/渲染/落盘都走 `p.pl`；`Controller` 仍按玩家列表维护。
- 新增 `TestRemoveMiddlePaneKeepsOthersWired`：移除中间一路后，另两路位置不动、nudge 只作用于目标；空出的画格可重新装入并只绑定该格。

**mpv 选项**：显式 `vo=libmpv`（渲染 API 需要）、`audio-pitch-correction=yes`（0.5–8x 保持音调）、`hwdec=auto-safe`。

### 2.11 端到端与空位（2026-11-20T19:55）

- **真 libmpv 端到端**：`TestRealMpvFramesIntoPanes`（`-tags mpv`）解码 1080p → 软件渲染 → 画格，截图 `impl-realmpv.png`（两路 1080p，letterbox 正确）。
- **空位**：`layout.EmptyCell` 给出 3 路四宫格空格的矩形；`emptySlot` 显示「空位 · 拖入视频」；0 路时整块舞台一个空位。截图 `impl-3q-empty-slot.png`、`impl-empty.png`。
- **Windows exe 实测**：25 MiB（xz 9.2 MiB，Fyne，无 mpv）；NSIS 安装包估算 37–40 MiB（含 libmpv LZMA ≈28–30 MiB），≤50MiB。
- 仍未闭环：Windows 版 libmpv 包下载被 GitHub 反复截断（11.8/30.3 MiB），拿不到 `libmpv.dll.a`/`libmpv-2.dll`；真窗口需用户机器验证（容器 GL 栈 SIGILL）。

### 2.7 收尾证据

```
$ ps -eo pid,args | grep -E 'mvplayer|xvfb'
（无残留进程）
```

---

## §3 审查意见

<!-- 审核者写。逐条编号，四要素：位置 / repro / 期望 / 实际（+影响）。 -->

### 3.1 第 1 轮

```
F1(P0) <file:line> | repro: <命令/测试> | 期望: <…> | 实际: <…> | 影响: <用户会怎样>
```

没有 finding 就写「无」。判定：<可合并 / 需修>（不要预填一个结论；预填的 finding 和结论会被当成真的读）

---

## §4 修复记录

<!-- owner 写。对每条 finding 的回应。 -->

| F | 改了什么 | 钉住它的测试 | 变异证据 |
|---|---|---|---|

---

## §5 收尾

- 闸口记录：G1 通过（22 条需求可追溯 + 设计锁 10 字段一致）；G2 完成（§2/§2.8–2.11）；G3 审核**按用户指示跳过**（用户 2026-11-20 明确「禁止你再派子代理」，无法派审核者，已记录）；G5 合并由 owner 按用户直接指示执行（用户 2026-11-20T20:38「先把项目收尾合并上去在 main 跑通」）。
- 合并：`04264e4`（merge: task/tech-stack-research）+ `31ff86c`（工具栏单行对齐修复）。合并前在克隆中解决 `.gitignore` add/add 冲突（取并集：`.dsh-tmp/`、`build/`、`.docs/evidence/`、`*.log`）。
- 合并后门禁（main 上实测）：
  - `go vet ./...` / `go vet -tags mpv ./...` → exit 0
  - `go test ./...` → 4 包全绿；`go test -tags mpv ./...` → 4 包全绿（含真 libmpv 端到端）
  - `go build -tags mpv` → 30 MiB；Windows 交叉编译（mingw + libmpv）→ 25 MiB exe
  - Windows NSIS 安装包 → 42.5 MiB（44,588,554 B），≤50MiB
- 推送：曾推 Gitee（`origin`），后按用户要求改回 **GitHub**：`github = https://github.com/raochongqi/mvplayer.git`，用用户提供的 PAT 推送 main；Gitee Go 的 `.workflow/` 配置已移除。
- CI 现状：`.github/workflows/ci.yml`（GitHub Actions：Linux vet + 两套测试 + libmpv 构建；Windows mingw 交叉编译）。`packaging/ci/Dockerfile` 保留为本地/容器化 CI 用（本机无 docker，未实跑）。Gitee Go 方案已按用户要求放弃并移除。
- 清理：`.dsh-tmp/`（样本/脚本/克隆）与 `.docs/evidence/` 不进版本库（已 gitignore）；`build/` 同样忽略。worktree `/root/wt/tech-stack-research` 与分支 `task/tech-stack-research` 可删（内容已合入 main）。
- 仓库只留文本与设计批准件：过程截图、变体图、测试用 mkv 均未提交；应用图标改为运行时用 `image/png` 生成（`internal/ui/icon.go`）；真实解码测试改为 `MVPLAYER_SAMPLE` 显式开启，CI 用 ffmpeg 现场生成样本。
- 遗留：真机窗口验证、长片历史手感、Windows 实机安装；CI 首次运行需用户 push。

### 交接块（parent 换人/上下文耗尽时接这五行）

```
停在：<G2 第 3 步 / G3 待复审>
已批：<设计锁 hash 与批准时间>
未做：<下一步动作，一条命令>
勿动：<本任务禁区>
文档：.docs/tasks/tech-stack-research.md §0–§5
```
