# UI Consistency Audit and Fix

## Goal

全面修复前端 UI 不一致问题，使所有页面遵循统一的设计系统、组件复用和 i18n 规范。

## Confirmed Facts (from codebase audit)

- 技术栈：React + Vite，全局 CSS（tokens.css + shell.css），无 CSS modules/Tailwind
- 主题系统：comic/flat × light/dark，通过 html class 切换，tokens.css 定义设计令牌
- 共享组件：AppShell, PosterCard, BrowseToolbar, ProgressBar, StatePanels, ConfirmDialog, AlertHost
- 路由：boot-driven（App.tsx switch on boot.pageId）
- `PosterCard` / `BrowseTitle` 需要：`id, name, display_name, file_name, sort_name, cover_url, entry_count, progress, modified_at, hidden`
- `GET /api/tags/{tag}/titles` **当前只返回**：`id, name, cover_url, entry_count, hidden`（无 progress / sort_name / modified_at 等）
- 已接入 `useI18n` 的页面：Home, Library, TitleDetail, Admin, Reader（部分）
- **未**接入 i18n 的页面：Login, TagsIndex, TagDetail, UserList, UserEdit, MissingItems, Placeholder
- `ErrorState` 目前仅 `message`，无 `onRetry`；Home 在组件旁另放 retry 按钮
- 表单标记不一致：`label.mango-field > span + control`（TitleDetail/Reader）vs `div.mango-field > label + input`（Login/UserEdit）
- Out of scope：后端 API 变更（本任务不扩展 tag titles 字段）

## Requirements

### R1: TagDetail 复用共享组件（前端适配，不改 API）

**决策（2026-07-21）：路径 A — 前端适配**

- TagDetailPage 使用 `PosterCard` 替代手写卡片 markup
  - **`showProgress?: boolean`（2026-07-21 决策 A）**：默认 `true`；TagDetail 传 `false`，避免全 0 空进度条
- 将 API 卡片映射为 `BrowseTitle` 兼容对象：缺失字段用安全默认
  - `progress: 0`
  - `display_name` / `file_name` / `sort_name` ← `name`
  - `modified_at: 0`
  - 其余字段透传
- 使用 `BrowseToolbar` 替代手写搜索/排序控件
- **BrowseToolbar 扩展（2026-07-21 决策 B）**：新增可选 `modes?: SortMode[]`（默认全量四种）
  - TagDetail 仅传入 `['natural', 'title']`，不展示 `modified` / `progress`
  - Library / TitleDetail 不传 `modes`，行为与现网一致
  - 若当前 `mode` 不在允许列表中，回退到列表首项（通常 `natural`）
- 移除本地 `TitleCard` 类型与重复卡片逻辑
- Admin hide/show 操作继续通过 `actions` 插槽挂到 `PosterCard`
- **不**扩展 `apiTagTitles` 响应字段

### R2: i18n 完整化
- Login, TagsIndex, TagDetail, UserList, UserEdit, MissingItems 页面接入 useI18n
- 所有硬编码中文/英文字符串替换为 t() 调用
- i18n 字典补全缺失的 key（zh-cn / zh-tw / en 三语同步）
- StatePanels 默认文案改为走 i18n 或强制由调用方传入已翻译 message（避免组件内硬编码中文默认值）
- **Login 语言切换（2026-07-21 决策 A）**
  - Login 接入 `useI18n` / `t()`
  - 在登录卡片 header 或 footer 提供语言 `<select>`（与 AppShell 同 keys：zh-cn / zh-tw / en）
  - 可抽轻量 `LanguageSelect` 供 AppShell 与 Login 复用，或 Login 内联同等逻辑；行为写 `localStorage['mango-language']` 与现网一致
  - Login **不**引入完整 AppShell topbar

### R3: Loading/Empty/Error 状态统一
- 所有页面使用共享 LoadingState/EmptyState/ErrorState 组件
- 移除手写加载文案（"正在加载…"、"加载中…"）
- **ErrorState retry（2026-07-21 决策 A）**
  - `ErrorState` 增加可选 `onRetry?: () => void` 与可选 `retryLabel?: string`
  - 传入 `onRetry` 时在错误区内渲染 retry 按钮；未传则保持纯文案
  - 文案由调用方传入已翻译 `retryLabel`（或组件内 useI18n 取 `t('retry')`——实现时二选一，优先 props 以保持 StatePanels 可测）
  - Home 等旁挂 retry 的写法改为走 `onRetry`
- 有数据加载的页面在 error 时优先接 `onRetry`（Home, Library, TagsIndex, TagDetail, TitleDetail, UserList, MissingItems 等）
- **Admin loading/error（2026-07-21 决策 B）**
  - 不引入整页 LoadingState（无整页数据加载）
  - scan / generate_thumbnails **启动失败**时，在动作卡片区域下方（或网格旁）展示 `ErrorState` + `onRetry` 重试该动作
  - 进行中文案与成功结果仍保留在 admin card 内；`pushAlert` 可保留作次要反馈或与 ErrorState 二选一以免双提示（实现时优先 ErrorState 区块，避免重复弹两条）

### R4: 内联样式替换为语义化工具类
- 新增少量语义化工具类到 shell.css（如 `.mango-mt-0`, `.mango-mb-1`, `.mango-max-w-search`, `.mango-scroll-x`）
- 替换各页面的 inline style（marginTop, maxWidth, overflowX 等）
- 保留真正动态的 inline（如 ProgressBar `width: ${bounded}%`、Reader margin）
- 不引入完整 spacing scale，保持 BEM-ish 命名传统

