const alert = (level, text) => {
	$('#alert, #alert-mobile').empty();
	const html = `<div class="uk-alert-${level}" uk-alert><a class="uk-alert-close" uk-close></a><p>${text}</p></div>`;
	$('#alert, #alert-mobile').append(html);
	$("html, body").animate({ scrollTop: 0 });
};
