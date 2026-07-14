const alert = (level, text, options = {}) => {
	const targets = $('#alert, #alert-mobile');
	const content = $("<p>");
	const container = $("<div>")
		.addClass(`uk-alert-${level}`)
		.attr("uk-alert", "");
	const close = $("<a>")
		.addClass("uk-alert-close")
		.html('<i class="fas fa-times"></i>');

	if (options.allowHtml === true)
		content.html(text);
	else
		content.text(text);

	container.append(close, content);
	targets.empty().append(container);
	$("html, body").animate({ scrollTop: 0 });
};