### R5: Reader 和硬编码颜色纳入 token 系统
- Reader chrome 颜色提取为 `--mango-reader-*` CSS 变量，保持深色沉浸式外观不变
- success alert 绿色加入 tokens.css
- ghost button 边框色改用 token
- primary button 文字色 token 化
- Reader 不接入四主题切换（仅 token 化以提升可维护性）

### R6: 按钮 comic 主题处理 + destructive 变体
- .mango-btn 在 comic 主题下应用厚边框 + offset shadow（与 panel/card 一致）
- destructive（`.mango-btn--danger`）使用独立红色（flat: `#d32f2f`, comic: `#c62828`），与 accent 区分
- 统一 ConfirmDialog/EditDialog/Login 的按钮用法

### R7: 表单字段标记统一
- 标准化为 `label.mango-field > span + control` 模式（label 作为容器）
- checkbox 使用 `.mango-field--inline` 变体
- file input 样式优化
- topbar select 使用 mango-input 样式
- （Login 当前 `div > label + input` 需迁移；若与 a11y/htmlFor 冲突再在 design 中定稿）

### R8: 删除 React foundation Placeholder 页（2026-07-21 决策）

用户确认该预览页已无用，本任务内移除脚手架而非纳入 i18n 验收。

- 删除 `frontend/src/pages/PlaceholderPage.tsx`
- `App.tsx` 移除 `react-preview` 分支与 import
- 删除后端 `GET /admin/react-preview`（`handleReactPreview` + 路由注册）
- `boot.ts` 的 `DEFAULT_BOOT.pageId` 改为更安全的开发默认（推荐 `home`，避免误落到已删页）
- 更新 `FRONTEND_DEV_GUIDE.md` 中对 `/admin/react-preview` 的说明
- **例外说明**：删除该 admin 预览路由属于移除死脚手架，**不**算作业务 API 变更；仍禁止改 library/tag 等数据 API

## Acceptance Criteria

- [ ] TagDetail 无手写卡片 markup；使用 PosterCard（`showProgress={false}`）；卡片数据经前端适配映射
- [ ] TagDetail 使用 BrowseToolbar；排序选项仅 natural + title（`modes` 限制）
- [ ] BrowseToolbar 支持可选 `modes`；未传时 Library/TitleDetail 行为不变
- [ ] PosterCard 支持可选 `showProgress`（默认 true）；Library/Home 行为不变
- [ ] Login 全量 i18n，并在登录页提供语言切换入口（localStorage 与 AppShell 一致）
- [ ] 语言切换后目标页面文案跟随切换（无残留硬编码用户可见中文）
- [ ] 所有业务页面的 loading/empty/error 使用共享组件
- [ ] ErrorState 支持可选 `onRetry`；可重试页面接入，Home 不再旁挂独立 retry 按钮
- [ ] `style={{` 仅保留真正动态样式（progress width、reader margin 等）
- [ ] Reader/success/ghost/primary 相关颜色引用 CSS 变量
- [ ] comic 主题下按钮有视觉区分；destructive 有独立色，不与 accent 混用
- [ ] 表单统一为约定的 field 标记模式
- [ ] 四种主题组合（comic/flat × light/dark）下无视觉回归
- [ ] 未改动业务后端 API（含 tag titles 响应形状）；仅允许删除 `react-preview` 脚手架路由
- [ ] PlaceholderPage / `react-preview` 路由与文档引用已移除；DEFAULT_BOOT 不再指向该页

## Out of Scope

- 新增页面或功能
- 业务后端 API 变更（含扩展 tag titles 字段以对齐 library card）
- 阅读器核心渲染逻辑重构
- 移动端响应式适配（现有行为保持不变）
- Placeholder 页的 i18n/美化（页本身删除）

## Open Questions

1. ~~TagDetail BrowseToolbar 排序呈现~~ → **B：`modes` 可选过滤**
2. ~~ErrorState retry~~ → **A：`onRetry` / `retryLabel` 可选 prop**
3. ~~Placeholder 验收~~ → **删除页 + 路由 + 文档引用（R8）**
4. ~~Admin loading/error 深度~~ → **B：动作启动失败用 ErrorState+onRetry，无整页 loading**
5. ~~Login 语言切换入口~~ → **A：Login 接 i18n + 页内 language select**
6. ~~TagDetail 空进度条~~ → **A：PosterCard `showProgress` 可选，TagDetail=false**
7. ~~任务拆分~~ → **A：单任务，design.md + implement.md 按底座→页面排序**

## Delivery Shape

- 单任务交付，不拆 parent/child
- 复杂任务：启动实现前完成 `design.md` + `implement.md`
- 实施顺序原则：共享组件 / tokens / 工具类 → 业务页面消费 → 删除 preview 脚手架 → 主题回归

## Planning Status

- [x] `prd.md` decisions locked (R1–R8)
- [x] `design.md` written
- [x] `implement.md` written
- [x] `implement.jsonl` / `check.jsonl` curated
- [ ] User review / approval to `task.py start`

## Notes

- 本任务跨多页面 + 设计 token + 共享组件，视为 **complex**：`design.md` + `implement.md` 已备齐。
- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- 实现未开始；待你确认规划后执行 `python3 ./.trellis/scripts/task.py start`。
