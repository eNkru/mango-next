# Implement: Comic Netflix-layout (3 batches)

## Batch 1 — Shell + top bar for comic

- [x] `flat-theme.less` / `.css`：顶栏可见性改为 flat|comic 共享（非 flat 且非 comic 才隐藏）
- [x] `comic-theme.less` / `.css`：
  - hide `.app-sidebar` under comic（含 html FOUC 选择器）
  - show/skin `.flat-topbar` with comic palette/border/type
  - `app-content` 全宽 + top padding 对齐 Flat 几何
  - 减弱 `comic-card-pop`（去 rotate；缓入 scale）
- [x] lessc 编译 `comic-theme.css`；flat-theme.css 手改可见性（避免冲掉 hand-sync 批次）
- [x] `top.tmpl` 注释更新（共享顶栏）
- [ ] Smoke：comic↔flat、dark↔light（用户目视）

**Validate:** 目视 + 无 JS 错误；侧栏在两套 UI 下均不主导航

## Batch 2 — Browse surfaces

- [x] Home continue-reading 单行 rail 高度结构对齐 Flat
- [x] Library/carousel 海报 2:3；卡片 hover 改为轻微抬升
- [x] page-heading 去倾斜
- [x] Tags/Title 既有 comic 组件语言保留（pill / select-bar 等）
- [ ] 用户 smoke continue reading / 打开书 / tags

## Batch 3 — Rest of app

- [x] Login / Admin / Reader / Select2 既有 comic 皮肤在新壳下可用（无删功能；壳几何来自 Batch 1）
- [ ] 用户 smoke admin / reader

## Cross-cutting

- [x] Flat 顶栏可见性最小 diff；flat 视觉 token 未改
- [x] 未将 comic 主 CTA 改为 Netflix 红
- [x] less + 编译 css
- [ ] 稳定契约写入 `.trellis/spec/frontend`（可选）

## Done when

PRD acceptance 用户 smoke 勾满；可提交。
