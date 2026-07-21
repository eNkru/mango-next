# Implement: restore icons UI/UX

## Ordered checklist

1. **Deps**
   - `npm install lucide-react`（repo root `package.json`）
2. **Foundation**
   - 新增 `frontend/src/shell/Icon.tsx`（薄封装）
   - 可选 `frontend/src/shell/icons.ts` 语义映射
   - 扩展 `frontend/src/styles/shell.css`：btn icon gap、icon-only、brand mark
3. **Chrome**
   - `AppShell.tsx`：mark + nav icons + logout
   - `BrowseComponents.tsx`：搜索/排序相关图标
4. **Pages（按操作密度）**
   - `TitleDetailPage.tsx`
   - `LibraryPage.tsx` / `TagDetailPage.tsx` / `TagsIndexPage.tsx`
   - `UserListPage.tsx` / `UserEditPage.tsx` / `MissingItemsPage.tsx`
   - `AdminPage.tsx` / `HomePage.tsx` / `LoginPage.tsx`
5. **Reader**
   - `ReaderTopBar.tsx` / `ReaderControls.tsx`（及必要的 ReaderPage 出口按钮）
6. **Shared**
   - `ConfirmDialog.tsx` / `StatePanels.tsx`：危险/重试可加图标
7. **Spec**
   - 更新 `.trellis/spec/frontend/component-guidelines.md`（Icon 契约）
   - 必要时补 `ui-theme-layout.md` 一句 brand/icon density
8. **Validate**
   - `npm run typecheck`
   - `npm run build`
   - 手动扫：topbar、library、title detail、reader、login、admin（default + comic）

## Validation commands

```bash
npm run typecheck
npm run build
```

## Risky files / rollback

| Area | Files | Risk |
|------|-------|------|
| Bundle size | package.json + lucide imports | 必须按图标具名 import，禁止整库 import |
| Brand path | AppShell + baseUrl | 错误 base path 导致 mark 404 |
| Layout | shell.css topbar | 图标撑破窄屏 topbar |
| a11y | icon-only buttons | 漏 aria-label |

Rollback：还原 frontend 改动与依赖锁定；无服务端变更。

## Review gates before start

- [x] prd.md 决策齐：全站 / 密度混用 / mark / lucide
- [x] design.md 有组件契约与覆盖面
- [x] implement.md 有序清单与验证命令
- [ ] 用户审阅规划并同意 `task.py start`

## implement.jsonl / check.jsonl（start 前配置）

建议纳入：

- `.trellis/spec/frontend/component-guidelines.md`
- `.trellis/spec/frontend/ui-theme-layout.md`
- 本任务 `prd.md` / `design.md` / `implement.md`
- 关键文件：`AppShell.tsx`、`shell.css`、`BrowseComponents.tsx`、`TitleDetailPage.tsx`、reader 组件
