# Implement: comic font consistency

## Ordered checklist

1. **Obtain fonts**
   - 下载 Fredoka WOFF2（Regular 400 + Bold 700）与 OFL 许可文本
   - 放入 `frontend/src/assets/fonts/fredoka/`（路径以 implement 时最终为准）
2. **@font-face**
   - 新增 `frontend/src/styles/fonts.css`（或等价）
   - `main.tsx` import
3. **Tokens**
   - 更新 `--mango-font-comic` 栈（Fredoka + Noto CJK + 系统 CJK）
   - 删除 `--mango-font-sound`
   - 删除/改写过时 webfonts 注释
4. **Shell CSS**
   - comic body → `--mango-font-comic`
   - 移除 brand / page-header h1 / login h1 冗余 `font-family` 覆盖
5. **Spec**
   - 更新 `.trellis/spec/frontend/ui-theme-layout.md` font 合同
6. **Validate**
   - `npm run typecheck`
   - `npm run build`
   - 检查 build 产物含 woff2；comic/flat 切换字体；Reader 未误用 comic 栈

## Validation commands

```bash
npm run typecheck
npm run build
# optional: ls go/web/public/react/assets/*.woff2
```

## Risky files

| Area | Risk |
|------|------|
| Vite asset names | woff2 路径错误 → 404 → 回退系统 |
| emptyOutDir | 确认字体随 CSS 引用被打包，不丢 |
| CJK stack order | 错误族名导致中文回退难看 |
| Reader | 误改 reader 字体 |

## Review gates before start

- [x] prd 决策齐：全站同一栈 / Fredoka+Noto CJK / 混合加载 / Reader 不跟
- [x] design 有栈、@font-face、CSS 应用
- [x] implement 有序清单
- [ ] 用户审阅并同意 `task.py start`

## implement.jsonl / check.jsonl

- `.trellis/spec/frontend/ui-theme-layout.md`
- 本任务 prd / design / implement
- 可选：`react-reader.md`（确认不改 reader 字体）
