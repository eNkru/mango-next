# 优化书籍详情页UI

## Goal

提升 `/book/{id}` 页面（TitleDetailPage）的信息密度和可浏览性，让用户更快找到想看的章节。

## Design Decisions (已确认)

- **分组模式**: 始终按未开始/阅读中/已完成分组。用户搜索时恢复平铺列表。每个组可折叠/展开，默认全部展开。
- **文件名**: 缩小字号、淡化颜色，保留在标签列表下方。
- **快速跳转**: 选中后直接跳转阅读器页面 `/reader/{title_id}/{entry_id}/1`。

## Requirements

1. **工具栏显示条目数量** - 章节标题旁显示 "共 X 章"，搜索过滤时同步变化
2. **标题区加入阅读摘要** - 封面旁边加一行进度文字，如 "已读 3/24 章"
3. **封面 hover 动效** - 封面图片 hover 时 scale(1.03) transition 0.2s
4. **Title overview 区布局优化** - 文件名缩小淡化；阅读摘要放在进度条上方
5. **章节列表按阅读状态分区** - 未开始 / 阅读中 / 已完成三组，每组可折叠（`showMore`/`showLess`），默认展开。搜索时恢复平铺列表
6. **快速跳转下拉框** - 工具栏增加 `<select>`，选中后 `window.location.href` 跳转阅读器
7. **子系列区域增加视觉分隔** - 加顶部边框和额外间距

## Acceptance Criteria

- [ ] 章节标题旁显示 "共 N 章"，搜索或过滤时计数同步更新
- [ ] 标题概览区显示阅读进度文字（如 "已读 3/24 章 · 12%"）
- [ ] 封面 hover 时有 0.2s 的 scale(1.03) 过渡动画
- [ ] 文件名缩小字号（0.85em）并淡化（--mango-text-muted）
- [ ] 章节按未开始/阅读中/已完成分组，每组可折叠/展开
- [ ] 快速跳转 `<select>` 选中后直接跳转到对应章节的阅读器
- [ ] 子系列区域和章节区域之间有明显的视觉分隔（顶部边框 + margin-top）
- [ ] 三个语言（zh-cn, zh-tw, en）的新增文案都有翻译
- [ ] 移动端（≤560px）布局正常
- [ ] comic theme 下样式正常

## Files

- `frontend/src/pages/TitleDetailPage.tsx` — 页面组件
- `frontend/src/styles/shell.css` — 样式
- `frontend/src/lib/i18n.tsx` — 新增翻译 key

## Out of Scope

- 后端 API 变更
- 数据结构变更
- 修改其他页面
