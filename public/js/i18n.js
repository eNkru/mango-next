/**
 * Mango i18n / Localization System
 *
 * Client-side translation engine. HTML templates include the original
 * Simplified Chinese text and `data-i18n` attributes. On page load the
 * engine swaps text to the user-selected language.
 *
 * Usage in HTML:
 *   <span data-i18n="home">主页</span>
 *   <a data-i18n="library" data-i18n-title="library">…</a>
 *   <input data-i18n-placeholder="search_manga" placeholder="查找漫画…">
 *
 * Supported languages:
 *   zh-cn  – Simplified Chinese  (default)
 *   zh-tw  – Traditional Chinese
 *   en     – English
 */

const I18N = (function () {
  'use strict';

  const DEFAULT_LANG = 'zh-cn';

  const DICT = {
    // ================================================================
    // NAVIGATION  (layout sidebar + mobile nav)
    // ================================================================
    en: {
      home: 'Home',
      library: 'Library',
      tags: 'Tags',
      admin: 'Admin',
      download: 'Download',
      plugins: 'Plugins',
      download_manager: 'Download Manager',
      subscription_manager: 'Subscription Manager',
      theme_toggle: 'Toggle Theme',
      ui_style_toggle: 'Toggle UI Style',
      view_on_github: 'View on GitHub',
      language: 'Language',
      logout: 'Logout',
      collapse_sidebar: 'Collapse / Expand Sidebar',

      // ================================================================
      // HOME PAGE
      // ================================================================
      add_first_comic: 'Add Your First Comic',
      no_files_found: "We can't find any files yet. Add some to your library and they will appear here.",
      current_library_path: 'Current Library Path',
      change_library_path: 'Want to change your library path?',
      config_path: 'Configure \u003ccode\u003econfig.yml\u003c/code\u003e at:',
      still_cant_see: "Still can't see your files?",
      wait_scan: 'You must wait {minutes} minutes for the library scan to complete',
      or_manual_scan: ', or manually scan from ',
      manual_scan_suffix: 'manually scan',
      read_first_comic: 'Read Your First Comic',
      reading_remember: 'Once you start reading, Mango will remember where you left off and show your entries here.',
      view_library: 'View Library',
      welcome_mango: 'Welcome to Mango',
      self_hosted_reader: 'A self-hosted manga server and reader',
      continue_reading: 'Continue Reading',
      pages: 'pages',
      read: 'read',
      continue_read_btn: 'Continue Reading',
      view_details: 'View Details',
      start_reading: 'Start Reading',
      view_all: 'View All',
      scroll_left: 'Scroll Left',
      scroll_right: 'Scroll Right',
      recently_added: 'Recently Added',

      // ================================================================
      // LOGIN PAGE
      // ================================================================
      login: 'Login',
      welcome_back: 'Welcome Back',
      login_to_mango: 'Log in to Mango',
      username: 'Username',
      enter_username: 'Enter username',
      password: 'Password',
      enter_password: 'Enter password',
      toggle_password: 'Toggle password visibility',
      login_btn: 'Login',
      mango_comic_server: 'Mango · Comic Server',

      // ================================================================
      // LIBRARY / TAGS PAGES
      // ================================================================
      library_heading: 'Library',
      files_count: 'files',
      search_manga: 'Search comics…',
      sort_auto: 'Auto',
      sort_title: 'Title',
      sort_time_modified: 'Modified Date',
      sort_time_added: 'Added Date',
      sort_progress: 'Progress',
      show_hidden: 'Show Hidden',
      hide_hidden: 'Hide Hidden',
      no_results: 'No Results',
      try_different_keywords: 'Try using different keywords',
      tags_heading: 'Tags',
      tags_count: 'tags found',
      tag_heading_prefix: 'Tag:',
      items_count: 'items',

      // ================================================================
      // TITLE PAGE
      // ================================================================
      items_selected: 'items selected',
      mark_read: 'Mark as Read',
      mark_unread: 'Mark as Unread',
      select_all: 'Select All',
      deselect_all: 'Deselect All',
      hidden_badge: 'Hidden',
      show_book: 'Show this book',
      hide_book: 'Hide this book',
      edit: 'Edit',
      found_label: 'found',
      display_name: 'Display Name',
      sort_title_label: 'Sort Title',
      cover_image: 'Cover Image',
      upload_cover_drag: 'Upload cover image, drag it here or',
      choose_one: 'choose one',
      progress: 'Progress',
      mark_all_read: 'Mark all as read (100%)',
      mark_all_unread: 'Mark all as unread (0%)',
      save: 'Save',
      new_password: 'New Password',
      enter_new_password: 'Enter new password',
      change_password: 'Change Password',
      admin_access: 'Admin Access',
      enter_username_placeholder: 'Enter username',
      enter_password_placeholder: 'Enter password',
      bulk_read_tooltip: 'title: Mark as Read',
      bulk_unread_tooltip: 'title: Mark as Unread',
      select_all_tooltip: 'title: Select All',
      deselect_all_tooltip: 'title: Deselect All',

      // ================================================================
      // READER PAGE
      // ================================================================
      reader_title: 'Reader',
      next_item: 'Next',
      exit_reader: 'Exit Reader',
      progress_label: 'Progress',
      jump_to_page: 'Jump to Page',
      mode: 'Mode',
      continuous: 'Continuous',
      paged: 'Paged',
      fit_page: 'Fit Page',
      fit_height: 'Fit Height',
      fit_width: 'Fit Width',
      original: 'Original',
      margin_px: 'Margin',
      enable_flip_animation: 'Enable Flip Animation',
      preload_images: 'Preload Images',
      right_to_left: 'Right-to-Left',
      jump_to_entry: 'Jump to Entry',
      previous: 'Previous',
      next: 'Next',
      back_to_title: 'Back to Title',
      error_title: 'Error',

      // ================================================================
      // ADMIN PAGE
      // ================================================================
      admin_heading: 'Admin',
      admin_subtitle: 'System Settings & Operations',
      user_management: 'User Management',
      user_management_desc: 'Add, edit, or delete users',
      missing_entries: 'Missing Entries',
      missing_entries_desc: 'View missing file entries',
      scan_library: 'Scan Library',
      scan_result: 'Scanned {count} entries in {ms}ms',
      scanning: 'Scanning…',
      generate_thumbnails: 'Generate Thumbnails',
      generating: 'Generating…',
      generate_thumbnails_desc: 'Generate cover thumbnails for comics',
      theme: 'Theme',
      theme_desc: 'Choose interface theme',
      ui_style: 'UI Style',
      ui_style_desc: 'Choose interface visual style',
      version_label: 'Version: v{version} · Localization:',
      logout_btn: 'Logout',

      // ================================================================
      // USER MANAGEMENT
      // ================================================================
      username_th: 'Username',
      admin_access_th: 'Admin Access',
      action_th: 'Action',
      yes: 'Yes',
      no: 'No',
      edit_user: 'Edit User',
      delete_user: 'Delete User',
      new_user: 'New User',

      // ================================================================
      // DOWNLOAD MANAGER
      // ================================================================
      download_manager_heading: 'Download Manager',
      download_manager_subtitle: 'Manage and monitor download queue',
      delete_completed: 'Delete Completed Tasks',
      retry_failed: 'Retry Failed Tasks',
      refresh_queue: 'Refresh Queue',
      resume_download: 'Resume Download',
      pause_download: 'Pause Download',
      chapter_th: 'Chapter',
      manga_th: 'Manga',
      progress_th: 'Progress',
      time_th: 'Time',
      status_th: 'Status',
      plugin_th: 'Plugin',
      operation_th: 'Operation',
      delete_action: 'Delete',
      retry_action: 'Retry',

      // ================================================================
      // SUBSCRIPTION MANAGER
      // ================================================================
      subscription_manager_heading: 'Subscription Manager',
      subscription_subtitle: 'Manage plugin subscriptions',
      no_plugins_found: 'No Plugins Found',
      no_plugins_desc: 'We could not find any plugins in',
      download_official_plugins: 'You can download official plugins from the',
      mango_plugins_repo: 'Mango Plugin Repository',
      choose_plugin: 'Choose a Plugin',
      no_subscriptions: 'No subscriptions found.',
      name_th: 'Name',
      plugin_id_th: 'Plugin ID',
      manga_title_th: 'Manga Title',
      created_at_th: 'Created At',
      last_checked_th: 'Last Checked',
      check_updates: 'Check Updates',
      subscription_details: 'Subscription Details',
      subscription_id: 'Subscription ID',
      manga_id: 'Manga ID',
      filter: 'Filter',
      keyword_th: 'Keyword',
      type_th: 'Type',
      value_th: 'Value',
      confirm: 'Confirm',

      // ================================================================
      // CARD COMPONENT
      // ================================================================
      hidden_badge_card: 'Hidden',
      new_entries_count: 'new entries',
      last_read_page: 'Last read page',
      error_card: 'Error',

      // ================================================================
      // ENTRY MODAL
      // ================================================================
      read_section: 'Read',
      beginning: 'Beginning',
      progress_section: 'Progress',
      mark_read_100: 'Mark as Read (100%)',
      mark_unread_0: 'Mark as Unread (0%)',
      edit_entry: 'Edit Entry',
      download_entry: 'Download Entry',

      // ================================================================
      // PLUGIN DOWNLOAD PAGE
      // ================================================================
      download_with_plugin: 'Download with Plugin',
      no_matching_manga: 'No matching manga found.',
      manga_found: 'manga found',
      download_selected: 'Download Selected',
      apply: 'Apply',
      clear: 'Clear',
      clear_selection: 'Clear Selection',
      target_api_version: 'Target API Version',
      chapter_search: 'Search chapters...',
      chapters_found: 'chapters found',
      too_many_chapters: 'This manga has {count} chapters, but Mango can only list {limit}. Please use filters to narrow the search.',
      no_chapters_found: 'No chapters found.',
      select_chapters_hint: 'Click table rows to select chapters. Drag the mouse across multiple rows to select them all. Hold Ctrl for multiple non-adjacent selections.',
      all: 'All',
      between: 'Between {min} and {max}',
      subscribe: 'Subscribe',
      subscription_confirm: 'Subscription Confirmation',
      subscription_create_desc: 'A subscription with the following filters will be created. All future chapters that match the filters will be downloaded automatically.',
      enter_subscription_name: 'Enter a meaningful name for the subscription to continue:',
      cancel: 'Cancel',

      // ================================================================
      // MISSING ITEMS PAGE
      // ================================================================
      no_missing_entries: 'No Missing Entries Found',
      all_entries_normal: 'All entries are normal',
      missing_items_description: 'The following items exist in your library, but we cannot find them now. If you deleted them by mistake, try restoring the files or folders back to their original location, then rescan the library. Otherwise, you can safely delete them and their associated metadata using the button below to free up database space.',
      delete_all: 'Delete All',
      missing_type_th: 'Type',
      relative_path_th: 'Relative Path',
      id_label: 'ID',
      title_label: 'Title',
      path_label: 'Path',

      // MISC
      // ================================================================
      page_title_suffix: 'Mango - {page}',
      mango_desc: 'Mango - Manga Server and Web Reader',
    },

    // ================================================================
    // SIMPLIFIED CHINESE  (zh-cn) — default, mirrors current UI text
    // ================================================================
    'zh-cn': {
      home: '主页',
      library: '资料库',
      tags: '标签',
      admin: '管理员',
      download: '下载',
      plugins: '插件',
      download_manager: '下载管理器',
      subscription_manager: '订阅管理器',
      theme_toggle: '切换主题',
      ui_style_toggle: '切换UI风格',
      view_on_github: '在 GitHub 上查看',
      language: '语言',
      logout: '登出',
      collapse_sidebar: '收起/展开侧栏',

      add_first_comic: '添加你的第一部漫画',
      no_files_found: '我们还找不到任何文件。添加一些到您的库中，它们将出现在此处。',
      current_library_path: '当前资源库路径',
      change_library_path: '想要更改您的资源库路径？',
      config_path: '配置 \u003ccode\u003econfig.yml\u003c/code\u003e 其路径为:',
      still_cant_see: '还看不到您的文件？',
      wait_scan: '您必须等待 {minutes} 分钟才能完成库扫描',
      or_manual_scan: '，或者从',
      manual_scan_suffix: '手动扫描',
      read_first_comic: '阅读你的第一部漫画',
      reading_remember: '一旦你开始阅读，Mango 会记住你离开的地方并在此处显示您的条目。',
      view_library: '查看库',
      welcome_mango: '欢迎来到 Mango',
      self_hosted_reader: '一个自托管的漫画服务器和阅读器',
      continue_reading: '继续阅读',
      pages: '页',
      read: '已读',
      continue_read_btn: '继续阅读',
      view_details: '查看详情',
      start_reading: '开始阅读',
      view_all: '查看全部',
      scroll_left: '向左滚动',
      scroll_right: '向右滚动',
      recently_added: '最近添加',

      login: '登录',
      welcome_back: '欢迎回来',
      login_to_mango: '登录到 Mango',
      username: '用户名',
      enter_username: '请输入用户名',
      password: '密码',
      enter_password: '请输入密码',
      toggle_password: '切换密码可见性',
      login_btn: '登录',
      mango_comic_server: 'Mango · 漫画服务器',

      library_heading: '资料库',
      files_count: '文件',
      search_manga: '查找漫画...',
      sort_auto: '自动',
      sort_title: '名称',
      sort_time_modified: '修改日期',
      sort_time_added: '添加日期',
      sort_progress: '进度',
      show_hidden: '显示隐藏',
      hide_hidden: '隐藏已隐藏',
      no_results: '未找到结果',
      try_different_keywords: '尝试使用不同的关键词搜索',
      tags_heading: '标签',
      tags_count: '标签找到',
      tag_heading_prefix: '标签:',
      items_count: '标记',

      items_selected: '条目选中',
      mark_read: '标记为已读',
      mark_unread: '标记为未读',
      select_all: '全选',
      deselect_all: '取消全选',
      hidden_badge: '已隐藏',
      show_book: '显示此书',
      hide_book: '隐藏此书',
      edit: '编辑',
      found_label: '找到',
      display_name: '显示名称',
      sort_title_label: '排序标题',
      cover_image: '封面图片',
      upload_cover_drag: '上传封面图片，将其拖放到此处或',
      choose_one: '选择一个',
      progress: '进度',
      mark_all_read: '全部标记为已读 (100%)',
      mark_all_unread: '全部标记为未读 (0%)',
      save: '保存',
      new_password: '新密码',
      enter_new_password: '请输入新密码',
      change_password: '更改密码',
      admin_access: '管理员权限',
      enter_username_placeholder: '请输入用户名',
      enter_password_placeholder: '请输入密码',
      bulk_read_tooltip: 'title: 标记为已读',
      bulk_unread_tooltip: 'title: 标记为未读',
      select_all_tooltip: 'title: 全选',
      deselect_all_tooltip: 'title: 取消全选',

      reader_title: '阅读',
      next_item: '下一项',
      exit_reader: '退出阅读',
      progress_label: '进度',
      jump_to_page: '跳转到页面',
      mode: '模式',
      continuous: '连续',
      paged: '分页',
      fit_page: '适合页面',
      fit_height: '适合高度',
      fit_width: '适合宽度',
      original: '原图',
      margin_px: '页边距',
      enable_flip_animation: '启用翻转动画',
      preload_images: '预加载图像',
      right_to_left: '从右到左',
      jump_to_entry: '跳转到条目',
      previous: '上一个',
      next: '下一个',
      back_to_title: '回到标题',
      error_title: '错误',

      admin_heading: '管理员',
      admin_subtitle: '系统设置与操作',
      user_management: '用户管理',
      user_management_desc: '添加、编辑或删除用户',
      missing_entries: '缺少条目',
      missing_entries_desc: '查看缺失的文件条目',
      scan_library: '扫描库文件',
      scan_result: '扫描 {count} 条目于 {ms}ms',
      scanning: '正在扫描...',
      generate_thumbnails: '生成缩略图',
      generating: '生成中...',
      generate_thumbnails_desc: '为漫画生成封面缩略图',
      theme: '主题',
      theme_desc: '选择界面主题',
      ui_style: 'UI风格',
      ui_style_desc: '选择界面视觉风格',
      version_label: '版本: v{version} · 汉化：',
      logout_btn: '登出',

      username_th: '用户名',
      admin_access_th: '管理员权限',
      action_th: '操作',
      yes: '是',
      no: '否',
      edit_user: '编辑',
      delete_user: '删除',
      new_user: '新用户',

      download_manager_heading: '下载管理器',
      download_manager_subtitle: '管理和监控下载队列',
      delete_completed: '删除已完成的任务',
      retry_failed: '重试失败的任务',
      refresh_queue: '刷新队列',
      resume_download: '恢复下载',
      pause_download: '暂停下载',
      chapter_th: '章节',
      manga_th: '漫画',
      progress_th: '进度',
      time_th: '时间',
      status_th: '状态',
      plugin_th: '插件',
      operation_th: '操作',
      delete_action: '删除',
      retry_action: '重试',

      subscription_manager_heading: '订阅管理器',
      subscription_subtitle: '管理插件订阅',
      no_plugins_found: '未找到插件',
      no_plugins_desc: '我们在下列目录中找不到任何插件',
      download_official_plugins: '您可以从以下网址下载官方插件',
      mango_plugins_repo: 'Mango 插件库',
      choose_plugin: '选择一个插件',
      no_subscriptions: '未找到订阅.',
      name_th: '名称',
      plugin_id_th: '插件 ID',
      manga_title_th: '漫画标题',
      created_at_th: '创建时间',
      last_checked_th: '上次检查',
      check_updates: '检查更新',
      subscription_details: '订阅详情',
      subscription_id: '订阅 ID',
      manga_id: '漫画 ID',
      filter: '过滤器',
      keyword_th: '关键词',
      type_th: '种类',
      value_th: '值',
      confirm: '确认',

      hidden_badge_card: '已隐藏',
      new_entries_count: '新条目',
      last_read_page: '上次阅读页',
      error_card: '错误',

      read_section: '阅读',
      beginning: '从头开始',
      progress_section: '进度',
      mark_read_100: '标记为已读 (100%)',
      mark_unread_0: '标记为未读 (0%)',
      edit_entry: '编辑条目',
      download_entry: '下载条目',

      // ================================================================
      // MISSING ITEMS PAGE
      // ================================================================
      no_missing_entries: '没有找到丢失的条目',
      all_entries_normal: '所有条目均正常',
      missing_items_description: '以下项目存在于您的资料库中，但现在我们找不到它们了。 如果您错误地删除了它们，请尝试恢复文件或文件夹，将它们放回原来的位置，然后重新扫描资料库。 除此之外，您可以使用下面的按钮安全地删除它们和相关的元数据以释放数据库空间。',
      delete_all: '删除全部',
      missing_type_th: '类型',
      relative_path_th: '相对路径',
      id_label: 'ID',
      title_label: '标题',
      path_label: '路径',

      // PLUGIN DOWNLOAD PAGE
      // ================================================================
      download_with_plugin: '使用插件下载',
      no_matching_manga: '没有找到匹配的漫画.',
      manga_found: '漫画找到',
      download_selected: '下载已选',
      apply: '应用',
      clear: '清除',
      clear_selection: '清空',
      target_api_version: '目标 API 版本',
      chapter_search: '搜索章节...',
      chapters_found: '章节已找到',
      too_many_chapters: '该漫画有 {count} 章节, 但 Mango 最多只能列出 {limit}. 请使用过滤器缩小搜索范围.',
      no_chapters_found: '没有找到章节.',
      select_chapters_hint: '单击表格行以选择章节。 将鼠标拖到多行上以将它们全部选中。 按住 Ctrl 进行多个不相邻的选择.',
      all: '全部',
      between: '在 {min} 和 {max} 之间',
      subscribe: '订阅',
      subscription_confirm: '订阅确认',
      subscription_create_desc: '将创建具有以下过滤器的订阅。 全部 <strong>未来</strong> 匹配过滤器的章节将被自动下载.',
      enter_subscription_name: '为订阅输入一个有意义的名称以继续:',
      cancel: '取消',

      page_title_suffix: 'Mango - {page}',
      mango_desc: 'Mango - 漫画服务器和网络阅读器',
    },

    // ================================================================
    // TRADITIONAL CHINESE  (zh-tw)
    // ================================================================
    'zh-tw': {
      home: '主頁',
      library: '資料庫',
      tags: '標籤',
      admin: '管理員',
      download: '下載',
      plugins: '外掛程式',
      download_manager: '下載管理器',
      subscription_manager: '訂閱管理器',
      theme_toggle: '切換主題',
      ui_style_toggle: '切換UI風格',
      view_on_github: '在 GitHub 上檢視',
      language: '語言',
      logout: '登出',
      collapse_sidebar: '收起/展開側欄',

      add_first_comic: '新增你的第一部漫畫',
      no_files_found: '我們還找不到任何檔案。新增一些到您的庫中，它們將出現在此處。',
      current_library_path: '目前資源庫路徑',
      change_library_path: '想要變更您的資源庫路徑？',
      config_path: '設定 \u003ccode\u003econfig.yml\u003c/code\u003e 其路徑為：',
      still_cant_see: '還看不到您的檔案？',
      wait_scan: '您必須等待 {minutes} 分鐘才能完成庫掃描',
      or_manual_scan: '，或者從',
      manual_scan_suffix: '手動掃描',
      read_first_comic: '閱讀你的第一部漫畫',
      reading_remember: '一旦你開始閱讀，Mango 會記住你離開的地方並在此處顯示您的條目。',
      view_library: '檢視庫',
      welcome_mango: '歡迎使用 Mango',
      self_hosted_reader: '一個自託管的漫畫伺服器和閱讀器',
      continue_reading: '繼續閱讀',
      pages: '頁',
      read: '已讀',
      continue_read_btn: '繼續閱讀',
      view_details: '檢視詳情',
      start_reading: '開始閱讀',
      view_all: '檢視全部',
      scroll_left: '向左捲動',
      scroll_right: '向右捲動',
      recently_added: '最近新增',

      login: '登入',
      welcome_back: '歡迎回來',
      login_to_mango: '登入 Mango',
      username: '使用者名稱',
      enter_username: '請輸入使用者名稱',
      password: '密碼',
      enter_password: '請輸入密碼',
      toggle_password: '切換密碼可見性',
      login_btn: '登入',
      mango_comic_server: 'Mango · 漫畫伺服器',

      library_heading: '資料庫',
      files_count: '檔案',
      search_manga: '搜尋漫畫...',
      sort_auto: '自動',
      sort_title: '名稱',
      sort_time_modified: '修改日期',
      sort_time_added: '新增日期',
      sort_progress: '進度',
      show_hidden: '顯示隱藏',
      hide_hidden: '隱藏已隱藏',
      no_results: '未找到結果',
      try_different_keywords: '嘗試使用不同的關鍵字搜尋',
      tags_heading: '標籤',
      tags_count: '標籤找到',
      tag_heading_prefix: '標籤：',
      items_count: '標記',

      items_selected: '條目選中',
      mark_read: '標記為已讀',
      mark_unread: '標記為未讀',
      select_all: '全選',
      deselect_all: '取消全選',
      hidden_badge: '已隱藏',
      show_book: '顯示此書',
      hide_book: '隱藏此書',
      edit: '編輯',
      found_label: '找到',
      display_name: '顯示名稱',
      sort_title_label: '排序標題',
      cover_image: '封面圖片',
      upload_cover_drag: '上傳封面圖片，將其拖放到此處或',
      choose_one: '選擇一個',
      progress: '進度',
      mark_all_read: '全部標記為已讀 (100%)',
      mark_all_unread: '全部標記為未讀 (0%)',
      save: '儲存',
      new_password: '新密碼',
      enter_new_password: '請輸入新密碼',
      change_password: '變更密碼',
      admin_access: '管理員權限',
      enter_username_placeholder: '請輸入使用者名稱',
      enter_password_placeholder: '請輸入密碼',
      bulk_read_tooltip: 'title: 標記為已讀',
      bulk_unread_tooltip: 'title: 標記為未讀',
      select_all_tooltip: 'title: 全選',
      deselect_all_tooltip: 'title: 取消全選',

      reader_title: '閱讀',
      next_item: '下一項',
      exit_reader: '退出閱讀',
      progress_label: '進度',
      jump_to_page: '跳轉到頁面',
      mode: '模式',
      continuous: '連續',
      paged: '分頁',
      fit_page: '適合頁面',
      fit_height: '適合高度',
      fit_width: '適合寬度',
      original: '原圖',
      margin_px: '頁邊距',
      enable_flip_animation: '啟用翻頁動畫',
      preload_images: '預載入圖像',
      right_to_left: '從右到左',
      jump_to_entry: '跳轉到條目',
      previous: '上一個',
      next: '下一個',
      back_to_title: '回到標題',
      error_title: '錯誤',

      admin_heading: '管理員',
      admin_subtitle: '系統設定與操作',
      user_management: '使用者管理',
      user_management_desc: '新增、編輯或刪除使用者',
      missing_entries: '缺少條目',
      missing_entries_desc: '檢視缺失的檔案條目',
      scan_library: '掃描庫檔案',
      scan_result: '掃描 {count} 條目於 {ms}ms',
      scanning: '正在掃描...',
      generate_thumbnails: '產生縮圖',
      generating: '產生中...',
      generate_thumbnails_desc: '為漫畫產生封面縮圖',
      theme: '主題',
      theme_desc: '選擇介面主題',
      ui_style: 'UI風格',
      ui_style_desc: '選擇介面視覺風格',
      version_label: '版本：v{version} · 本地化：',
      logout_btn: '登出',

      username_th: '使用者名稱',
      admin_access_th: '管理員權限',
      action_th: '操作',
      yes: '是',
      no: '否',
      edit_user: '編輯',
      delete_user: '刪除',
      new_user: '新增使用者',

      download_manager_heading: '下載管理器',
      download_manager_subtitle: '管理和監控下載佇列',
      delete_completed: '刪除已完成的任務',
      retry_failed: '重試失敗的任務',
      refresh_queue: '重新整理佇列',
      resume_download: '恢復下載',
      pause_download: '暫停下載',
      chapter_th: '章節',
      manga_th: '漫畫',
      progress_th: '進度',
      time_th: '時間',
      status_th: '狀態',
      plugin_th: '外掛程式',
      operation_th: '操作',
      delete_action: '刪除',
      retry_action: '重試',

      subscription_manager_heading: '訂閱管理器',
      subscription_subtitle: '管理外掛程式訂閱',
      no_plugins_found: '未找到外掛程式',
      no_plugins_desc: '我們在下列目錄中找不到任何外掛程式',
      download_official_plugins: '您可以從以下網址下載官方外掛程式',
      mango_plugins_repo: 'Mango 外掛程式庫',
      choose_plugin: '選擇一個外掛程式',
      no_subscriptions: '未找到訂閱。',
      name_th: '名稱',
      plugin_id_th: '外掛程式 ID',
      manga_title_th: '漫畫標題',
      created_at_th: '建立時間',
      last_checked_th: '上次檢查',
      check_updates: '檢查更新',
      subscription_details: '訂閱詳情',
      subscription_id: '訂閱 ID',
      manga_id: '漫畫 ID',
      filter: '篩選器',
      keyword_th: '關鍵字',
      type_th: '種類',
      value_th: '值',
      confirm: '確認',

      hidden_badge_card: '已隱藏',
      new_entries_count: '新條目',
      last_read_page: '上次閱讀頁',
      error_card: '錯誤',

      read_section: '閱讀',
      beginning: '從頭開始',
      progress_section: '進度',
      mark_read_100: '標記為已讀 (100%)',
      mark_unread_0: '標記為未讀 (0%)',
      edit_entry: '編輯條目',
      download_entry: '下載條目',

      // ================================================================
      // MISSING ITEMS PAGE
      // ================================================================
      no_missing_entries: '沒有找到丟失的條目',
      all_entries_normal: '所有條目均正常',
      missing_items_description: '以下項目存在於您的資料庫中，但現在我們找不到它們了。 如果您錯誤地刪除了它們，請嘗試恢復檔案或資料夾，將它們放回原來的位置，然後重新掃描資料庫。 除此之外，您可以使用下面的按鈕安全地刪除它們和相關的中繼資料以釋放資料庫空間。',
      delete_all: '刪除全部',
      missing_type_th: '類型',
      relative_path_th: '相對路徑',
      id_label: 'ID',
      title_label: '標題',
      path_label: '路徑',

      // PLUGIN DOWNLOAD PAGE
      // ================================================================
      download_with_plugin: '使用外掛程式下載',
      no_matching_manga: '沒有找到匹配的漫畫.',
      manga_found: '漫畫找到',
      download_selected: '下載已選',
      apply: '套用',
      clear: '清除',
      clear_selection: '清空',
      target_api_version: '目標 API 版本',
      chapter_search: '搜尋章節...',
      chapters_found: '章節已找到',
      too_many_chapters: '該漫畫有 {count} 章節, 但 Mango 最多只能列出 {limit}. 請使用篩選器縮小搜尋範圍.',
      no_chapters_found: '沒有找到章節.',
      select_chapters_hint: '點擊表格行以選擇章節。 拖曳滑鼠跨越多行以將它們全部選中。 按住 Ctrl 進行多個不相鄰的選擇.',
      all: '全部',
      between: '在 {min} 和 {max} 之間',
      subscribe: '訂閱',
      subscription_confirm: '訂閱確認',
      subscription_create_desc: '將建立具有以下篩選器的訂閱。 全部 <strong>未來</strong> 匹配篩選器的章節將被自動下載.',
      enter_subscription_name: '為訂閱輸入一個有意義的名稱以繼續:',
      cancel: '取消',

      page_title_suffix: 'Mango - {page}',
      mango_desc: 'Mango - 漫畫伺服器和網路閱讀器',
    }
  };

  // --- Helpers ------------------------------------------------------

  function getText(key, lang) {
    const d = DICT[lang] || DICT[DEFAULT_LANG];
    return d[key] !== undefined ? d[key] : key;
  }

  function interpolate(str, vars) {
    if (!vars) return str;
    return str.replace(/\{([^}]+)\}/g, function (_match, name) {
      return vars[name] !== undefined ? vars[name] : '{' + name + '}';
    });
  }

  // --- Public API ---------------------------------------------------

  return {
    DEFAULT_LANG: DEFAULT_LANG,
    SUPPORTED_LANGS: ['zh-cn', 'zh-tw', 'en'],
    LANG_LABELS: {
      'en': 'EN',
      'zh-cn': '简',
      'zh-tw': '繁'
    },

    get: function (key, lang, vars) {
      return interpolate(getText(key, lang), vars);
    },

    translatePage: function (lang) {
      if (!lang) lang = DEFAULT_LANG;
      if (lang === DEFAULT_LANG) {
        // Even for the default language, we need to re-apply in case
        // the page was previously translated to another language.
        // However, the HTML already contains the default Chinese text,
        // so we can skip for initial load and only run when switching.
        // We always run to be safe (restores original text from data-i18n).
      }

      // --- Text content ---
      document.querySelectorAll('[data-i18n]').forEach(function (el) {
        const key = el.dataset.i18n;
        // Collect interpolation vars from data-i18n-var-* attributes
        const vars = {};
        for (let attr in el.dataset) {
          const match = attr.match(/^i18nVar[s]?(.+)$/);
          if (match) {
            const varName = match[1].charAt(0).toLowerCase() + match[1].slice(1);
            vars[varName] = el.dataset[attr];
          }
        }
        const text = interpolate(getText(key, lang), vars);
        if (text !== undefined) {
          // For elements that contain only text or a single text node,
          // replace textContent. If the element has child elements
          // (like icons), preserve them and replace only the text nodes.
          if (el.children.length === 0) {
            el.textContent = text;
          } else {
            // Replace only direct text nodes, leave child elements alone
            for (let i = 0; i < el.childNodes.length; i++) {
              const node = el.childNodes[i];
              if (node.nodeType === Node.TEXT_NODE && node.textContent.trim()) {
                node.textContent = text;
                break; // only replace the first meaningful text node
              }
            }
          }
        }
      });

      // --- title attribute ---
      document.querySelectorAll('[data-i18n-title]').forEach(function (el) {
        const key = el.dataset.i18nTitle;
        const text = getText(key, lang);
        if (text !== undefined) el.title = text;
      });

      // --- aria-label attribute ---
      document.querySelectorAll('[data-i18n-aria-label]').forEach(function (el) {
        const key = el.dataset.i18nAriaLabel;
        const text = getText(key, lang);
        if (text !== undefined) el.setAttribute('aria-label', text);
      });

      // --- placeholder attribute ---
      document.querySelectorAll('[data-i18n-placeholder]').forEach(function (el) {
        const key = el.dataset.i18nPlaceholder;
        const text = getText(key, lang);
        if (text !== undefined) el.placeholder = text;
      });

      // --- value attribute ---
      document.querySelectorAll('[data-i18n-value]').forEach(function (el) {
        const key = el.dataset.i18nValue;
        const text = getText(key, lang);
        if (text !== undefined) el.value = text;
      });

      // --- content attribute (for meta tags) ---
      document.querySelectorAll('[data-i18n-content]').forEach(function (el) {
        const key = el.dataset.i18nContent;
        const text = getText(key, lang);
        if (text !== undefined) el.content = text;
      });

      // --- uk-tooltip attribute  (format: "title: text") ---
      document.querySelectorAll('[data-i18n-tooltip]').forEach(function (el) {
        const key = el.dataset.i18nTooltip;
        const text = getText(key, lang);
        if (text !== undefined) {
          const prefix = text.indexOf('title:') === 0 ? '' : 'title: ';
          el.setAttribute('uk-tooltip', prefix + text);
        }
      });

      // --- document title (pages set data-page-title on <body> or a meta el) ---
      const pageTitleEl = document.querySelector('[data-i18n-page-title]');
      if (pageTitleEl) {
        const key = pageTitleEl.dataset.i18nPageTitle;
        const pageName = getText(key, lang);
        document.title = interpolate(getText('page_title_suffix', lang), { page: pageName });
      }

      // --- Update compact language-toggle label ---
      document.querySelectorAll('.lang-toggle-label').forEach(function (el) {
        el.textContent = I18N.LANG_LABELS[lang] || lang;
      });
    }
  };
})();
