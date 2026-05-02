$(function(){
	$('.uk-search-input').keyup(function(){
		var input = $('.uk-search-input').val();
		$('.topbar-search-input').val(input);
		if (typeof filterPageItems === 'function') filterPageItems(input);
	});
});
