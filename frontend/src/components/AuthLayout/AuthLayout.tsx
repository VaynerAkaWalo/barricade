import type { ReactNode } from "react";
import styles from "./AuthLayout.module.css";

interface AuthLayoutProps {
	title: string;
	subtitle: string;
	children: ReactNode;
}

export function AuthLayout({ title, subtitle, children }: AuthLayoutProps) {
	return (
		<div className={styles.page}>
			<div className={styles.card}>
				<h1 className={styles.title}>{title}</h1>
				<p className={styles.subtitle}>{subtitle}</p>
				{children}
			</div>
		</div>
	);
}
