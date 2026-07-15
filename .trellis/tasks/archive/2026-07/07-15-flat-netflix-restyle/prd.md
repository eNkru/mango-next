# Restyle Flat UI toward Netflix aesthetic

## Goal

在 **不改动 Comic 风格** 的前提下，把 **Flat UI** 做成接近 Netflix 浏览体验的全站视觉与布局：顶栏 chrome、全宽海报/内容区、深/浅双主题的流媒体语言、Netflix 红主强调色。

## User value

Flat 从「通用后台扁平」升级为「本地漫画库的流媒体浏览感」：封面更大、层级更清晰、全站一致（含 Admin/Reader 壳），同时 Comic 仍可一切换回漫画风。

## Confirmed facts

1. UI 风格仅 `comic` | `flat`（`common.js` / admin）。
2. Flat 去掉 `comic-theme*`；暗色现 `#121212`。
3. 色板在 `_variables.less`；Flat 现珊瑚 `#D96A4B`。
4. 全局壳为 **侧栏**（`top.tmpl`）+ 主区；首页已有 hero/carousel 雏形。
5. 模板多 `comic-*` class；Flat 下样式多不生效。
6. Netflix 参考：近黑/高对比、红 CTA、海报行、hover——不克隆商标与营销落地。

## Product decisions

| Decision | Choice |
|----------|--------|
| 目标 | Flat only → Netflix-inspired browse |
| 还原度 | **C**：顶栏化、全宽、模板/布局大改（不仅色板） |
| 页面范围 | **尽量全站 Flat**（Home/Library/Title/Tags/Login/Admin/Reader 壳/插件下载等） |
| 主题 | **Dark + Light 均 Netflix 化** |
| 强调色 | **Netflix 红**（约 `#E50914` 系） |
| Comic | **零回归**：侧栏与 comic 视觉保持 |
| 分支 | `ui/restyling` |
| 商标 | 不用 Netflix logo/素材 |

## Requirements

- **R1 Layout (Flat only)**：桌面主导航为 **顶栏**（logo + 主导航 + 工具入口）；移动端保留可折叠导航；**不得**在 Flat 下再依赖侧栏为主导航。Comic 继续侧栏。
- **R2 Canvas**：Dark Flat 近黑画布 + 白/灰字；Light Flat 为 Netflix 风浅色浏览（浅灰底、深字、同结构），非旧珊瑚后台。
- **R3 Accent**：Primary CTA/进度/焦点用 Netflix 红系；hover 态配套。
- **R4 Content surfaces**：书卡/条目标题海报化（比例、圆角克制、hover 放大或抬升、标题叠层渐变优先）。
- **R5 Home**：Continue / 新书等 section 呈横向 rails；可选顶部 featured 区更强（在现有数据能力内）。
- **R6 Library / Title / Tags**：网格与详情页与 rails 语言一致；操作按钮克制。
- **R7 Login / Admin / Reader 壳 / 插件相关**：同一 token 与顶栏/chrome 语言；功能不删减。
- **R8 Toggle**：Admin + FAB 切换 comic/flat 立即生效；刷新后保持 localStorage。
- **R9 Build**：样式源优先 less（`_variables.less` / 新 flat 层 / mango.less），提交编译后的 css。
- **R10 A11y**：对比度可用；焦点可见；不单靠颜色传达状态。

## Acceptance Criteria

- [x] Flat + Dark：全站主浏览页观感接近 Netflix browse（顶栏、近黑、红 CTA、海报卡）
- [x] Flat + Light：同布局语言的浅色 Netflix 化（非旧 Flat 珊瑚后台）
- [x] Comic + 任意 theme：侧栏与 comic 视觉与改前一致（CSS 全 scoped 于 flat-theme）
- [x] 切换 ui-style / theme：flat-theme / comic-theme class 互斥清理
- [x] Home / Library / Title / Login / Admin / Reader Flat 样式已覆盖主路径
- [x] 无 Netflix 商标或官方素材

## Out of scope

- Comic 改版
- Netflix 商标/官方图/营销文案
- 新后端 API 或推荐算法
- 阅读器翻页/图片管线性能（壳样式可以改）
- 第三套 UI style 命名（除非后续单独决定把 Flat 改名）

## Delivery

分 **3 批 PR**（设计一次、实现分批）：

1. 全局壳 + token + 顶栏  
2. Home / Library / Title / Tags  
3. Login / Admin / Reader 壳 / 插件相关  

Reader：壳与控件 Netflix 化，**阅读模式/图片管线行为不变**。

## Open questions

- 无阻塞产品问题；实现细节见 `design.md`。

## Notes

- Research: `research/netflix-vs-flat.md`
- Design: `design.md` · Implement: `implement.md`

