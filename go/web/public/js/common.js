/**
 * --- Alpine helper functions
 */

/**
 * Set an alpine.js property
 *
 * @function setProp
 * @param {string} key - Key of the data property
 * @param {*} prop - The data property
 * @param {string} selector - The jQuery selector to the root element
 */
const setProp = (key, prop, selector = '#root') => {
	$(selector).get(0).__x.$data[key] = prop;
};

/**
 * Get an alpine.js property
 *
 * @function getProp
 * @param {string} key - Key of the data property
 * @param {string} selector - The jQuery selector to the root element
 * @return {*} The data property
 */
const getProp = (key, selector = '#root') => {
	return $(selector).get(0).__x.$data[key];
};

/**
 * --- Language / i18n related functions
 */

/**
 * Load the language setting from localStorage, or use the default
 *
 * @function loadLanguageSetting
 * @return {string} A language code ('zh-cn', 'zh-tw', or 'en')
 */
const loadLanguageSetting = () => {
	let lang = localStorage.getItem('mango-language');
	if (!lang || I18N.SUPPORTED_LANGS.indexOf(lang) < 0) {
		lang = I18N.DEFAULT_LANG;
	}
	return lang;
};

/**
 * Save a language setting
 *
 * @function saveLanguageSetting
 * @param {string} lang - A language code
 */
const saveLanguageSetting = lang => {
	if (I18N.SUPPORTED_LANGS.indexOf(lang) < 0) lang = I18N.DEFAULT_LANG;
	localStorage.setItem('mango-language', lang);
};

/**
 * Set the page language and apply translations
 *
 * @function setLanguage
 * @param {string} lang - A language code
 */
const setLanguage = (lang) => {
	saveLanguageSetting(lang);
	I18N.translatePage(lang);
};

/**
 * Cycle to the next language in the supported list
 *
 * @function cycleLanguage
 */
const cycleLanguage = () => {
	const current = loadLanguageSetting();
	const langs = I18N.SUPPORTED_LANGS;
	const idx = langs.indexOf(current);
	const next = langs[(idx + 1) % langs.length];
	setLanguage(next);
};

/**
 * --- UI Style functions (comic vs flat)
 */

/**
 * Check whether a given string represents a valid UI style
 *
 * @function validUIStyle
 * @param {string} style - The string representing the UI style
 * @return {bool}
 */
const validUIStyle = (style) => {
	return ['comic', 'flat'].indexOf(style) >= 0;
};

/**
 * Load UI style from local storage, or use 'comic' as default
 *
 * @function loadUIStyle
 * @return {string} The UI style ('comic' or 'flat')
 */
const loadUIStyle = () => {
	let style = localStorage.getItem('ui-style');
	if (!style || !validUIStyle(style)) style = 'comic';
	return style;
};

/**
 * Save a UI style setting
 *
 * @function saveUIStyle
 * @param {string} style - A UI style ('comic' or 'flat')
 */
const saveUIStyle = style => {
	if (!validUIStyle(style)) style = 'comic';
	localStorage.setItem('ui-style', style);
};

/**
 * Apply classList changes to documentElement and body when present.
 * Avoids jQuery no-ops when script runs before <body> exists (FOUC).
 */
const applyThemeClass = (action, className) => {
	const roots = [document.documentElement];
	if (document.body) roots.push(document.body);
	roots.forEach(el => {
		if (action === 'add') el.classList.add(className);
		else el.classList.remove(className);
	});
	// Keep jQuery in sync for any late-bound body selectors
	if (action === 'add') {
		$('html').addClass(className);
		$('body').addClass(className);
	} else {
		$('html').removeClass(className);
		$('body').removeClass(className);
	}
};

/**
 * Apply a UI style to the body/html elements.
 * Also syncs the dark-mode visual classes so that switching between
 * comic/flat doesn't require a full theme re-apply.
 *
 * @function setUIStyle
 * @param {string} style - The UI style to apply
 */
