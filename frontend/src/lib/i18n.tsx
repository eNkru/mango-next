import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';

export type Language = 'zh-cn' | 'zh-tw' | 'en';

const messages = {
  'zh-cn': {
    home: '主页', library: '资料库', tags: '标签', admin: '管理员', logout: '退出',
    language: '语言', loading: '正在加载…', retry: '重试', search: '搜索', sort: '排序',
    ascending: '升序', descending: '降序', automatic: '自然顺序', title: '名称',
    modified: '最近更新', progress: '阅读进度', entries: '项', hidden: '已隐藏',
    showHidden: '显示隐藏', hideHidden: '隐藏已隐藏', hide: '隐藏', show: '显示',
    continueReading: '继续阅读', startReading: '开始阅读', recentlyAdded: '最近添加',
    emptyLibrary: '资料库是空的', emptyLibraryAdmin: '添加漫画到资料库目录后运行扫描。',
    emptyLibraryUser: '资料库暂时没有可阅读的内容。', welcome: '欢迎使用 Mango',
    welcomeBody: '从下方选择一部漫画开始阅读。', noResults: '没有匹配结果',
    open: '打开', read: '已读', unread: '未读', download: '下载', edit: '编辑',
    save: '保存', cancel: '取消', displayName: '显示名称', sortName: '排序名称',
    fileName: '文件名称', cover: '封面', upload: '上传封面', children: '子标题',
    chapters: '条目', selectAll: '全选', clearSelection: '清除选择', selected: '已选择',
    markRead: '标记已读', markUnread: '标记未读', addTag: '添加标签', remove: '移除',
    titleDetail: '标题详情', librarySubtitle: '浏览全部漫画', homeSubtitle: '继续你的阅读',
    page: '页', begin: '从头阅读', continue: '继续', actions: '操作',
    reader: '阅读器', readerControls: '阅读设置', exitReader: '退出阅读',
    jumpToPage: '跳转到页面', jumpToEntry: '跳转条目', readingMode: '阅读模式',
    modeContinuous: '连续', modePaged: '分页', fitPage: '适合页面', fitHeight: '高度',
    fitWidth: '宽度', fitOriginal: '原始', pageMargin: '页间距', preloadLookahead: '预加载',
    flipAnimation: '翻页动画', rightToLeft: '从右到左', nextEntry: '下一项',
    previousEntry: '上一项', readerError: '无法打开阅读器',
    adminSubtitle: '系统设置与操作', userManagement: '用户管理',
    userManagementDesc: '添加、编辑或删除用户', missingEntries: '缺少条目',
    missingEntriesDesc: '查看缺失的文件条目', scanLibrary: '扫描库文件',
    scanLibraryDesc: '重新扫描资料库目录', scanning: '正在扫描…',
    scanResult: '扫描完成：{count} 个标题，用时 {ms} ms', scanFailed: '扫描失败',
    generateThumbnails: '生成缩略图', generateThumbnailsDesc: '为漫画生成封面缩略图',
    thumbFailed: '缩略图任务失败', theme: '主题', themeSystem: '跟随系统',
    themeLight: '浅色', themeDark: '深色', uiStyle: 'UI风格',
    uiStyleComic: 'Comic', uiStyleFlat: 'Flat',
    showMore: '展开更多', showLess: '收起',
    totalChapters: '共 {count} 章', totalEntries: '共 {count} 项',
    readProgress: '已读 {read}/{total} 章',
    unreadGroup: '未开始', readingGroup: '阅读中', completedGroup: '已完成',
    quickJumpTo: '跳转到...',
    // Login
    loginWelcome: '欢迎回来', loginSubtitle: '登录到 Mango', username: '用户名',
    password: '密码', usernamePlaceholder: '请输入用户名', passwordPlaceholder: '请输入密码',
    showPassword: '显示', hidePassword: '隐藏', login: '登录', loggingIn: '登录中…',
    loginFailed: '登录失败，请检查用户名和密码', loginFooter: 'Mango · 漫画服务器',
    // Tags
    tagsCount: '{count} 个标签', tagCount: '{count} 个标记', filterTags: '筛选标签…',
    refresh: '刷新', noMatchingTags: '未找到匹配标签', noTagsYet: '还没有标签',
    loadTagsFailed: '加载标签失败', allTags: '全部标签', findManga: '查找漫画…',
    noMangaInTag: '该标签下没有漫画', missingTag: '缺少标签', loadTagFailed: '加载标签失败',
    tagTitle: '标签: {tag}', hiddenDone: '已隐藏', shownDone: '已显示', actionFailed: '操作失败',
    // Users
    userListSubtitle: '创建、编辑和管理可登录用户', loadingUsers: '正在加载用户…',
    noUsersYet: '还没有用户', newUser: '新用户', currentUser: '当前', yes: '是', no: '否',
    delete: '删除', confirmDeleteUser: '确认删除用户？',
    deleteUserMessage: '将删除用户 {username}。此操作不可撤销。',
    userDeleted: '已删除用户 {username}', deleteFailed: '删除失败', loadUsersFailed: '加载用户失败',
    adminPermission: '管理员权限', editUser: '编辑用户', createAccount: '创建可登录账户',
    editUserSubtitle: '编辑 {username}', newPassword: '新密码', changePassword: '更改密码',
    userCreated: '用户已创建', userUpdated: '用户已更新', saveFailed: '保存失败',
    saving: '保存中…', backToList: '返回列表', loadUserFailed: '加载用户失败',
    // Missing items
    missingTitle: '缺失条目',
    missingSubtitle: '资料库中记录存在，但磁盘上已找不到对应文件的项目',
    loadingMissing: '正在加载缺失条目…',
    noMissingItems: '没有找到丢失的条目，所有条目均正常',
    missingHelp:
      '以下项目存在于资料库元数据中，但现在找不到对应文件。若误删，请恢复文件后重新扫描；否则可删除元数据以释放数据库空间。',
    deleteAll: '删除全部', type: '类型', relativePath: '相对路径', id: 'ID',
    titleKind: '标题', pathKind: '路径', deleted: '已删除',
    deletedAllMissing: '已删除全部缺失项', bulkDeleteFailed: '批量删除失败',
    confirmDeleteAll: '确认删除全部？',
    confirmDeleteAllMessage: '与这些项目相关的所有元数据，包括标签和缩略图，都将从数据库中删除。',
    confirmDeleteAllYes: '是的，删除它们', loadFailed: '加载失败',
    // Misc
    unknownPage: '未知页面',
    unknownPageMessage: '没有为 pageId={pageId} 注册的 React 页面',
    missingTitleId: '缺少标题 ID',
  },
  'zh-tw': {
    home: '首頁', library: '資料庫', tags: '標籤', admin: '管理員', logout: '登出',
    language: '語言', loading: '正在載入…', retry: '重試', search: '搜尋', sort: '排序',
    ascending: '升序', descending: '降序', automatic: '自然順序', title: '名稱',
    modified: '最近更新', progress: '閱讀進度', entries: '項', hidden: '已隱藏',
    showHidden: '顯示隱藏', hideHidden: '隱藏已隱藏', hide: '隱藏', show: '顯示',
    continueReading: '繼續閱讀', startReading: '開始閱讀', recentlyAdded: '最近加入',
    emptyLibrary: '資料庫是空的', emptyLibraryAdmin: '加入漫畫到資料庫目錄後執行掃描。',
    emptyLibraryUser: '資料庫暫時沒有可閱讀的內容。', welcome: '歡迎使用 Mango',
    welcomeBody: '從下方選擇一部漫畫開始閱讀。', noResults: '沒有符合結果',
    open: '開啟', read: '已讀', unread: '未讀', download: '下載', edit: '編輯',
    save: '儲存', cancel: '取消', displayName: '顯示名稱', sortName: '排序名稱',
    fileName: '檔案名稱', cover: '封面', upload: '上傳封面', children: '子標題',
    chapters: '條目', selectAll: '全選', clearSelection: '清除選擇', selected: '已選擇',
    markRead: '標記已讀', markUnread: '標記未讀', addTag: '加入標籤', remove: '移除',
    titleDetail: '標題詳情', librarySubtitle: '瀏覽全部漫畫', homeSubtitle: '繼續你的閱讀',
    page: '頁', begin: '從頭閱讀', continue: '繼續', actions: '操作',
    reader: '閱讀器', readerControls: '閱讀設定', exitReader: '退出閱讀',
    jumpToPage: '跳到頁面', jumpToEntry: '跳轉條目', readingMode: '閱讀模式',
    modeContinuous: '連續', modePaged: '分頁', fitPage: '適合頁面', fitHeight: '高度',
    fitWidth: '寬度', fitOriginal: '原始', pageMargin: '頁間距', preloadLookahead: '預載入',
    flipAnimation: '翻頁動畫', rightToLeft: '從右到左', nextEntry: '下一項',
    previousEntry: '上一項', readerError: '無法開啟閱讀器',
    adminSubtitle: '系統設定與操作', userManagement: '使用者管理',
    userManagementDesc: '新增、編輯或刪除使用者', missingEntries: '缺少條目',
    missingEntriesDesc: '查看缺失的檔案條目', scanLibrary: '掃描資料庫',
    scanLibraryDesc: '重新掃描資料庫目錄', scanning: '正在掃描…',
    scanResult: '掃描完成：{count} 個標題，用時 {ms} ms', scanFailed: '掃描失敗',
    generateThumbnails: '產生縮圖', generateThumbnailsDesc: '為漫畫產生封面縮圖',
    thumbFailed: '縮圖任務失敗', theme: '主題', themeSystem: '跟隨系統',
    themeLight: '淺色', themeDark: '深色', uiStyle: 'UI風格',
    uiStyleComic: 'Comic', uiStyleFlat: 'Flat',
    showMore: '展開更多', showLess: '收起',
    totalChapters: '共 {count} 章', totalEntries: '共 {count} 項',
    readProgress: '已讀 {read}/{total} 章',
    unreadGroup: '未開始', readingGroup: '閱讀中', completedGroup: '已完成',
    quickJumpTo: '跳轉到...',
    loginWelcome: '歡迎回來', loginSubtitle: '登入到 Mango', username: '使用者名稱',
    password: '密碼', usernamePlaceholder: '請輸入使用者名稱', passwordPlaceholder: '請輸入密碼',
    showPassword: '顯示', hidePassword: '隱藏', login: '登入', loggingIn: '登入中…',
    loginFailed: '登入失敗，請檢查使用者名稱和密碼', loginFooter: 'Mango · 漫畫伺服器',
    tagsCount: '{count} 個標籤', tagCount: '{count} 個標記', filterTags: '篩選標籤…',
    refresh: '重新整理', noMatchingTags: '未找到符合標籤', noTagsYet: '還沒有標籤',
    loadTagsFailed: '載入標籤失敗', allTags: '全部標籤', findManga: '尋找漫畫…',
    noMangaInTag: '該標籤下沒有漫畫', missingTag: '缺少標籤', loadTagFailed: '載入標籤失敗',
    tagTitle: '標籤: {tag}', hiddenDone: '已隱藏', shownDone: '已顯示', actionFailed: '操作失敗',
    userListSubtitle: '建立、編輯和管理可登入使用者', loadingUsers: '正在載入使用者…',
    noUsersYet: '還沒有使用者', newUser: '新使用者', currentUser: '目前', yes: '是', no: '否',
    delete: '刪除', confirmDeleteUser: '確認刪除使用者？',
    deleteUserMessage: '將刪除使用者 {username}。此操作無法復原。',
    userDeleted: '已刪除使用者 {username}', deleteFailed: '刪除失敗', loadUsersFailed: '載入使用者失敗',
    adminPermission: '管理員權限', editUser: '編輯使用者', createAccount: '建立可登入帳戶',
    editUserSubtitle: '編輯 {username}', newPassword: '新密碼', changePassword: '更改密碼',
    userCreated: '使用者已建立', userUpdated: '使用者已更新', saveFailed: '儲存失敗',
    saving: '儲存中…', backToList: '返回列表', loadUserFailed: '載入使用者失敗',
    missingTitle: '缺失條目',
    missingSubtitle: '資料庫中記錄存在，但磁碟上已找不到對應檔案的項目',
    loadingMissing: '正在載入缺失條目…',
    noMissingItems: '沒有找到遺失的條目，所有條目均正常',
    missingHelp:
      '以下項目存在於資料庫中繼資料中，但現在找不到對應檔案。若誤刪，請還原檔案後重新掃描；否則可刪除中繼資料以釋放資料庫空間。',
    deleteAll: '刪除全部', type: '類型', relativePath: '相對路徑', id: 'ID',
    titleKind: '標題', pathKind: '路徑', deleted: '已刪除',
    deletedAllMissing: '已刪除全部缺失項', bulkDeleteFailed: '批量刪除失敗',
    confirmDeleteAll: '確認刪除全部？',
    confirmDeleteAllMessage: '與這些項目相關的所有中繼資料，包括標籤和縮圖，都將從資料庫中刪除。',
    confirmDeleteAllYes: '是的，刪除它們', loadFailed: '載入失敗',
    unknownPage: '未知頁面',
    unknownPageMessage: '沒有為 pageId={pageId} 註冊的 React 頁面',
    missingTitleId: '缺少標題 ID',
  },
  en: {
    home: 'Home', library: 'Library', tags: 'Tags', admin: 'Admin', logout: 'Log out',
    language: 'Language', loading: 'Loading…', retry: 'Retry', search: 'Search', sort: 'Sort',
    ascending: 'Ascending', descending: 'Descending', automatic: 'Natural order', title: 'Title',
    modified: 'Recently updated', progress: 'Reading progress', entries: 'items', hidden: 'Hidden',
    showHidden: 'Show hidden', hideHidden: 'Hide hidden', hide: 'Hide', show: 'Show',
    continueReading: 'Continue reading', startReading: 'Start reading', recentlyAdded: 'Recently added',
    emptyLibrary: 'Your library is empty', emptyLibraryAdmin: 'Add manga to the library path, then run a scan.',
    emptyLibraryUser: 'There is nothing available to read yet.', welcome: 'Welcome to Mango',
    welcomeBody: 'Choose a title below to begin reading.', noResults: 'No matching results',
    open: 'Open', read: 'Read', unread: 'Unread', download: 'Download', edit: 'Edit',
    save: 'Save', cancel: 'Cancel', displayName: 'Display name', sortName: 'Sort name',
    fileName: 'File name', cover: 'Cover', upload: 'Upload cover', children: 'Child titles',
    chapters: 'Entries', selectAll: 'Select all', clearSelection: 'Clear selection', selected: 'selected',
    markRead: 'Mark read', markUnread: 'Mark unread', addTag: 'Add tag', remove: 'Remove',
    titleDetail: 'Title details', librarySubtitle: 'Browse every title', homeSubtitle: 'Pick up where you left off',
    page: 'pages', begin: 'Read from start', continue: 'Continue', actions: 'Actions',
    reader: 'Reader', readerControls: 'Reading settings', exitReader: 'Exit reader',
    jumpToPage: 'Jump to page', jumpToEntry: 'Jump to entry', readingMode: 'Reading mode',
    modeContinuous: 'Continuous', modePaged: 'Paged', fitPage: 'Fit page', fitHeight: 'Height',
    fitWidth: 'Width', fitOriginal: 'Original', pageMargin: 'Page margin', preloadLookahead: 'Preload',
    flipAnimation: 'Flip animation', rightToLeft: 'Right to left', nextEntry: 'Next entry',
    previousEntry: 'Previous entry', readerError: 'Unable to open reader',
    adminSubtitle: 'System settings and operations', userManagement: 'User management',
    userManagementDesc: 'Add, edit, or remove users', missingEntries: 'Missing entries',
    missingEntriesDesc: 'Review missing library files', scanLibrary: 'Scan library',
    scanLibraryDesc: 'Rescan the library directory', scanning: 'Scanning…',
    scanResult: 'Scan finished: {count} titles in {ms} ms', scanFailed: 'Scan failed',
    generateThumbnails: 'Generate thumbnails', generateThumbnailsDesc: 'Build cover thumbnails',
    thumbFailed: 'Thumbnail job failed', theme: 'Theme', themeSystem: 'System',
    themeLight: 'Light', themeDark: 'Dark', uiStyle: 'UI style',
    uiStyleComic: 'Comic', uiStyleFlat: 'Flat',
    showMore: 'Show more', showLess: 'Show less',
    totalChapters: '{count} chapters', totalEntries: '{count} entries',
    readProgress: '{read}/{total} read',
    unreadGroup: 'Unread', readingGroup: 'In progress', completedGroup: 'Completed',
    quickJumpTo: 'Jump to...',
    loginWelcome: 'Welcome back', loginSubtitle: 'Sign in to Mango', username: 'Username',
    password: 'Password', usernamePlaceholder: 'Enter username', passwordPlaceholder: 'Enter password',
    showPassword: 'Show', hidePassword: 'Hide', login: 'Sign in', loggingIn: 'Signing in…',
    loginFailed: 'Sign-in failed. Check username and password.', loginFooter: 'Mango · Manga server',
    tagsCount: '{count} tags', tagCount: '{count} items', filterTags: 'Filter tags…',
    refresh: 'Refresh', noMatchingTags: 'No matching tags', noTagsYet: 'No tags yet',
    loadTagsFailed: 'Failed to load tags', allTags: 'All tags', findManga: 'Find manga…',
    noMangaInTag: 'No manga under this tag', missingTag: 'Missing tag', loadTagFailed: 'Failed to load tag',
    tagTitle: 'Tag: {tag}', hiddenDone: 'Hidden', shownDone: 'Shown', actionFailed: 'Action failed',
    userListSubtitle: 'Create, edit, and manage sign-in users', loadingUsers: 'Loading users…',
    noUsersYet: 'No users yet', newUser: 'New user', currentUser: 'You', yes: 'Yes', no: 'No',
    delete: 'Delete', confirmDeleteUser: 'Delete this user?',
    deleteUserMessage: 'Delete user {username}? This cannot be undone.',
    userDeleted: 'Deleted user {username}', deleteFailed: 'Delete failed', loadUsersFailed: 'Failed to load users',
    adminPermission: 'Administrator', editUser: 'Edit user', createAccount: 'Create a sign-in account',
    editUserSubtitle: 'Edit {username}', newPassword: 'New password', changePassword: 'Change password',
    userCreated: 'User created', userUpdated: 'User updated', saveFailed: 'Save failed',
    saving: 'Saving…', backToList: 'Back to list', loadUserFailed: 'Failed to load user',
    missingTitle: 'Missing entries',
    missingSubtitle: 'Library records whose files are no longer on disk',
    loadingMissing: 'Loading missing entries…',
    noMissingItems: 'No missing entries — everything looks fine',
    missingHelp:
      'These items exist in library metadata but the files are gone. Restore the files and rescan, or delete the metadata to free database space.',
    deleteAll: 'Delete all', type: 'Type', relativePath: 'Relative path', id: 'ID',
    titleKind: 'Title', pathKind: 'Entry', deleted: 'Deleted',
    deletedAllMissing: 'Deleted all missing items', bulkDeleteFailed: 'Bulk delete failed',
    confirmDeleteAll: 'Delete everything?',
    confirmDeleteAllMessage: 'All related metadata for these items, including tags and thumbnails, will be removed from the database.',
    confirmDeleteAllYes: 'Yes, delete them', loadFailed: 'Load failed',
    unknownPage: 'Unknown page',
    unknownPageMessage: 'No React page registered for pageId={pageId}',
    missingTitleId: 'Missing title id',
  },
} as const;

