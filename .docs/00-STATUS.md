# 状态

## 任务队列

| 任务 | 级别 | 状态 | 文档 |
|---|---|---|---|
| tech-stack-research | 待 owner 定级 | 规划中（owner 调研 + 写任务书） | `.docs/tasks/tech-stack-research.md` |

## 项目概况

- 目标：多 mkv 同步播放小工具（1～4 路、联动进度、多布局、独立倍速/跳转、全局控制、历史记忆、可拖分割线）。立项阶段，技术栈未定。
- 候选：纯 web/HTML、Go、Python；优先级：原生跨平台 > WebView。
- 仓库：`main` 分支，尚无源码。

## 长期指令 / 经验

- 包管理只用 `uv` 与 `npm`；临时文件只放 `.dsh-tmp`。
- `.docs` 是唯一文档目录；parent 不读源码、不写需求、不定级。