const setUIStyle = (style) => {
	if (!style) style = loadUIStyle();
	const theme = loadTheme();
	if (style === 'comic') {
		applyThemeClass('add', 'comic-theme');
		// If dark theme is active, use CSS classes for background
		if (theme === 'dark') {
			document.documentElement.style.background = '';
			$('html').css('background', '');
			applyThemeClass('add', 'comic-theme-dark');
		}
	} else {
		applyThemeClass('remove', 'comic-theme');
		applyThemeClass('remove', 'comic-theme-dark');
		// If dark theme is active, use inline style for background
		if (theme === 'dark') {
			document.documentElement.style.background = '#121212';
			$('html').css('background', '#121212');
		}
	}
};

/**
 * Toggle the UI style between comic and flat
 *
 * @function toggleUIStyle
 */
const toggleUIStyle = () => {
	const current = loadUIStyle();
	const next = current === 'comic' ? 'flat' : 'comic';
	saveUIStyle(next);
	setUIStyle(next);
};

/**
 * --- Theme related functions
 *  	Note: In the comments below we treat "theme" and "theme setting"
 *  		differently. A theme can have only two values, either "dark" or
 *  		"light", while a theme setting can have the third value "system".
 */

/**
 * Check if the system setting prefers dark theme.
 * 		from https://flaviocopes.com/javascript-detect-dark-mode/
 *
 * @function preferDarkMode
 * @return {bool}
 */
const preferDarkMode = () => {
	return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
};

/**
 * Check whether a given string represents a valid theme setting
 *
 * @function validThemeSetting
 * @param {string} theme - The string representing the theme setting
 * @return {bool}
 */
const validThemeSetting = (theme) => {
	return ['dark', 'light', 'system'].indexOf(theme) >= 0;
};

/**
 * Load theme setting from local storage, or use 'light'
 *
 * @function loadThemeSetting
 * @return {string} A theme setting ('dark', 'light', or 'system')
 */
const loadThemeSetting = () => {
	let str = localStorage.getItem('theme');
	if (!str || !validThemeSetting(str)) str = 'system';
	return str;
};

/**
 * Load the current theme (not theme setting)
 *
 * @function loadTheme
 * @return {string} The current theme to use ('dark' or 'light')
 */
const loadTheme = () => {
	let setting = loadThemeSetting();
	if (setting === 'system') {
		setting = preferDarkMode() ? 'dark' : 'light';
	}
	return setting;
};

/**
 * Save a theme setting
 *
 * @function saveThemeSetting
 * @param {string} setting - A theme setting
 */
const saveThemeSetting = setting => {
	if (!validThemeSetting(setting)) setting = 'system';
	localStorage.setItem('theme', setting);
};

/**
 * Toggle the current theme. When the current theme setting is 'system', it
 *		will be changed to either 'light' or 'dark'
 *
 * @function toggleTheme
 */
const toggleTheme = () => {
	const theme = loadTheme();
	const newTheme = theme === 'dark' ? 'light' : 'dark';
	saveThemeSetting(newTheme);
	setTheme(newTheme);
};

/**
 * Apply a theme, or load a theme and then apply it.
 * Reads the current UI style (comic/flat) to decide whether to use CSS
 * classes or inline styles for the dark background, but does NOT modify
 * the UI style setting.
 *
 * @function setTheme
 * @param {string?} theme - (Optional) The theme to apply. When omitted, use
 * 		`loadTheme` to get a theme and apply it.
 */
const setTheme = (theme) => {
	if (!theme) theme = loadTheme();
	const uiStyle = loadUIStyle();
	if (theme === 'dark') {
		// In comic mode, CSS handles the html/body backgrounds
		// (comic-theme-dark classes). In flat mode, html bg via inline;
		// body bg via body.uk-light:not(.comic-theme*) in comic-theme.less.
		if (uiStyle !== 'comic') {
			document.documentElement.style.background = '#121212';
			$('html').css('background', '#121212');
		} else {
			document.documentElement.style.background = '';
			$('html').css('background', '');
		}
		if (document.body) document.body.classList.add('uk-light');
		$('body').addClass('uk-light');
		$('.ui-widget-content').addClass('dark');
		if (uiStyle === 'comic') {
			applyThemeClass('add', 'comic-theme-dark');
		}
	} else {
		document.documentElement.style.background = '';
		$('html').css('background', '');
		if (document.body) document.body.classList.remove('uk-light');
		$('body').removeClass('uk-light');
		$('.ui-widget-content').removeClass('dark');
		applyThemeClass('remove', 'comic-theme-dark');
	}
};

