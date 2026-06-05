export interface LayoutOptions {
	title?: string;
	isAuthenticated?: boolean;
}

export function layout(content: string, options?: LayoutOptions): string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>${escapeHtml(options?.title ?? "Barricade")}</title>
	<link rel="stylesheet" href="/styles.css">
</head>
<body>
	<div class="page">
		<header class="header">
			<div class="inner">
				<a href="/" class="logo">Barricade</a>
				${
					options?.isAuthenticated
						? `
				<div class="spacer"></div>
				<form method="POST" action="/logout">
					<button type="submit" class="logout-button">Sign out</button>
				</form>`
						: ""
				}
			</div>
		</header>
		<main id="main-content" class="main">
			${content}
		</main>
		<footer class="footer">
			<div class="inner">
				<p class="copyright">&copy; ${new Date().getFullYear()} Barricade</p>
			</div>
		</footer>
	</div>
</body>
</html>`;
}

export function escapeHtml(str: string): string {
	return str
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#39;");
}
