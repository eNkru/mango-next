$(function(){
	$('.uk-search-input').keyup(function(){
		var input = $('.uk-search-input').val();
		if (typeof filterPageItems === 'function') filterPageItems(input);
	});
});
