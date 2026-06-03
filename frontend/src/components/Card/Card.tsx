import type { HTMLAttributes, ReactNode } from "react";
import styles from "./Card.module.css";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
	children: ReactNode;
	variant?: "default" | "elevated";
	padding?: "none" | "sm" | "md" | "lg";
}

export function Card({
	children,
	variant = "default",
	padding = "md",
	className = "",
	...props
}: CardProps) {
	return (
		<div
			className={`${styles.card} ${styles[variant]} ${styles[`pad-${padding}`]} ${className}`}
			{...props}
		>
			{children}
		</div>
	);
}
