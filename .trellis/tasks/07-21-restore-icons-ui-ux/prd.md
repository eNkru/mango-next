# Restore icons with better UI/UX after React rewrite

## Goal

在 React 全量前端中系统化恢复有意义的图标，并统一成一致、可访问的 UI/UX 模式，消除「按钮全是纯文字、视觉层级弱」的问题。

## Confirmed facts（仓库可证）

- 产品 UI 已全部走 React（`frontend/`）；legacy 模板 / Font Awesome webfonts / UIkit 已在 `07-19-frontend-legacy-retirement` 删除。
- React foundation 刻意不做第三方组件库；页面用 `.mango-btn` + i18n 文案，**从未引入图标系统**。
- `package.json` 仅有 `react` / `react-dom`，无图标库。
- 品牌资产仍在：`go/web/public/img/icons/mango-mark.svg`（AppShell 未使用）。
- 主要操作面几乎全是纯文本按钮（AppShell、Browse、Library/Tags/Title/Users/Missing/Admin、Reader、Login）。
- `.trellis/spec/frontend/*` 尚无 Icon 约定。

## Root cause

不是运行时图标加载失败，而是 **React 实现从未建模图标**，随后 legacy 图标资产被清掉。

## Decisions

| 决策 | 选择 |
|------|------|
| 范围 | **系统化全站** |
| 展示模式 | **按密度混用**：导航/主操作 = icon+label；紧凑工具区/关闭/密码切换等 = icon-only（带 aria-label） |
| 品牌 | **恢复** 顶栏 mark + “Mango” |
| 图标库 | **lucide-react** + 薄封装 |

## Requirements

1. 依赖 `lucide-react`，提供统一 `Icon` 封装与语义名映射（页面不直接散落不一致用法）。
2. 扩展按钮/工具栏 CSS：icon+label 间距；icon-only 尺寸与 hit target。
3. 覆盖主要操作面：
   - AppShell：品牌 mark+文字、主导航 icon+label、Logout icon+label（或紧凑时 icon-only）
   - Browse 工具栏：搜索/排序方向等
   - Library / Tags / Title / Users / Missing / Admin 操作按钮
   - Reader 顶栏与控制
   - Login 密码可见切换（icon-only + aria-label）
4. 保持 dual theme / comic UI 一致。
5. 可访问性：icon-only 必须有可感知名称；装饰性图标 `aria-hidden`。
6. 实现后更新 frontend spec（Icon + 按钮图标变体）。

## Acceptance Criteria

- [ ] 存在统一 Icon 抽象；业务页通过语义名或薄封装使用图标
- [ ] 主要操作面按「按密度混用」规则使用图标
- [ ] 顶栏展示 `mango-mark.svg` + “Mango”
- [ ] 纯 icon 控件具备 aria-label / 等价可访问名称
- [ ] 主题切换（含 comic）下图标与按钮间距/对比度正常
- [ ] 不重新引入 Font Awesome webfonts / UIkit
- [ ] `npm run typecheck` / `npm run build` 通过
- [ ] frontend spec 补充 Icon / 按钮图标变体约定

## Out of scope

- 重新引入 Font Awesome webfonts / UIkit
- 引入完整第三方组件库（MUI/Ant 等）
- 重做整站视觉体系（仅图标与按钮/工具栏相关模式）
- Reader 翻页热区强制加可见图标

## Notes

- 复杂任务：需 `design.md` + `implement.md`；用户审阅后再 `task.py start`。
