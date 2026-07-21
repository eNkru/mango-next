# Unify comic style font family

## Goal

让 comic UI 的字体族一致、可预期，并选用适合漫画气质的字体栈（含可靠加载方式），消除「只在少数标题碰巧变成 Comic Sans / 系统默认」的割裂感。

## Confirmed facts（仓库可证）

- Token：`--mango-font-body`（系统 sans）、`--mango-font-comic`（Fredoka One…）、`--mango-font-sound`（未用）。
- Comic 下 body 仍用 body 栈；仅 brand / page-header h1 / login h1 用 comic token。
- 无 `@font-face`、无字体文件、无 Google Fonts；legacy webfonts 已删。
- Fredoka 无 CJK；Noto Sans CJK 全量过大不宜进仓。

## Decisions

| 决策 | 选择 |
|------|------|
| 应用范围 | **全站同一 comic 字体栈**（AppShell / 非 Reader 路由） |
| 字体族 | **Fredoka（拉丁）+ Noto Sans CJK 族名 + 系统 CJK 回退** |
| 加载 | **混合**：Fredoka 自托管 WOFF2；CJK 系统优先（栈内写 SC/TC + PingFang/YaHei 等） |
| Reader | **不跟随** comic 字体（保持现有 reader chrome） |

## Requirements

1. Comic 主题下 UI（body、导航、按钮、表单、标题等）共用同一 `--mango-font-comic` 栈，去掉「仅 3 处 heading」分裂。
2. Fredoka 经 `@font-face` 自托管可靠加载（至少 regular + bold）。
3. 字体栈含中文路径：`"Noto Sans CJK SC", "Noto Sans CJK TC", "PingFang SC", "Microsoft YaHei", sans-serif` 等（design 定最终顺序）。
4. Flat 主题保持 `--mango-font-body` 不变。
5. 删除/停用未用的 `--mango-font-sound`；修正过时 webfonts 注释。
6. Reader 不强制 comic 字体。
7. 更新 `ui-theme-layout.md`（及必要时 component-guidelines）font 合同。

## Acceptance Criteria

- [ ] `html.comic-theme` / `comic-theme-dark` 下 body 与主要 UI 使用同一 comic font-family token
- [ ] 拉丁：Fredoka 经自托管 webfont 生效（无本机字体时仍可用）
- [ ] 中文：栈含 Noto Sans CJK SC/TC 与常见系统 CJK 回退
- [ ] flat 主题字体不变
- [ ] 仓库不引入全量 Noto CJK 二进制
- [ ] Reader 字体策略不因本任务改为 comic display
- [ ] frontend spec 写明 comic 字体规则
- [ ] `npm run typecheck` / `npm run build` 通过

## Out of scope

- Flat 改字体
- Font Awesome / 图标 webfonts
- 全量自托管 Noto CJK / 默认外网 CDN
- Reader 沉浸 chrome 漫画字体化

## Notes

- 复杂任务：design + implement 后再 `task.py start`。
