import { Outlet } from "@tanstack/react-router";
import styles from "../routeTree.module.css";
import { Footer } from "./Footer/Footer";
import { Header } from "./Header/Header";

export function RootLayout() {
	return (
		<div className={styles.page}>
			<a href="#main-content" className={styles.skipLink}>
				Skip to main content
			</a>
			<Header />
			<main id="main-content" className={styles.main} tabIndex={-1}>
				<Outlet />
			</main>
			<Footer />
		</div>
	);
}
