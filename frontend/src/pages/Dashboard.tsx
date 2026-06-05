import styles from "./Dashboard.module.css";

export function DashboardPage() {
	return (
		<div className={styles.page}>
			<div className={styles.content}>
				<h1 className={styles.title}>Welcome!</h1>
				<p className={styles.subtitle}>You are signed in to Barricade.</p>
			</div>
		</div>
	);
}
