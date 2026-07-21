# Design: comic font consistency

## Architecture

```
frontend/public/fonts/fredoka/   # or frontend/src/assets/fonts/
  Fredoka-Regular.woff2
  Fredoka-Bold.woff2
        │
        ▼
frontend/src/styles/fonts.css    # @font-face Fredoka
tokens.css                       # --mango-font-comic stack
shell.css                        # comic body → comic token; drop redundant heading-only overrides
vite build → go/web/public/react/assets/*  (woff2 hashed/named)
```

## Font stack

### Comic (`--mango-font-comic`)

```css
--mango-font-comic:
  "Fredoka",
  "Noto Sans CJK SC",
  "Noto Sans CJK TC",
  "Noto Sans SC",
  "Noto Sans TC",
  "PingFang SC",
  "Hiragino Sans GB",
  "Microsoft YaHei",
  "Segoe UI",
  sans-serif;
```

- Latin → Fredoka（自托管）
- CJK → 系统已装 Noto / PingFang / YaHei 等（不打包全量 CJK）
- 去掉 `Fredoka One` / `Comic Sans MS` / `cursive` 作为主路径（避免运气渲染）

### Flat (`--mango-font-body`)

保持：

```css
"Segoe UI", "Helvetica Neue", Arial, sans-serif
```

### Remove

- `--mango-font-sound`（零引用死 token）
- 过时 “served from /webfonts” 注释

## CSS application

| 规则 | 行为 |
|------|------|
| `body` | 默认 `--mango-font-body` |
| `html.comic-theme body`, `html.comic-theme-dark body` | **改为** `--mango-font-comic` |
| 原先 brand / page-header h1 / login h1 的 comic 覆盖 | **删除**（已继承 body，避免双重规则） |
| Reader | 不新增 comic font 规则；沿用现有 reader 文本样式 |

## @font-face

```css
@font-face {
  font-family: "Fredoka";
  src: url("../fonts/fredoka/Fredoka-Regular.woff2") format("woff2");
  font-weight: 400;
  font-style: normal;
  font-display: swap;
}
@font-face {
  font-family: "Fredoka";
  src: url("../fonts/fredoka/Fredoka-Bold.woff2") format("woff2");
  font-weight: 700;
  font-style: normal;
  font-display: swap;
}
```

- 字重：400 + 700 足够 UI；不引入可变字体除非体积更优。
- 许可：Fredoka 为 SIL OFL；在字体目录旁保留 `OFL.txt` 或 README 一行来源。
- 获取：实现时从 Google Fonts 官方下载 WOFF2（不在运行时拉 CDN）。

## Vite / asset path

- 字体放在 `frontend/src/assets/fonts/fredoka/`（或 `frontend/public/fonts/`）。
- 若走 `src/assets`：由 CSS `url()` 引入，Vite 打进 `go/web/public/react/assets/`（`assetFileNames` 已允许非 CSS 资源）。
- `base: './'` 下相对 URL 可工作。
- `main.tsx` 增加 `import './styles/fonts.css'`（或并入 tokens）。

## Spec updates

`ui-theme-layout.md`：

- Token 表增加 font tokens
- Comic：body 用 `--mango-font-comic`；Fredoka 自托管；CJK 系统栈
- Flat：`--mango-font-body` 不变
- Reader：不强制 comic font

## Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| 全 comic 同一栈 | 一致、实现简单 | display 感正文略圆；用户已选 B |
| Fredoka 自托管 | 可控、无 CDN | 增加 ~几十 KB WOFF2 |
| CJK 系统栈 | 无百 MB 包体 | 未装 Noto 时用 PingFang/YaHei，仍一致于系统 |
| 删 heading 特例 | 真·全站一致 | 标题不再单独「更漫画」——与决策 B 一致 |

## Rollback

还原 CSS/token/fonts 文件与 build 产物；无后端/API 变更。
