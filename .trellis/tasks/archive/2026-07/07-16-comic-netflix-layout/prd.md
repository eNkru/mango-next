# Restyle Comic UI layout like Flat Netflix

## Goal

在 **保留 Comic 漫画视觉气质**（半调网点、纸感、粗边框、漫画字体、高饱和原色板）的前提下，把 **Comic 的布局与信息架构** 对齐已完成的 Flat Netflix 化：顶栏 chrome、全宽内容区、海报化书卡、Home rails、全站壳一致。

## User value

切换到 Comic 时不再是「旧侧栏后台 + 漫画皮肤」，而是「与 Flat 同级的流媒体浏览布局 + 中等强度漫画装饰」：结构熟悉、封面更大、层级清晰，同时仍可识别 Kirby/Stan Lee 式漫画味道。

## Confirmed facts

1. UI 风格仅 `comic` | `flat`（`common.js` / admin / localStorage `ui-style`）。
2. Flat 已 Netflix 化：`flat-theme.less` 顶栏、全宽；`top.tmpl` 含 `flat-topbar` + comic 侧栏双壳。
3. 当前 CSS：`body:not(.flat-theme) .flat-topbar { display: none }` — 顶栏仅 flat 可见；comic 仍侧栏。
4. Comic 样式集中于 `comic-theme.less`（~1200 行），含侧栏、半调、卡片 pop 等。
5. 前序归档：`archive/2026-07/07-15-flat-netflix-restyle`。

## Product decisions

| Decision | Choice |
|----------|--------|
| 主导航 | **顶栏（同 Flat）**；侧栏不再作 comic 主导航 |
| 页面范围 | **全站同 Flat** |
| 色板/强调色 | **全保留 comic 原色**（不强制 Netflix 红） |
| Dark + Light | **双主题均改造** |
| 装饰强度 | **B 中等**：半调/纸感 + 粗边 + 漫画字体；减弱大面积 pop/倾斜动画 |
| Flat 回归 | **必须零回归** |

## Requirements

- **R1 Layout**：桌面顶栏；移动端可折叠与 Flat 同级；侧栏 hidden under comic（与 flat 一致策略）。
- **R2 Feel**：半调/纸感/粗描边/漫画字体保留；非第二套 Flat 换色。
- **R3 Palette**：CTA/边框/装饰用 comic 原色板。
- **R4 Themes**：Dark + Light 均顶栏 + 全宽 + rails 结构。
- **R5 Decor**：中等强度 — 保留背景纹理与边框语言；卡片 hover 克制（轻微抬升/阴影），避免强 rotate/pop 打断 rails 浏览。
- **R6 Scope**：Home / Library / Title / Tags / Login / Admin / Reader 壳 / 插件相关。
- **R7 Surfaces**：海报卡与 rails 结构对齐 Flat；装饰仍 comic。
- **R8 Toggle**：切换立即生效；刷新保持；class 互斥。
- **R9 Isolation**：新规则 scoped `comic-theme` / `comic-theme-dark`；不改 flat 行为。
- **R10 Build**：less 源 + 提交编译 css。
- **R11 A11y**：对比度、焦点可见。

## Acceptance Criteria

- [x] Comic Dark + Light：顶栏 + 全宽主区；侧栏不作主导航
- [x] Comic 仍可识别为漫画风（原色板 + 半调/纸感/粗边/字体）
- [x] 装饰中等：无强 pop/倾斜主导的卡片进场；rails/网格可读
- [x] Flat + 任意 theme 无回归（仅顶栏可见性选择器共享）
- [x] 切换 ui-style / theme 稳定
- [x] Home / Library / Title / Login / Admin / Reader 主路径已覆盖（壳几何 + 既有 comic 组件）
- [x] 无 Netflix 商标/官方素材
- [x] Comic 全站直角（border-radius: 0）

## Out of scope

- 新后端 API / 推荐算法
- Netflix 商标/官方素材
- 阅读器翻页与图片管线行为变更（壳样式可改）
- 第三套 UI style 命名
- 保留 comic 侧栏作为可选布局
- 将 comic 主色改为 Netflix `#E50914`
- Flat 视觉再改版

## Delivery

建议 **3 批**（对齐 Flat 批次）：

1. 全局壳 + 顶栏可见性 + 内容区 padding + 装饰强度基线  
2. Home / Library / Title / Tags  
3. Login / Admin / Reader 壳 / 插件相关  

## Notes

- 前序：`07-15-flat-netflix-restyle`
- 复杂度：需 `design.md` + `implement.md`
