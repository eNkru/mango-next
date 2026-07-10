const component = () => {
	return {
		subscriptions: [],
		plugins: [],
		pid: undefined,
		subscription: undefined, // selected subscription
		loading: false,

		init() {
			fetch(`${base_url}api/admin/plugin`)
				.then((res) => res.json())
				.then((data) => {
					if (!data.success) throw new Error(data.error);
					this.plugins = data.plugins;

					const pid = localStorage.getItem("plugin");
					if (pid && this.plugins.map((p) => p.id).includes(pid))
						this.pid = pid;
					else if (this.plugins.length > 0)
						this.pid = this.plugins[0].id;

					this.list(pid);
				})
				.catch((e) => {
					alert(
						"danger",
						`无法列出可用插件. Error: ${e}`
					);
				});
		},
		pluginChanged() {
			localStorage.setItem("plugin", this.pid);
			this.list(this.pid);
		},
		list(pid) {
			if (!pid) return;
			fetch(
				`${base_url}api/admin/plugin/subscriptions?${new URLSearchParams(
					{
						plugin: pid,
					}
				)}`,
				{
					method: "GET",
				}
			)
				.then((response) => response.json())
				.then((data) => {
					if (!data.success) throw new Error(data.error);
					this.subscriptions = data.subscriptions;
				})
				.catch((e) => {
					alert(
						"danger",
						`未能列出订阅. Error: ${e}`
					);
				});
		},
		formatCellValue(value) {
			if (value === undefined || value === null) return "";
			return value.toString();
		},
		renderStrCellText(str) {
			const text = this.formatCellValue(str);
			const maxLength = 40;
			if (text.length > maxLength)
				return `${text.substring(0, maxLength)}...`;
			return text;
		},
		renderStrCellTitle(str) {
			const text = this.formatCellValue(str);
			return text.length > 40 ? text : "";
		},
		renderDateCell(timestamp) {
			return moment
				.duration(moment.unix(timestamp).diff(moment()))
				.humanize(true);
		},
		selected(event, modal) {
			const id = event.currentTarget.getAttribute("sid");
			this.subscription = this.subscriptions.find((s) => s.id === id);
			UIkit.modal(modal).show();
		},
		renderFilterType(ft) {
			let type = ft.type;
			switch (type) {
				case "number-min":
					return "number (minimum value)";
				case "number-max":
					return "number (maximum value)";
				case "date-min":
					return "minimum date";
				case "date-max":
					return "maximum date";
				default:
					return type;
			}
		},
		renderFilterValue(ft) {
			let value = ft.value;

			if (ft.type.startsWith("number") && isNaN(value)) value = "";
			else if (ft.type.startsWith("date") && value)
				value = moment(Number(value)).format("MMM D, YYYY");

			return this.formatCellValue(value);
		},
		actionHandler(event, type) {
			const id = $(event.currentTarget).closest("tr").attr("sid");
			if (type !== 'delete') return this.action(id, type);
			UIkit.modal.confirm('您确定要删除订阅吗？ 这不能被撤消。', {
				labels: {
					ok: '是的，删除',
					cancel: '取消'
				}
			}).then(() => {
				this.action(id, type);
			});
		},
		action(id, type) {
			if (this.loading) return;
			this.loading = true;
			fetch(
				`${base_url}api/admin/plugin/subscriptions${type === 'update' ? '/update' : ''}?${new URLSearchParams(
					{
						plugin: this.pid,
						subscription: id,
					}
				)}`,
				{
					method: type === 'delete' ? "DELETE" : 'POST'
				}
			)
				.then((response) => response.json())
				.then((data) => {
					if (!data.success) throw new Error(data.error);
					if (type === 'update')
						alert("success", `检查订阅更新 ${id}. 检查日志以了解进度或稍后返回此页面.`);
				})
				.catch((e) => {
					alert(
						"danger",
						`未能 ${type} 成功订阅. Error: ${e}`
					);
				})
				.finally(() => {
					this.loading = false;
					this.list(this.pid);
				});
		},
	};
};