// Expose functions used in inline onclick handlers on window
// (const at top-level of non-module scripts doesn't always resolve
//  in inline event handlers across all browsers/caching scenarios)
window.toggleTheme = toggleTheme;
window.toggleUIStyle = toggleUIStyle;
window.cycleLanguage = cycleLanguage;
window.setLanguage = setLanguage;

/**
 * Floating utility speed-dial (language / theme / style / GitHub / logout)
 */
const setUtilityFabOpen = (open) => {
	const root = document.getElementById('utility-fab');
	if (!root) return;
	const primary = document.getElementById('utility-fab-primary');
	const menu = document.getElementById('utility-fab-menu');
	if (!primary || !menu) return;

	if (open) {
		root.classList.add('is-open');
		menu.hidden = false;
		primary.setAttribute('aria-expanded', 'true');
	} else {
		root.classList.remove('is-open');
		menu.hidden = true;
		primary.setAttribute('aria-expanded', 'false');
	}
};

const closeUtilityFab = () => setUtilityFabOpen(false);

const toggleUtilityFab = () => {
	const root = document.getElementById('utility-fab');
	if (!root) return;
	setUtilityFabOpen(!root.classList.contains('is-open'));
};

const initUtilityFab = () => {
	const root = document.getElementById('utility-fab');
	const primary = document.getElementById('utility-fab-primary');
	const menu = document.getElementById('utility-fab-menu');
	if (!root || !primary || !menu) return;

	primary.addEventListener('click', (event) => {
		event.stopPropagation();
		toggleUtilityFab();
	});

	menu.addEventListener('click', (event) => {
		const actionEl = event.target.closest('[data-utility-action]');
		if (!actionEl || !menu.contains(actionEl)) return;

		const action = actionEl.getAttribute('data-utility-action');
		if (action === 'language') {
			event.preventDefault();
			cycleLanguage();
			closeUtilityFab();
		} else if (action === 'theme') {
			event.preventDefault();
			toggleTheme();
			closeUtilityFab();
		} else if (action === 'ui-style') {
			event.preventDefault();
			toggleUIStyle();
			closeUtilityFab();
		} else {
			// GitHub / logout: let navigation proceed, still close for consistency
			closeUtilityFab();
		}
	});

	document.addEventListener('click', (event) => {
		if (!root.classList.contains('is-open')) return;
		if (root.contains(event.target)) return;
		closeUtilityFab();
	});

	document.addEventListener('keydown', (event) => {
		if (event.key !== 'Escape') return;
		if (!root.classList.contains('is-open')) return;
		closeUtilityFab();
		primary.focus();
	});
};

window.toggleUtilityFab = toggleUtilityFab;
window.closeUtilityFab = closeUtilityFab;

// Apply UI style and theme before document is ready to prevent the
// initial flash of white on most pages. applyThemeClass targets
// documentElement immediately; body classes apply when body exists
// (layout/login also call setUIStyle+setTheme after body is present).
setUIStyle();
setTheme();
// Apply i18n translation on page load (non-default only to avoid flash)
$(() => {
	const lang = loadLanguageSetting();
	if (lang !== I18N.DEFAULT_LANG) {
		I18N.translatePage(lang);
	}
	// Apply UI style and theme now that <body> exists in the DOM.
	// setUIStyle handles comic/flat classes; setTheme handles uk-light,
	// .ui-widget-content.dark, and comic-theme-dark on the body element.
	setUIStyle();
	setTheme();
	initUtilityFab();

	// on system dark mode setting change
	if (window.matchMedia) {
		window.matchMedia('(prefers-color-scheme: dark)')
			.addEventListener('change', event => {
				if (loadThemeSetting() === 'system')
					setTheme(event.matches ? 'dark' : 'light');
			});
	}
});
