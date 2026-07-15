# Design: Comic layout aligned to Flat Netflix shell

## Intent

Comic 采用与 Flat **相同的布局/chrome 结构**（顶栏、全宽、海报卡、rails），**视觉 token 与装饰** 仍走 comic 原色板与中等漫画气质。Flat 路径零回归。

## Dual layout (after change)

| UI style | Chrome | CSS markers |
|----------|--------|-------------|
| `flat` | 顶栏 + 全宽 | `flat-theme` / `flat-theme-dark` |
| `comic` | **同顶栏 + 全宽**（不再侧栏） | `comic-theme` / `comic-theme-dark` |

侧栏 DOM（`top.tmpl` `.app-sidebar`）可保留但 **comic 与 flat 均隐藏**；或仅 flat 隐藏、comic 也 `display: none`。顶栏 DOM 现有 `.flat-topbar`：

- **现状**：`body:not(.flat-theme) .flat-topbar { display: none }` → comic 看不到顶栏
- **目标**：comic 也显示顶栏；class 可仍叫 `flat-topbar`（共享结构）或重命名为中性 `app-topbar`（可选，非必须）

推荐：**共享 `.flat-topbar` 结构**，用 CSS 分别 skin：

```
body.flat-theme .flat-topbar { … Netflix 皮 … }
body.comic-theme .flat-topbar,
body.comic-theme-dark .flat-topbar { … comic 皮（粗边/原色/字体）… }
body.comic-theme .app-sidebar,
body.comic-theme-dark .app-sidebar { display: none !important; }
```

并 **删除或改写** `body:not(.flat-theme) .flat-topbar { display: none }`，改为仅在无 theme 时隐藏，或显式：

```
body.flat-theme .flat-topbar,
body.comic-theme .flat-topbar,
body.comic-theme-dark .flat-topbar { display: flex; }
```

## Feel vs structure

| Layer | Source of truth |
|-------|-----------------|
| Shell geometry (height, padding, full-width content) | Mirror Flat (`@flat-topbar-height`, `4vw` gutters, content top padding) |
| Color / border / texture / type | Comic tokens (`_variables.less` comic section) |
| Hover / motion | **Medium**：scale/shadow 轻微；关闭或大幅减弱 `comic-card-pop` 强 rotate 进场 |
| Primary CTA | Comic 原色（如 `@comic-red` / 现有 comic 按钮色），非 `@accent` Netflix 红 |

## Component language (Comic)

1. **Top bar**：同 Flat 信息架构；背景可用半透明纸感或 ink 条 + 底边粗描边；logo 可用原 comic icon 或 mark + Bangers/Fredoka 字重。
2. **Poster card**：2:3 比例与网格密度对齐 Flat；边框更粗、角标/半调可选；hover 轻微抬升。
3. **Rails / sections**：横向滚动或 flex row 同 Home Flat；section 标题用 comic heading 字体，字号层级对齐。
4. **Buttons**：粗边 ink + comic fill；flat 的 ghost/red 逻辑不照搬。
5. **Forms / Admin / Login / Reader shell**：同结构密度；输入框与表用 comic 边框语言。
6. **FAB**：comic 既有 speed-dial 风格可保留，位置不挡顶栏。

## CSS architecture

```
_variables.less      # 不改 flat Netflix tokens；comic tokens 可微调但不换主色哲学
flat-theme.less      # 调整顶栏可见性规则：不再「仅 flat」；其余 flat 规则不动
comic-theme.less     # 主战场：顶栏 skin、hide sidebar、content padding、减弱 pop、页面 surfaces
mango.less           # 仅当共享结构缺类时最小改动
```

**Scope rule**：comic 新规则仅在 `html/body.comic-theme` / `comic-theme-dark` 下。  
**Flat isolation**：除「顶栏可见性共享」必要改动外，不改 flat 视觉 token 与页面皮肤。

## JS / templates

- `common.js`：class 切换逻辑大体不变；确认 comic 不再依赖侧栏宽度/class。
- `top.tmpl`：优先 CSS-only；若 logo 源需 comic 专用，可加 class 分支或 CSS `content`/`background`。
- `toggleSidebar`：侧栏隐藏后可成为 no-op（保留函数无害）。

## Page mapping

| Page | Treatment |
|------|-----------|
| Global shell | 顶栏 + full-width `app-content`；无侧栏 margin |
| Home | Featured + rails，comic 装饰中等 |
| Library / Tag / Title | 海报网格 + 详情结构对齐 Flat |
| Login / Admin / Reader shell / plugins | 同壳 + comic 表单/卡片语言 |

## Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| 共享 `.flat-topbar` DOM | 无重复导航 HTML | 命名略 flat；需 comic skin 覆盖 |
| 重命名 `app-topbar` | 语义中性 | 触碰 flat 选择器，回归面更大 |
| 中等装饰 | 浏览可读 + 仍 comic | 老 comic 重度动画粉丝可能觉得「淡」了 |

**推荐**：共享 DOM + comic skin；不重命名除非实现中痛点明显。

## Rollback

- 恢复 `body:not(.flat-theme) .flat-topbar { display: none }`
- 恢复 comic 侧栏 `display` 与 `app-content` margin
- 还原 comic 卡片动画

## Delivery batches

| Batch | Deliverable |
|-------|-------------|
| **1 Shell** | 顶栏对 comic 可见；侧栏隐藏；content padding；顶栏 comic skin；减弱全局 pop 基线 |
| **2 Browse** | Home / Library / Title / Tags 海报与 rails |
| **3 Rest** | Login / Admin / Reader 壳 / 插件相关 |