export type MessageKey = keyof typeof messages['zh-cn'];
type I18nValue = {
  language: Language;
  setLanguage: (language: Language) => void;
  t: (key: MessageKey, vars?: Record<string, string | number>) => string;
};

const I18nContext = createContext<I18nValue>({
  language: 'zh-cn',
  setLanguage: () => undefined,
  t: (key) => messages['zh-cn'][key],
});

function storedLanguage(): Language {
  const value = localStorage.getItem('mango-language');
  return value === 'zh-tw' || value === 'en' ? value : 'zh-cn';
}

function formatMessage(template: string, vars?: Record<string, string | number>) {
  if (!vars) return template;
  return template.replace(/\{(\w+)\}/g, (_, name: string) =>
    vars[name] !== undefined ? String(vars[name]) : `{${name}}`,
  );
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(storedLanguage);
  const value = useMemo<I18nValue>(() => ({
    language,
    setLanguage: (next) => {
      localStorage.setItem('mango-language', next);
      setLanguageState(next);
    },
    t: (key, vars) =>
      formatMessage(messages[language][key] ?? messages['zh-cn'][key], vars),
  }), [language]);

  useEffect(() => {
    document.documentElement.lang =
      language === 'en' ? 'en' : language === 'zh-tw' ? 'zh-Hant' : 'zh-Hans';
    document.title = `Mango - ${messages[language].library}`;
  }, [language]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  return useContext(I18nContext);
}
