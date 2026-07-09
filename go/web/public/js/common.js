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
		$('html').addClass('comic-theme');
		$('body').addClass('comic-theme');
		// If dark theme is active, use CSS classes for background
		if (theme === 'dark') {
			$('html').css('background', '');
			$('html').addClass('comic-theme-dark');
			$('body').addClass('comic-theme-dark');
		}
	} else {
		$('html').removeClass('comic-theme').removeClass('comic-theme-dark');
		$('body').removeClass('comic-theme').removeClass('comic-theme-dark');
		// If dark theme is active, use inline style for background
		if (theme === 'dark') {
			$('html').css('background', 'rgb(20, 20, 20)');
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
		// (comic-theme-dark classes). In flat mode, use inline style.
		if (uiStyle !== 'comic') {
			$('html').css('background', 'rgb(20, 20, 20)');
		} else {
			$('html').css('background', '');
		}
		$('body').addClass('uk-light');
		$('.ui-widget-content').addClass('dark');
		if (uiStyle === 'comic') {
			$('html').addClass('comic-theme-dark');
			$('body').addClass('comic-theme-dark');
		}
	} else {
		$('html').css('background', '');
		$('body').removeClass('uk-light');
		$('.ui-widget-content').removeClass('dark');
		$('html').removeClass('comic-theme-dark');
		$('body').removeClass('comic-theme-dark');
	}
};

// Expose functions used in inline onclick handlers on window
// (const at top-level of non-module scripts doesn't always resolve
//  in inline event handlers across all browsers/caching scenarios)
window.toggleTheme = toggleTheme;
window.toggleUIStyle = toggleUIStyle;
window.cycleLanguage = cycleLanguage;
window.setLanguage = setLanguage;

// Apply UI style and theme before document is ready to prevent the
// initial flash of white on most pages
setUIStyle();
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

	// on system dark mode setting change
	if (window.matchMedia) {
		window.matchMedia('(prefers-color-scheme: dark)')
			.addEventListener('change', event => {
				if (loadThemeSetting() === 'system')
					setTheme(event.matches ? 'dark' : 'light');
			});
	}
});
