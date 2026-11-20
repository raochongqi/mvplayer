# 状态

## 任务队列

| 任务 | 级别 | 状态 | 文档 |
|---|---|---|---|
| tech-stack-research | T3 | 已完成并合入 main（`86d053e`）；待推送 GitHub + 真机验证 | `.docs/tasks/tech-stack-research.md` |

## 项目概况

- 目标：多 mkv 播放小工具（1～4 路、多布局、独立倍速/跳转、全局控制、历史记忆、可拖分割线）。
- 路线（已定）：Go 1.27 + Fyne 2.8 + libmpv（cgo；容器内需 `CC=/usr/bin/gcc-13`）。
- 实测门槛：4×1080p 同播 CPU 均值 0.91 核；Linux 二进制 30 MiB；Windows exe 25 MiB；**NSIS 安装包 42.5 MiB（≤50MiB）**。
- 仓库：`main`，49 个受控文件。远端 `origin` = https://github.com/raochongqi/mvplayer.git。
- **阻塞**：`~/.git-credentials` 只有 gitee 凭据，`git push origin main` 失败，需用户提供 GitHub token 或自行 push。
- 产物：`build/mvPlayer-0.1.0-setup.exe`（42.5MiB，gitignore）。

## 待用户机器验证

1. 真窗口运行与点按（本容器 GL 栈 SIGILL，只能软件画布截图）。
2. Windows 安装包实机安装、拖分割线、4 路 1080p 长播。
3. 长片历史恢复手感。

## 长期指令 / 经验

- 包管理只用 `uv` 与 `npm`；临时文件只放 `.dsh-tmp`。
- `.docs` 是唯一文档目录；parent 不读源码、不写需求、不定级。
- 容器 `uptime`/`loadavg` 来自 proot 绑定的 sysdata，不可信；测 CPU 用逐进程 `/proc/<pid>/stat` 采样。
- cgo 必须显式 `CC=/usr/bin/gcc-13`（容器默认 gcc 是 termux clang + Android sysroot）。
