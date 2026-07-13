$(() => {
	var target = base_url + 'admin/user/edit';
	// Must be /admin/user/edit/{username} — without slash becomes /admin/user/editadmin (404).
	if (username) target += '/' + username;
	$('form').attr('action', target);
	if (error) alert('danger', error);
});
