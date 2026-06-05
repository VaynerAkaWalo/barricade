interface DashboardProps {
	ownerId?: string;
}

export function dashboardPage(_props: DashboardProps): string {
	return `
<div class="dashboard-page">
	<div class="dashboard-content">
		<h1 class="dashboard-title">Welcome!</h1>
		<p class="dashboard-subtitle">You are signed in to Barricade.</p>
	</div>
</div>`;
}
