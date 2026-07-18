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
  },
} as const;

export type MessageKey = keyof typeof messages['zh-cn'];
type I18nValue = { language: Language; setLanguage: (language: Language) => void; t: (key: MessageKey) => string };

const I18nContext = createContext<I18nValue>({ language: 'zh-cn', setLanguage: () => undefined, t: (key) => messages['zh-cn'][key] });

function storedLanguage(): Language {
  const value = localStorage.getItem('mango-language');
  return value === 'zh-tw' || value === 'en' ? value : 'zh-cn';
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [language, setLanguageState] = useState<Language>(storedLanguage);
  const value = useMemo<I18nValue>(() => ({
    language,
    setLanguage: (next) => { localStorage.setItem('mango-language', next); setLanguageState(next); },
    t: (key) => messages[language][key] ?? messages['zh-cn'][key],
  }), [language]);

  useEffect(() => {
    document.documentElement.lang = language === 'en' ? 'en' : language === 'zh-tw' ? 'zh-Hant' : 'zh-Hans';
    document.title = `Mango - ${messages[language].library}`;
  }, [language]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() { return useContext(I18nContext); }
